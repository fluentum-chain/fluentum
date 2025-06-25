package proxy

import (
	"context"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
)

type defaultAppConn struct {
	client abcicli.Client
}

// Mempool implementation

type mempoolConn struct{ defaultAppConn }

func (a *mempoolConn) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	return a.client.CheckTx(ctx, req)
}

func (a *mempoolConn) CheckTxAsync(req *types.RequestCheckTx) *abcicli.ReqRes {
	return a.client.CheckTxAsync(req)
}

func (a *mempoolConn) Flush(ctx context.Context) error {
	return a.client.Flush(ctx)
}

// Consensus implementation

type consensusConn struct{ defaultAppConn }

func (a *consensusConn) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	return a.client.FinalizeBlock(ctx, req)
}
func (a *consensusConn) PrepareProposal(ctx context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	return a.client.PrepareProposal(ctx, req)
}
func (a *consensusConn) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	return a.client.ProcessProposal(ctx, req)
}
func (a *consensusConn) ExtendVote(ctx context.Context, req *types.RequestExtendVote) (*types.ResponseExtendVote, error) {
	return a.client.ExtendVote(ctx, req)
}
func (a *consensusConn) VerifyVoteExtension(ctx context.Context, req *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error) {
	return a.client.VerifyVoteExtension(ctx, req)
}
func (a *consensusConn) Commit(ctx context.Context) (*types.ResponseCommit, error) {
	return a.client.Commit(ctx)
}

// Query implementation

type queryConn struct{ defaultAppConn }

func (a *queryConn) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	return a.client.Info(ctx, req)
}
func (a *queryConn) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	return a.client.Query(ctx, req)
}
func (a *queryConn) ABCIInfo(ctx context.Context) (*types.ResponseInfo, error) {
	return a.client.Info(ctx, &types.RequestInfo{})
}

// Snapshot implementation

type snapshotConn struct{ defaultAppConn }

func (a *snapshotConn) ListSnapshots(ctx context.Context, req *types.RequestListSnapshots) (*types.ResponseListSnapshots, error) {
	return a.client.ListSnapshots(ctx, req)
}
func (a *snapshotConn) OfferSnapshot(ctx context.Context, req *types.RequestOfferSnapshot) (*types.ResponseOfferSnapshot, error) {
	return a.client.OfferSnapshot(ctx, req)
}
func (a *snapshotConn) LoadSnapshotChunk(ctx context.Context, req *types.RequestLoadSnapshotChunk) (*types.ResponseLoadSnapshotChunk, error) {
	return a.client.LoadSnapshotChunk(ctx, req)
}
func (a *snapshotConn) ApplySnapshotChunk(ctx context.Context, req *types.RequestApplySnapshotChunk) (*types.ResponseApplySnapshotChunk, error) {
	return a.client.ApplySnapshotChunk(ctx, req)
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
