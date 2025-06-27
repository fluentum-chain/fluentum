package proxy

import (
	"context"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	abcicli "github.com/fluentum-chain/fluentum/abci/client"
	abci "github.com/fluentum-chain/fluentum/abci/types"
)

type defaultAppConn struct {
	client abcicli.Client
}

// Mempool implementation

type mempoolConn struct{ defaultAppConn }

func (a *mempoolConn) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	cmtReq := &cmtabci.RequestCheckTx{
		Tx:   req.Tx,
		Type: cmtabci.CheckTxType(req.Type),
	}
	res, err := a.client.CheckTx(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.CheckTxResponse{
		Code:      res.Code,
		Data:      res.Data,
		Log:       res.Log,
		Info:      res.Info,
		GasWanted: res.GasWanted,
		GasUsed:   res.GasUsed,
		Events:    res.Events,
		Codespace: res.Codespace,
	}, nil
}

func (a *mempoolConn) CheckTxAsync(req *abci.CheckTxRequest) *abcicli.ReqRes {
	cmtReq := &cmtabci.RequestCheckTx{
		Tx:   req.Tx,
		Type: cmtabci.CheckTxType(req.Type),
	}
	reqRes := a.client.CheckTxAsync(context.Background(), cmtReq)
	return reqRes
}

func (a *mempoolConn) Flush(ctx context.Context) error {
	return a.client.Flush(ctx)
}

func (a *mempoolConn) SetResponseCallback(cb func(*abci.Request, *abci.Response)) {
	a.client.SetResponseCallback(cb)
}

func (a *mempoolConn) Error() error {
	return a.client.Error()
}

func (a *mempoolConn) FlushAsync() *abcicli.ReqRes {
	return a.client.FlushAsync(context.Background())
}

// Consensus implementation

type consensusConn struct{ defaultAppConn }

func (a *consensusConn) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	cmtReq := &cmtabci.RequestFinalizeBlock{
		Txs:                req.Txs,
		DecidedLastCommit:  req.DecidedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}
	res, err := a.client.FinalizeBlock(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.FinalizeBlockResponse{
		Events:                res.Events,
		TxResults:             res.TxResults,
		ValidatorUpdates:      res.ValidatorUpdates,
		ConsensusParamUpdates: res.ConsensusParamUpdates,
		AppHash:               res.AppHash,
	}, nil
}

func (a *consensusConn) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	cmtReq := &cmtabci.RequestPrepareProposal{
		MaxTxBytes: req.MaxTxBytes,
		Txs:        req.Txs,
		LocalLastCommit: cmtabci.ExtendedCommitInfo{
			Round: req.LocalLastCommit.Round,
			Votes: req.LocalLastCommit.Votes,
		},
		Misbehavior:        req.Misbehavior,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}
	res, err := a.client.PrepareProposal(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.PrepareProposalResponse{
		Txs: res.Txs,
	}, nil
}

func (a *consensusConn) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	cmtReq := &cmtabci.RequestProcessProposal{
		Txs:                req.Txs,
		ProposedLastCommit: req.ProposedLastCommit,
		Misbehavior:        req.Misbehavior,
		Hash:               req.Hash,
		Height:             req.Height,
		Time:               req.Time,
		NextValidatorsHash: req.NextValidatorsHash,
		ProposerAddress:    req.ProposerAddress,
	}
	res, err := a.client.ProcessProposal(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.ProcessProposalResponse{
		Status: res.Status,
	}, nil
}

func (a *consensusConn) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	cmtReq := &cmtabci.RequestExtendVote{
		Hash:   req.Hash,
		Height: req.Height,
	}
	res, err := a.client.ExtendVote(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.ExtendVoteResponse{
		VoteExtension: res.VoteExtension,
	}, nil
}

func (a *consensusConn) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	cmtReq := &cmtabci.RequestVerifyVoteExtension{
		Hash:             req.Hash,
		ValidatorAddress: req.ValidatorAddress,
		Height:           req.Height,
		VoteExtension:    req.VoteExtension,
	}
	res, err := a.client.VerifyVoteExtension(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.VerifyVoteExtensionResponse{
		Status: res.Status,
	}, nil
}

func (a *consensusConn) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	res, err := a.client.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return &abci.CommitResponse{
		RetainHeight: res.RetainHeight,
	}, nil
}

func (a *consensusConn) CommitSync(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	return a.Commit(ctx, nil)
}

// Query implementation

type queryConn struct{ defaultAppConn }

