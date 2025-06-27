package kvstore

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/fluentum-chain/fluentum/abci/example/code"
	abci "github.com/fluentum-chain/fluentum/abci/types"
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
	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication() *Application {
	state := loadState(dbm.NewMemDB())
	return &Application{state: state}
}

func (app *Application) Info(ctx context.Context, req *abci.InfoRequest) (*abci.InfoResponse, error) {
	return &abci.InfoResponse{
		Data:             fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:          version.ABCIVersion,
		AppVersion:       ProtocolVersion,
		LastBlockHeight:  app.state.Height,
		LastBlockAppHash: app.state.AppHash,
	}, nil
}

func (app *Application) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	return &abci.CheckTxResponse{Code: code.CodeTypeOK, GasWanted: 1}, nil
}

func (app *Application) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)

	resp := &abci.CommitResponse{}
	if app.RetainBlocks > 0 && app.state.Height >= app.RetainBlocks {
		resp.RetainHeight = app.state.Height - app.RetainBlocks + 1
	}
	return resp, nil
}

// Returns an associated value or nil if missing.
func (app *Application) Query(ctx context.Context, reqQuery *abci.QueryRequest) (*abci.QueryResponse, error) {
	if reqQuery.Prove {
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}
		if value == nil {
			return &abci.QueryResponse{Log: "does not exist"}, nil
		} else {
			return &abci.QueryResponse{Log: "exists", Index: -1, Key: reqQuery.Data, Value: value, Height: app.state.Height}, nil
		}
	}

	value, err := app.state.db.Get(prefixKey(reqQuery.Data))
	if err != nil {
		panic(err)
	}
	if value == nil {
		return &abci.QueryResponse{Log: "does not exist", Key: reqQuery.Data, Height: app.state.Height}, nil
	} else {
		return &abci.QueryResponse{Log: "exists", Value: value, Height: app.state.Height}, nil
	}
}

func (app *Application) InitChain(ctx context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	return &abci.InitChainResponse{}, nil
}

func (app *Application) ListSnapshots(ctx context.Context, req *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	return &abci.ListSnapshotsResponse{}, nil
}

func (app *Application) LoadSnapshotChunk(ctx context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	return &abci.LoadSnapshotChunkResponse{}, nil
}

func (app *Application) OfferSnapshot(ctx context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	return &abci.OfferSnapshotResponse{Result: abci.ResponseOfferSnapshot_REJECT}, nil
}

func (app *Application) ApplySnapshotChunk(ctx context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	return &abci.ApplySnapshotChunkResponse{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
}

func (app *Application) Echo(ctx context.Context, req *abci.EchoRequest) (*abci.EchoResponse, error) {
	return &abci.EchoResponse{Message: req.Message}, nil
}

// Additional methods required by the Application interface
func (app *Application) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	return &abci.PrepareProposalResponse{}, nil
}

func (app *Application) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	return &abci.ProcessProposalResponse{}, nil
}

func (app *Application) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	return &abci.FinalizeBlockResponse{}, nil
}

func (app *Application) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	return &abci.ExtendVoteResponse{}, nil
}

func (app *Application) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	return &abci.VerifyVoteExtensionResponse{}, nil
}
