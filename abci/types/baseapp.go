package types

import (
	"context"
	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// BaseApplication provides default implementations for all ABCI methods
// Applications can embed this struct and override only the methods they need
type BaseApplication struct{}

// Info/Query Connection
func (BaseApplication) Info(ctx context.Context, req *RequestInfo) (*ResponseInfo, error) {
	return &cmtabci.ResponseInfo{
		Data:             "base_application",
		Version:          "1.0.0",
		AppVersion:       1,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}, nil
}

func (BaseApplication) Query(ctx context.Context, req *RequestQuery) (*ResponseQuery, error) {
	return &cmtabci.ResponseQuery{
		Code:   CodeTypeOK,
		Log:    "query successful",
		Height: req.Height,
	}, nil
}

// Mempool Connection
func (BaseApplication) CheckTx(ctx context.Context, req *RequestCheckTx) (*ResponseCheckTx, error) {
	return &cmtabci.ResponseCheckTx{
		Code:      CodeTypeOK,
		Data:      []byte{},
		Log:       "check tx successful",
		GasWanted: 0,
		GasUsed:   0,
		Events:    []cmtabci.Event{},
		Codespace: "",
	}, nil
}

// Consensus Connection
func (BaseApplication) PrepareProposal(ctx context.Context, req *RequestPrepareProposal) (*ResponsePrepareProposal, error) {
	// Default implementation: return transactions as-is
	return &cmtabci.ResponsePrepareProposal{
		Txs: req.Txs,
	}, nil
}

func (BaseApplication) ProcessProposal(ctx context.Context, req *RequestProcessProposal) (*ResponseProcessProposal, error) {
	// Default implementation: accept all proposals
	return &cmtabci.ResponseProcessProposal{
		Status: cmtabci.ResponseProcessProposal_ACCEPT,
	}, nil
}

func (BaseApplication) FinalizeBlock(ctx context.Context, req *RequestFinalizeBlock) (*ResponseFinalizeBlock, error) {
	// Default implementation: create empty results for all transactions
	txResults := make([]*cmtabci.ExecTxResult, len(req.Txs))
	for i := range req.Txs {
		txResults[i] = &cmtabci.ExecTxResult{
			Code:      CodeTypeOK,
			Data:      []byte{},
			Log:       "transaction processed",
			Info:      "",
			Events:    []cmtabci.Event{},
			GasUsed:   0,
			GasWanted: 0,
		}
	}

	return &cmtabci.ResponseFinalizeBlock{
		TxResults:             txResults,
		ValidatorUpdates:      []cmtabci.ValidatorUpdate{},
		ConsensusParamUpdates: nil,
		AppHash:               []byte{},
		Events:                []cmtabci.Event{},
	}, nil
}

func (BaseApplication) ExtendVote(ctx context.Context, req *RequestExtendVote) (*ResponseExtendVote, error) {
	// Default implementation: return empty vote extension
	return &cmtabci.ResponseExtendVote{
		VoteExtension: []byte{},
	}, nil
}

func (BaseApplication) VerifyVoteExtension(ctx context.Context, req *RequestVerifyVoteExtension) (*ResponseVerifyVoteExtension, error) {
	// Default implementation: accept all vote extensions
	return &cmtabci.ResponseVerifyVoteExtension{
		Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

func (BaseApplication) Commit(ctx context.Context, req *RequestCommit) (*ResponseCommit, error) {
	return &cmtabci.ResponseCommit{
		Data:         []byte{},
		RetainHeight: 0,
	}, nil
}

func (BaseApplication) InitChain(ctx context.Context, req *RequestInitChain) (*ResponseInitChain, error) {
	return &cmtabci.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
		AppHash:         []byte{},
	}, nil
}

// State Sync Connection (optional)
func (BaseApplication) ListSnapshots(ctx context.Context, req *RequestListSnapshots) (*ResponseListSnapshots, error) {
	return &cmtabci.ResponseListSnapshots{
		Snapshots: []*cmtabci.Snapshot{},
	}, nil
}

func (BaseApplication) OfferSnapshot(ctx context.Context, req *RequestOfferSnapshot) (*ResponseOfferSnapshot, error) {
	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_REJECT,
	}, nil
}

func (BaseApplication) LoadSnapshotChunk(ctx context.Context, req *RequestLoadSnapshotChunk) (*ResponseLoadSnapshotChunk, error) {
	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: []byte{},
	}, nil
}

func (BaseApplication) ApplySnapshotChunk(ctx context.Context, req *RequestApplySnapshotChunk) (*ResponseApplySnapshotChunk, error) {
	return &cmtabci.ResponseApplySnapshotChunk{
		Result:        cmtabci.ResponseApplySnapshotChunk_REJECT,
		RefetchChunks: []uint32{},
		RejectSenders: []string{},
	}, nil
} 