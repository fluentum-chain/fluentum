package kvstore

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/abci/example/code"
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

var _ abci.Application = (*Application)(nil)

type Application struct {
	abci.BaseApplication

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

// FinalizeBlock handles the ABCI 2.0 FinalizeBlock call
func (app *Application) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	txResults := make([]*abci.ExecTxResult, len(req.Txs))

	for i, tx := range req.Txs {
		// Process each transaction
		var key, value []byte
		parts := bytes.Split(tx, []byte("="))
		if len(parts) == 2 {
			key, value = parts[0], parts[1]
		} else {
			key, value = tx, tx
		}

		err := app.state.db.Set(prefixKey(key), value)
		if err != nil {
			panic(err)
		}
		app.state.Size++

		events := []abci.Event{
			{
				Type: "app",
				Attributes: []abci.EventAttribute{
					{Key: "creator", Value: "Cosmoshi Netowoko", Index: true},
					{Key: "key", Value: string(key), Index: true},
					{Key: "index_key", Value: "index is working", Index: true},
					{Key: "noindex_key", Value: "index is working", Index: false},
				},
			},
		}

		txResults[i] = &abci.ExecTxResult{
			Code:   code.CodeTypeOK,
			Events: events,
		}
	}

	return &abci.ResponseFinalizeBlock{
		TxResults: txResults,
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

func (app *Application) BeginBlock(ctx context.Context, req *abci.RequestBeginBlock) (*abci.ResponseBeginBlock, error) {
	return &abci.ResponseBeginBlock{}, nil
}

func (app *Application) EndBlock(ctx context.Context, req *abci.RequestEndBlock) (*abci.ResponseEndBlock, error) {
	return &abci.ResponseEndBlock{}, nil
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
