package kvstore

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/fluentum-chain/fluentum/abci/example/code"
	abci "github.com/fluentum-chain/fluentum/proto/tendermint/abci"
	"github.com/fluentum-chain/fluentum/version"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")

	ProtocolVersion uint64 = 0x1
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	var state State
	state.db = db
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	err = state.db.Set(stateKey, stateBytes)
	if err != nil {
		panic(err)
	}
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

//---------------------------------------------------

// Custom interface that matches the available local ABCI types
type ApplicationInterface interface {
	Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error)
	CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error)
	Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error)
	Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error)
	InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error)
	ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
	SetOption(ctx context.Context, req *abci.RequestSetOption) (*abci.ResponseSetOption, error)
	Echo(ctx context.Context, req *abci.RequestEcho) (*abci.ResponseEcho, error)
	Flush(ctx context.Context, req *abci.RequestFlush) (*abci.ResponseFlush, error)
}

var _ ApplicationInterface = (*Application)(nil)

type Application struct {
	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication() *Application {
	state := loadState(dbm.NewMemDB())
	return &Application{state: state}
}

func (app *Application) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return &abci.ResponseInfo{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCIVersion,
		AppVersion:       ProtocolVersion,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}, nil
}

func (app *Application) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return &abci.ResponseCheckTx{Code: code.CodeTypeOK, GasWanted: 1}, nil
}

func (app *Application) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)

	resp := &abci.ResponseCommit{}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp, nil
}

// Returns an associated value or nil if missing.
func (app *Application) Query(ctx context.Context, reqQuery *abci.RequestQuery) (*abci.ResponseQuery, error) {
	if reqQuery.Prove {
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}
		if value == nil {
			return &abci.ResponseQuery{Log: "does not exist"}, nil
		} else {
			return &abci.ResponseQuery{Log: "exists", Index: -1, Key: reqQuery.Data, Value: value, Height: app.state.Height}, nil
		}
	}

	value, err := app.state.db.Get(prefixKey(reqQuery.Data))
	if err != nil {
		panic(err)
	}
	if value == nil {
		return &abci.ResponseQuery{Log: "does not exist", Key: reqQuery.Data, Height: app.state.Height}, nil
	} else {
		return &abci.ResponseQuery{Log: "exists", Value: value, Height: app.state.Height}, nil
	}
}

func (app *Application) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return &abci.ResponseInitChain{}, nil
}

func (app *Application) ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return &abci.ResponseListSnapshots{}, nil
}

func (app *Application) LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return &abci.ResponseLoadSnapshotChunk{}, nil
}

func (app *Application) OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return &abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ABORT}, nil
}

func (app *Application) ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return &abci.ResponseApplySnapshotChunk{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
}

func (app *Application) SetOption(ctx context.Context, req *abci.RequestSetOption) (*abci.ResponseSetOption, error) {
	return &abci.ResponseSetOption{}, nil
}

func (app *Application) Echo(ctx context.Context, req *abci.RequestEcho) (*abci.ResponseEcho, error) {
	return &abci.ResponseEcho{Message: req.Message}, nil
}

func (app *Application) Flush(ctx context.Context, req *abci.RequestFlush) (*abci.ResponseFlush, error) {
	return &abci.ResponseFlush{}, nil
}
