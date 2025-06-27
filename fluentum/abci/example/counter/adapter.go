package counter

import (
	"context"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	localabci "github.com/fluentum-chain/fluentum/abci/types"
)

// CometBFTAdapter adapts the local Fluentum Application to CometBFT's Application interface
type CometBFTAdapter struct {
	app *Application
}

// NewCometBFTAdapter creates a new adapter for the counter application
func NewCometBFTAdapter(app *Application) *CometBFTAdapter {
	return &CometBFTAdapter{app: app}
}

// Info implements CometBFT's Application interface
func (a *CometBFTAdapter) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	resp, err := a.app.Info(ctx, &localabci.InfoRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseInfo{
		Data: resp.Data,
	}, nil
}

// CheckTx implements CometBFT's Application interface
func (a *CometBFTAdapter) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	resp, err := a.app.CheckTx(ctx, &localabci.CheckTxRequest{Tx: req.Tx})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseCheckTx{
		Code:      resp.Code,
		Data:      resp.Data,
		Log:       resp.Log,
		Info:      resp.Info,
		GasWanted: resp.GasWanted,
		GasUsed:   resp.GasUsed,
		Events:    nil,
		Codespace: resp.Codespace,
	}, nil
}

// FinalizeBlock implements CometBFT's Application interface
func (a *CometBFTAdapter) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	resp, err := a.app.FinalizeBlock(ctx, &localabci.FinalizeBlockRequest{Txs: req.Txs})
	if err != nil {
		return nil, err
	}

	// Convert ExecTxResults from local to CometBFT types
	txResults := make([]*cmtabci.ExecTxResult, len(resp.TxResults))
	for i, txResult := range resp.TxResults {
		txResults[i] = &cmtabci.ExecTxResult{
			Code:      txResult.Code,
			Data:      txResult.Data,
			Log:       txResult.Log,
			Info:      txResult.Info,
			GasWanted: txResult.GasWanted,
			GasUsed:   txResult.GasUsed,
			Events:    txResult.Events,
			Codespace: txResult.Codespace,
		}
	}

	return &cmtabci.ResponseFinalizeBlock{
		TxResults:             txResults,
		Events:                resp.Events,
		ValidatorUpdates:      resp.ValidatorUpdates,
		ConsensusParamUpdates: resp.ConsensusParamUpdates,
	}, nil
}

// Commit implements CometBFT's Application interface
func (a *CometBFTAdapter) Commit(ctx context.Context, req *cmtabci.RequestCommit) (*cmtabci.ResponseCommit, error) {
	_, err := a.app.Commit(ctx, &localabci.CommitRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseCommit{}, nil
}

// Query implements CometBFT's Application interface
func (a *CometBFTAdapter) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	resp, err := a.app.Query(ctx, &localabci.QueryRequest{
		Path:   req.Path,
		Data:   req.Data,
		Height: req.Height,
		Prove:  req.Prove,
	})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseQuery{
		Code:      resp.Code,
		Log:       resp.Log,
		Info:      resp.Info,
		Index:     resp.Index,
		Key:       resp.Key,
		Value:     resp.Value,
		ProofOps:  resp.ProofOps,
		Height:    resp.Height,
		Codespace: resp.Codespace,
	}, nil
}

// Echo implements CometBFT's Application interface
func (a *CometBFTAdapter) Echo(ctx context.Context, req *cmtabci.RequestEcho) (*cmtabci.ResponseEcho, error) {
	resp, err := a.app.Echo(ctx, &localabci.EchoRequest{Message: req.Message})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseEcho{Message: resp.Message}, nil
}

// Flush implements CometBFT's Application interface
func (a *CometBFTAdapter) Flush(ctx context.Context, req *cmtabci.RequestFlush) (*cmtabci.ResponseFlush, error) {
	return &cmtabci.ResponseFlush{}, nil
}

// InitChain implements CometBFT's Application interface
func (a *CometBFTAdapter) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	resp, err := a.app.InitChain(ctx, &localabci.InitChainRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseInitChain{
		ConsensusParams: resp.ConsensusParams,
		Validators:      resp.Validators,
		AppHash:         resp.AppHash,
	}, nil
}

// PrepareProposal implements CometBFT's Application interface
func (a *CometBFTAdapter) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	resp, err := a.app.PrepareProposal(ctx, &localabci.PrepareProposalRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponsePrepareProposal{
		Txs: resp.Txs,
	}, nil
}

// ProcessProposal implements CometBFT's Application interface
func (a *CometBFTAdapter) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	resp, err := a.app.ProcessProposal(ctx, &localabci.ProcessProposalRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseProcessProposal{
		Status: resp.Status,
	}, nil
}

// ExtendVote implements CometBFT's Application interface
func (a *CometBFTAdapter) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	resp, err := a.app.ExtendVote(ctx, &localabci.ExtendVoteRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseExtendVote{
		VoteExtension: resp.VoteExtension,
	}, nil
}

// VerifyVoteExtension implements CometBFT's Application interface
func (a *CometBFTAdapter) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	resp, err := a.app.VerifyVoteExtension(ctx, &localabci.VerifyVoteExtensionRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseVerifyVoteExtension{
		Status: resp.Status,
	}, nil
}

// ListSnapshots implements CometBFT's Application interface
func (a *CometBFTAdapter) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	resp, err := a.app.ListSnapshots(ctx, &localabci.ListSnapshotsRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseListSnapshots{
		Snapshots: resp.Snapshots,
	}, nil
}

// LoadSnapshotChunk implements CometBFT's Application interface
func (a *CometBFTAdapter) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	resp, err := a.app.LoadSnapshotChunk(ctx, &localabci.LoadSnapshotChunkRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseLoadSnapshotChunk{
		Chunk: resp.Chunk,
	}, nil
}

// OfferSnapshot implements CometBFT's Application interface
func (a *CometBFTAdapter) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	resp, err := a.app.OfferSnapshot(ctx, &localabci.OfferSnapshotRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseOfferSnapshot{
		Result: cmtabci.ResponseOfferSnapshot_Result(resp.Result),
	}, nil
}

// ApplySnapshotChunk implements CometBFT's Application interface
func (a *CometBFTAdapter) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	resp, err := a.app.ApplySnapshotChunk(ctx, &localabci.ApplySnapshotChunkRequest{})
	if err != nil {
		return nil, err
	}
	return &cmtabci.ResponseApplySnapshotChunk{
		Result: cmtabci.ResponseApplySnapshotChunk_Result(resp.Result),
	}, nil
}
