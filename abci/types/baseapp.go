package types

import (
	"context"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
)

// BaseApplication provides default implementations for all ABCI methods
// Applications can embed this struct and override only the methods they need
type BaseApplication struct{}

// NewBaseApplication creates a new BaseApplication instance
func NewBaseApplication() *BaseApplication {
	return &BaseApplication{}
}

// Info/Query Connection
func (BaseApplication) Info(ctx context.Context, req *InfoRequest) (*InfoResponse, error) {
	return &InfoResponse{
		Data:             "base_application",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}, nil
}

func (BaseApplication) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	return &QueryResponse{
		Code:   CodeTypeOK,
		Log:    "query successful",
		Height: req.Height,
	}, nil
}

// Mempool Connection
func (BaseApplication) CheckTx(ctx context.Context, req *CheckTxRequest) (*CheckTxResponse, error) {
	return &CheckTxResponse{
		Code:      CodeTypeOK,
		Data:      []byte{},
		Log:       "check tx successful",
		GasWanted: 0,
		GasUsed:   0,
		Events:    []*Event{},
		Codespace: "",
	}, nil
}

// Consensus Connection
func (BaseApplication) PrepareProposal(ctx context.Context, req *PrepareProposalRequest) (*PrepareProposalResponse, error) {
	// Default implementation: return transactions as-is
	return &PrepareProposalResponse{
		Txs: req.Txs,
	}, nil
}

func (BaseApplication) ProcessProposal(ctx context.Context, req *ProcessProposalRequest) (*ProcessProposalResponse, error) {
	// Default implementation: accept all proposals
	return &ProcessProposalResponse{
		Status: ResponseProcessProposal_ACCEPT,
	}, nil
}

func (BaseApplication) FinalizeBlock(ctx context.Context, req *FinalizeBlockRequest) (*FinalizeBlockResponse, error) {
	// Default implementation: create empty results for all transactions
	txResults := make([]*cometbftabci.ExecTxResult, len(req.Txs))
	for i := range req.Txs {
		txResults[i] = &cometbftabci.ExecTxResult{
			Code:      CodeTypeOK,
			Data:      []byte{},
			Log:       "transaction processed",
			Info:      "",
			Events:    []cometbftabci.Event{},
			GasUsed:   0,
			GasWanted: 0,
		}
	}

	return &FinalizeBlockResponse{
		TxResults:             txResults,
		ValidatorUpdates:      []ValidatorUpdate{},
		ConsensusParamUpdates: nil,
		AppHash:               []byte{},
		Events:                []cometbftabci.Event{},
	}, nil
}

func (BaseApplication) ExtendVote(ctx context.Context, req *ExtendVoteRequest) (*ExtendVoteResponse, error) {
	// Default implementation: return empty vote extension
	return &ExtendVoteResponse{
		VoteExtension: []byte{},
	}, nil
}

func (BaseApplication) VerifyVoteExtension(ctx context.Context, req *VerifyVoteExtensionRequest) (*VerifyVoteExtensionResponse, error) {
	// Default implementation: accept all vote extensions
	return &VerifyVoteExtensionResponse{
		Status: ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

func (BaseApplication) Commit(ctx context.Context, req *CommitRequest) (*CommitResponse, error) {
	return &CommitResponse{
		RetainHeight: 0,
	}, nil
}

func (BaseApplication) InitChain(ctx context.Context, req *InitChainRequest) (*InitChainResponse, error) {
	return &InitChainResponse{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppHash:         []byte{},
	}, nil
}

// State Sync Connection (optional)
func (BaseApplication) ListSnapshots(ctx context.Context, req *ListSnapshotsRequest) (*ListSnapshotsResponse, error) {
	return &ListSnapshotsResponse{
		Snapshots: []*Snapshot{},
	}, nil
}

func (BaseApplication) OfferSnapshot(ctx context.Context, req *OfferSnapshotRequest) (*OfferSnapshotResponse, error) {
	return &OfferSnapshotResponse{
		Result: ResponseOfferSnapshot_REJECT,
	}, nil
}

func (BaseApplication) LoadSnapshotChunk(ctx context.Context, req *LoadSnapshotChunkRequest) (*LoadSnapshotChunkResponse, error) {
	return &LoadSnapshotChunkResponse{
		Chunk: []byte{},
	}, nil
}

func (BaseApplication) ApplySnapshotChunk(ctx context.Context, req *ApplySnapshotChunkRequest) (*ApplySnapshotChunkResponse, error) {
	return &ApplySnapshotChunkResponse{
		Result:        ResponseApplySnapshotChunk_ABORT,
		RefetchChunks: []uint32{},
		RejectSenders: []string{},
	}, nil
}

func (BaseApplication) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: req.Message}, nil
}
