package kvstore

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	pc "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"github.com/fluentum-chain/fluentum/abci/example/code"
	"github.com/fluentum-chain/fluentum/libs/log"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

// Custom interface that doesn't require ExtendVote
type ApplicationInterface interface {
	Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error)
	CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error)
	Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error)
	InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error)
	BeginBlock(ctx context.Context, req *abci.RequestBeginBlock) (*abci.ResponseBeginBlock, error)
	EndBlock(ctx context.Context, req *abci.RequestEndBlock) (*abci.ResponseEndBlock, error)
	ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
}

var _ ApplicationInterface = (*PersistentKVStoreApplication)(nil)

type PersistentKVStoreApplication struct {
	app *Application

	// validator set
	ValUpdates []abci.ValidatorUpdate

	valAddrToPubKeyMap map[string]pc.PublicKey

	logger log.Logger
}

func NewPersistentKVStoreApplication(dbDir string) *PersistentKVStoreApplication {
	name := "kvstore"
	db, err := dbm.NewDB(name, "pebble", dbDir)
	if err != nil {
		panic(err)
	}

	// Create a new Application instance
	app := NewApplication()

	return &PersistentKVStoreApplication{
		app:                app,
		valAddrToPubKeyMap: make(map[string]pc.PublicKey),
		logger:             log.NewNopLogger(),
	}
}

func (app *PersistentKVStoreApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *PersistentKVStoreApplication) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{
		Data:             "kvstore",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}, nil
}

func (app *PersistentKVStoreApplication) SetOption(ctx context.Context, req *abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	return &abci.ResponseSetOption{}, nil
}

// tx is either "val:pubkey!power" or "key=value" or just arbitrary bytes
func (app *PersistentKVStoreApplication) DeliverTx(ctx context.Context, req *abci.RequestDeliverTx) (*abci.ResponseDeliverTx, error) {
	// if it starts with "val:", update the validator set
	// format is "val:pubkey!power"
	if isValidatorTx(req.Tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		resp := app.execValidatorTx(req.Tx)
		return &resp, nil
	}

	// otherwise, update the key-value store
	return &abci.ResponseDeliverTx{Code: code.CodeTypeOK}, nil
}

func (app *PersistentKVStoreApplication) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return &abci.ResponseCheckTx{Code: code.CodeTypeOK}, nil
}

// Commit will panic if InitChain was not called
func (app *PersistentKVStoreApplication) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return &abci.ResponseCommit{}, nil
}

// When path=/val and data={validator address}, returns the validator update (types.ValidatorUpdate) varint encoded.
// For any other path, returns an associated value or nil if missing.
func (app *PersistentKVStoreApplication) Query(ctx context.Context, reqQuery *abci.RequestQuery) (*abci.ResponseQuery, error) {
	switch reqQuery.Path {
	case "/val":
		key := []byte("val:" + string(reqQuery.Data))
		// For now, return empty response
		resQuery := &abci.ResponseQuery{
			Key:   reqQuery.Data,
			Value: []byte{},
		}
		return resQuery, nil
	default:
		return &abci.ResponseQuery{}, nil
	}
}

// Save the validators in the merkle tree
func (app *PersistentKVStoreApplication) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.Code != 0 {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return &abci.ResponseInitChain{}, nil
}

// Track the block hash and header information
func (app *PersistentKVStoreApplication) BeginBlock(ctx context.Context, req *abci.RequestBeginBlock) (*abci.ResponseBeginBlock, error) {
	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)

	// Punish validators who committed equivocation.
	for _, ev := range req.ByzantineValidators {
		if ev.Type == abci.EvidenceType_DUPLICATE_VOTE {
			addr := string(ev.Validator.Address)
			if pubKey, ok := app.valAddrToPubKeyMap[addr]; ok {
				app.updateValidator(abci.ValidatorUpdate{
					PubKey: pubKey,
					Power:  ev.Validator.Power - 1,
				})
				app.logger.Info("Decreased val power by 1 because of the equivocation",
					"val", addr)
			} else {
				app.logger.Error("Wanted to punish val, but can't find it",
					"val", addr)
			}
		}
	}

	return &abci.ResponseBeginBlock{}, nil
}

// Update the validator set
func (app *PersistentKVStoreApplication) EndBlock(ctx context.Context, req *abci.RequestEndBlock) (*abci.ResponseEndBlock, error) {
	return &abci.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}, nil
}

func (app *PersistentKVStoreApplication) ListSnapshots(
	ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return &abci.ResponseListSnapshots{}, nil
}

func (app *PersistentKVStoreApplication) LoadSnapshotChunk(
	ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return &abci.ResponseLoadSnapshotChunk{}, nil
}

func (app *PersistentKVStoreApplication) OfferSnapshot(
	ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
}

func (app *PersistentKVStoreApplication) ApplySnapshotChunk(
	ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
}

//---------------------------------------------
// update validators

func (app *PersistentKVStoreApplication) Validators() (validators []abci.ValidatorUpdate) {
	// For now, return empty validators list
	return []abci.ValidatorUpdate{}
}

func MakeValSetChangeTx(pubkey pc.PublicKey, power int64) []byte {
	pk, err := cryptoenc.PubKeyFromProto(pubkey)
	if err != nil {
		panic(err)
	}
	pubStr := base64.StdEncoding.EncodeToString(pk.Bytes())
	return []byte(fmt.Sprintf("val:%s!%d", pubStr, power))
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

// format is "val:pubkey!power"
// pubkey is a base64-encoded 32-byte ed25519 key
func (app *PersistentKVStoreApplication) execValidatorTx(tx []byte) abci.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

	//  get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "!")
	if len(pubKeyAndPower) != 2 {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey!power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid base64", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return abci.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(abci.ValidatorUpdate{
		PubKey: pubkey,
		Power:  power,
	})
}

// add, update, or remove a validator
func (app *PersistentKVStoreApplication) updateValidator(v abci.ValidatorUpdate) abci.ResponseDeliverTx {
	// For now, just add to ValUpdates without database persistence
	app.ValUpdates = append(app.ValUpdates, v)
	return abci.ResponseDeliverTx{Code: code.CodeTypeOK}
}
