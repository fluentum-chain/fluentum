package kvstore

import (
	"context"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/fluentum-chain/fluentum/abci/example/code"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/libs/log"
)

var _ abci.Application = (*PersistentKVStoreApplication)(nil)

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

func (app *PersistentKVStoreApplication) Info(ctx context.Context, req *abci.InfoRequest) (*abci.InfoResponse, error) {
	return &abci.InfoResponse{
		Data:             "kvstore",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}, nil
}

func (app *PersistentKVStoreApplication) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	return &abci.CheckTxResponse{Code: code.CodeTypeOK}, nil
}

// Commit will panic if InitChain was not called
func (app *PersistentKVStoreApplication) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	return &abci.CommitResponse{}, nil
}

// When path=/val and data={validator address}, returns the validator update (types.ValidatorUpdate) varint encoded.
// For any other path, returns an associated value or nil if missing.
func (app *PersistentKVStoreApplication) Query(ctx context.Context, reqQuery *abci.QueryRequest) (*abci.QueryResponse, error) {
	switch reqQuery.Path {
	case "/val":
		// For now, return empty response
		resQuery := &abci.QueryResponse{
			Key:   reqQuery.Data,
			Value: []byte{},
		}
		return resQuery, nil
	default:
		return &abci.QueryResponse{}, nil
	}
}

// Save the validators in the merkle tree
func (app *PersistentKVStoreApplication) InitChain(ctx context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	for _, v := range req.Validators {
		// For now, just log the validator
		app.logger.Info("Validator in InitChain", "validator", v)
	}
	return &abci.InitChainResponse{}, nil
}

func (app *PersistentKVStoreApplication) ListSnapshots(ctx context.Context, req *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	return &abci.ListSnapshotsResponse{}, nil
}

func (app *PersistentKVStoreApplication) LoadSnapshotChunk(ctx context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	return &abci.LoadSnapshotChunkResponse{}, nil
}

func (app *PersistentKVStoreApplication) OfferSnapshot(ctx context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	return &abci.OfferSnapshotResponse{Result: abci.ResponseOfferSnapshot_ABORT}, nil
}

func (app *PersistentKVStoreApplication) ApplySnapshotChunk(ctx context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	return &abci.ApplySnapshotChunkResponse{Result: abci.ResponseApplySnapshotChunk_ABORT}, nil
}

func (app *PersistentKVStoreApplication) Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}

// Additional required methods for the Application interface
func (app *PersistentKVStoreApplication) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	return &abci.PrepareProposalResponse{}, nil
}

func (app *PersistentKVStoreApplication) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	return &abci.ProcessProposalResponse{}, nil
}

func (app *PersistentKVStoreApplication) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	return &abci.FinalizeBlockResponse{}, nil
}

func (app *PersistentKVStoreApplication) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	return &abci.ExtendVoteResponse{}, nil
}

func (app *PersistentKVStoreApplication) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	return &abci.VerifyVoteExtensionResponse{}, nil
}
