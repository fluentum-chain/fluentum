package kvstore

import (
	"context"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/fluentum-chain/fluentum/abci/example/code"
	"github.com/fluentum-chain/fluentum/libs/log"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
)

var _ ApplicationInterface = (*PersistentKVStoreApplication)(nil)

type PersistentKVStoreApplication struct {
	app *Application

	// validator set
	ValUpdates []abci.ValidatorUpdate

	logger log.Logger
}

func NewPersistentKVStoreApplication(dbDir string) *PersistentKVStoreApplication {
	name := "kvstore"
	_, err := dbm.NewDB(name, "pebble", dbDir)
	if err != nil {
		panic(err)
	}

	// Create a new Application instance
	app := NewApplication()

	return &PersistentKVStoreApplication{
		app:    app,
		logger: log.NewNopLogger(),
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
		// For now, just log the validator
		app.logger.Info("Validator in InitChain", "validator", v)
	}
	return &abci.ResponseInitChain{}, nil
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

func (app *PersistentKVStoreApplication) Echo(ctx context.Context, req *abci.RequestEcho) (*abci.ResponseEcho, error) {
	return &abci.ResponseEcho{Message: req.Message}, nil
}

func (app *PersistentKVStoreApplication) Flush(ctx context.Context, req *abci.RequestFlush) (*abci.ResponseFlush, error) {
	return &abci.ResponseFlush{}, nil
}
