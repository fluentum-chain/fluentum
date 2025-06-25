package mocks

import (
	"context"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
)

// MockAppConnMempool implements AppConnMempool for testing
type MockAppConnMempool struct {
	CheckTxFn      func(context.Context, *types.RequestCheckTx) (*types.ResponseCheckTx, error)
	CheckTxAsyncFn func(*types.RequestCheckTx) *abcicli.ReqRes
	FlushFn        func(context.Context) error
}

func (m *MockAppConnMempool) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	if m.CheckTxFn == nil {
		return &types.ResponseCheckTx{Code: 0}, nil
	}
	return m.CheckTxFn(ctx, req)
}

func (m *MockAppConnMempool) CheckTxAsync(req *types.RequestCheckTx) *abcicli.ReqRes {
	if m.CheckTxAsyncFn == nil {
		return abcicli.NewReqRes(req)
	}
	return m.CheckTxAsyncFn(req)
}

func (m *MockAppConnMempool) Flush(ctx context.Context) error {
	if m.FlushFn == nil {
		return nil
	}
	return m.FlushFn(ctx)
}

// MockAppConnConsensus implements AppConnConsensus for testing
type MockAppConnConsensus struct {
	FinalizeBlockFn       func(context.Context, *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error)
	PrepareProposalFn     func(context.Context, *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error)
	ProcessProposalFn     func(context.Context, *types.RequestProcessProposal) (*types.ResponseProcessProposal, error)
	ExtendVoteFn          func(context.Context, *types.RequestExtendVote) (*types.ResponseExtendVote, error)
	VerifyVoteExtensionFn func(context.Context, *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error)
	CommitFn              func(context.Context) (*types.ResponseCommit, error)
}

func (m *MockAppConnConsensus) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	if m.FinalizeBlockFn == nil {
		return &types.ResponseFinalizeBlock{}, nil
	}
	return m.FinalizeBlockFn(ctx, req)
}

func (m *MockAppConnConsensus) PrepareProposal(ctx context.Context, req *types.RequestPrepareProposal) (*types.ResponsePrepareProposal, error) {
	if m.PrepareProposalFn == nil {
		return &types.ResponsePrepareProposal{}, nil
	}
	return m.PrepareProposalFn(ctx, req)
}

func (m *MockAppConnConsensus) ProcessProposal(ctx context.Context, req *types.RequestProcessProposal) (*types.ResponseProcessProposal, error) {
	if m.ProcessProposalFn == nil {
		return &types.ResponseProcessProposal{}, nil
	}
	return m.ProcessProposalFn(ctx, req)
}

func (m *MockAppConnConsensus) ExtendVote(ctx context.Context, req *types.RequestExtendVote) (*types.ResponseExtendVote, error) {
	if m.ExtendVoteFn == nil {
		return &types.ResponseExtendVote{}, nil
	}
	return m.ExtendVoteFn(ctx, req)
}

func (m *MockAppConnConsensus) VerifyVoteExtension(ctx context.Context, req *types.RequestVerifyVoteExtension) (*types.ResponseVerifyVoteExtension, error) {
	if m.VerifyVoteExtensionFn == nil {
		return &types.ResponseVerifyVoteExtension{}, nil
	}
	return m.VerifyVoteExtensionFn(ctx, req)
}

func (m *MockAppConnConsensus) Commit(ctx context.Context) (*types.ResponseCommit, error) {
	if m.CommitFn == nil {
		return &types.ResponseCommit{}, nil
	}
	return m.CommitFn(ctx)
}

// Helper to create ready-to-use mocks
func NewMockMempool() *MockAppConnMempool {
	return &MockAppConnMempool{
		CheckTxFn: func(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
			return &types.ResponseCheckTx{Code: 0}, nil
		},
	}
}