func (a *queryConn) Info(ctx context.Context, req *abci.InfoRequest) (*abci.InfoResponse, error) {
	cmtReq := &cmtabci.RequestInfo{
		Version:      req.Version,
		BlockVersion: req.BlockVersion,
		P2PVersion:   req.P2PVersion,
	}
	res, err := a.client.Info(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.InfoResponse{
		Data:             res.Data,
		Version:          res.Version,
		AppVersion:       res.AppVersion,
		LastBlockHeight:  res.LastBlockHeight,
		LastBlockAppHash: res.LastBlockAppHash,
	}, nil
}

func (a *queryConn) Query(ctx context.Context, req *abci.QueryRequest) (*abci.QueryResponse, error) {
	cmtReq := &cmtabci.RequestQuery{
		Data:   req.Data,
		Path:   req.Path,
		Height: req.Height,
		Prove:  req.Prove,
	}
	res, err := a.client.Query(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.QueryResponse{
		Code:      res.Code,
		Log:       res.Log,
		Info:      res.Info,
		Index:     res.Index,
		Key:       res.Key,
		Value:     res.Value,
		ProofOps:  res.ProofOps,
		Height:    res.Height,
		Codespace: res.Codespace,
	}, nil
}

func (a *queryConn) ABCIInfo(ctx context.Context) (*abci.InfoResponse, error) {
	cmtReq := &cmtabci.RequestInfo{}
	res, err := a.client.Info(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.InfoResponse{
		Data:             res.Data,
		Version:          res.Version,
		AppVersion:       res.AppVersion,
		LastBlockHeight:  res.LastBlockHeight,
		LastBlockAppHash: res.LastBlockAppHash,
	}, nil
}

// Snapshot implementation

type snapshotConn struct{ defaultAppConn }

func (a *snapshotConn) ListSnapshots(ctx context.Context, req *abci.ListSnapshotsRequest) (*abci.ListSnapshotsResponse, error) {
	cmtReq := &cmtabci.RequestListSnapshots{}
	res, err := a.client.ListSnapshots(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.ListSnapshotsResponse{
		Snapshots: res.Snapshots,
	}, nil
}

func (a *snapshotConn) OfferSnapshot(ctx context.Context, req *abci.OfferSnapshotRequest) (*abci.OfferSnapshotResponse, error) {
	cmtReq := &cmtabci.RequestOfferSnapshot{
		Snapshot: req.Snapshot,
		AppHash:  req.AppHash,
	}
	res, err := a.client.OfferSnapshot(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.OfferSnapshotResponse{
		Result: abci.ResponseOfferSnapshot_Result(res.Result),
	}, nil
}

func (a *snapshotConn) LoadSnapshotChunk(ctx context.Context, req *abci.LoadSnapshotChunkRequest) (*abci.LoadSnapshotChunkResponse, error) {
	cmtReq := &cmtabci.RequestLoadSnapshotChunk{
		Height: req.Height,
		Format: req.Format,
		Chunk:  req.Chunk,
	}
	res, err := a.client.LoadSnapshotChunk(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.LoadSnapshotChunkResponse{
		Chunk: res.Chunk,
	}, nil
}

func (a *snapshotConn) ApplySnapshotChunk(ctx context.Context, req *abci.ApplySnapshotChunkRequest) (*abci.ApplySnapshotChunkResponse, error) {
	cmtReq := &cmtabci.RequestApplySnapshotChunk{
		Index:  req.Index,
		Chunk:  req.Chunk,
		Sender: req.Sender,
	}
	res, err := a.client.ApplySnapshotChunk(ctx, cmtReq)
	if err != nil {
		return nil, err
	}
	return &abci.ApplySnapshotChunkResponse{
		Result:        abci.ResponseApplySnapshotChunk_Result(res.Result),
		RefetchChunks: res.RefetchChunks,
		RejectSenders: res.RejectSenders,
	}, nil
}

// Factory functions

func NewAppConnMempool(client abcicli.Client) AppConnMempool {
	return &mempoolConn{defaultAppConn{client}}
}

func NewAppConnConsensus(client abcicli.Client) AppConnConsensus {
	return &consensusConn{defaultAppConn{client}}
}

func NewAppConnQuery(client abcicli.Client) AppConnQuery {
	return &queryConn{defaultAppConn{client}}
}

func NewAppConnSnapshot(client abcicli.Client) AppConnSnapshot {
	return &snapshotConn{defaultAppConn{client}}
}

// Interface compliance checks
var _ AppConnMempool = (*mempoolConn)(nil)
var _ AppConnConsensus = (*consensusConn)(nil)
var _ AppConnQuery = (*queryConn)(nil)
var _ AppConnSnapshot = (*snapshotConn)(nil)
