package kvstore

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	pc "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/fluentum-chain/fluentum/abci/example/code"
	"github.com/fluentum-chain/fluentum/abci/example/kvstore"
	"github.com/fluentum-chain/fluentum/abci/types"
	cryptoenc "github.com/fluentum-chain/fluentum/crypto/encoding"
	"github.com/fluentum-chain/fluentum/libs/log"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

//-----------------------------------------

var _ abci.Application = (*PersistentKVStoreApplication)(nil)

type PersistentKVStoreApplication struct {
	app *kvstore.Application

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

	state := kvstore.LoadState(db)

	return &PersistentKVStoreApplication{
		app:                &kvstore.Application{State: state},
		valAddrToPubKeyMap: make(map[string]pc.PublicKey),
		logger:             log.NewNopLogger(),
	}
}

func (app *PersistentKVStoreApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *PersistentKVStoreApplication) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	res, err := app.app.Info(ctx, req)
	if err != nil {
		return nil, err
	}
	res.LastBlockHeight = app.app.State.Height
	res.LastBlockAppHash = app.app.State.AppHash
	return res, nil
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
	return app.app.CheckTx(ctx, req)
}

// Commit will panic if InitChain was not called
func (app *PersistentKVStoreApplication) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return app.app.Commit(ctx, req)
}

// When path=/val and data={validator address}, returns the validator update (types.ValidatorUpdate) varint encoded.
// For any other path, returns an associated value or nil if missing.
func (app *PersistentKVStoreApplication) Query(ctx context.Context, reqQuery *abci.RequestQuery) (*abci.ResponseQuery, error) {
	switch reqQuery.Path {
	case "/val":
		key := []byte("val:" + string(reqQuery.Data))
		value, err := app.app.State.DB.Get(key)
		if err != nil {
			panic(err)
		}

		resQuery := &abci.ResponseQuery{
			Key:   reqQuery.Data,
			Value: value,
		}
		return resQuery, nil
	default:
		return app.app.Query(ctx, reqQuery)
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
	itr, err := app.app.State.DB.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(abci.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	if err = itr.Error(); err != nil {
		panic(err)
	}
	return
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
	pubkey, err := cryptoenc.PubKeyFromProto(v.PubKey)
	if err != nil {
		panic(fmt.Errorf("can't decode public key: %w", err))
	}
	key := []byte("val:" + string(pubkey.Bytes()))

	if v.Power == 0 {
		// remove validator
		hasKey, err := app.app.State.DB.Has(key)
		if err != nil {
			panic(err)
		}
		if !hasKey {
			pubStr := base64.StdEncoding.EncodeToString(pubkey.Bytes())
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %s", pubStr)}
		}
		if err = app.app.State.DB.Delete(key); err != nil {
			panic(err)
		}
		delete(app.valAddrToPubKeyMap, string(pubkey.Address()))
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return abci.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		if err = app.app.State.DB.Set(key, value.Bytes()); err != nil {
			panic(err)
		}
		app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return abci.ResponseDeliverTx{Code: code.CodeTypeOK}
}
