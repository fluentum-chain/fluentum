package mocks

import (
	"context"

	abcicli "github.com/cometbft/cometbft/abci/client"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	abci "github.com/fluentum-chain/fluentum/abci/types"
)

// MockAppConnMempool implements AppConnMempool for testing
type MockAppConnMempool struct {
	CheckTxFn      func(context.Context, *abci.CheckTxRequest) (*abci.CheckTxResponse, error)
	CheckTxAsyncFn func(*abci.CheckTxRequest) *abcicli.ReqRes
	FlushFn        func(context.Context) error
}

func (m *MockAppConnMempool) CheckTx(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	if m.CheckTxFn == nil {
		return &abci.CheckTxResponse{Code: 0}, nil
	}
	return m.CheckTxFn(ctx, req)
}

func (m *MockAppConnMempool) CheckTxAsync(req *abci.CheckTxRequest) *abcicli.ReqRes {
	if m.CheckTxAsyncFn == nil {
		cmtReq := &cmtabci.Request{
			Value: &cmtabci.Request_CheckTx{
				CheckTx: &cmtabci.RequestCheckTx{
					Tx:   req.Tx,
					Type: cmtabci.CheckTxType(req.Type),
				},
			},
		}
		return abcicli.NewReqRes(cmtReq)
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
	FinalizeBlockFn       func(context.Context, *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error)
	PrepareProposalFn     func(context.Context, *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error)
	ProcessProposalFn     func(context.Context, *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error)
	ExtendVoteFn          func(context.Context, *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error)
	VerifyVoteExtensionFn func(context.Context, *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error)
	CommitFn              func(context.Context, *abci.CommitRequest) (*abci.CommitResponse, error)
}

func (m *MockAppConnConsensus) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	if m.FinalizeBlockFn == nil {
		return &abci.FinalizeBlockResponse{}, nil
	}
	return m.FinalizeBlockFn(ctx, req)
}

func (m *MockAppConnConsensus) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	if m.PrepareProposalFn == nil {
		return &abci.PrepareProposalResponse{}, nil
	}
	return m.PrepareProposalFn(ctx, req)
}

func (m *MockAppConnConsensus) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	if m.ProcessProposalFn == nil {
		return &abci.ProcessProposalResponse{}, nil
	}
	return m.ProcessProposalFn(ctx, req)
}

func (m *MockAppConnConsensus) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	if m.ExtendVoteFn == nil {
		return &abci.ExtendVoteResponse{}, nil
	}
	return m.ExtendVoteFn(ctx, req)
}

func (m *MockAppConnConsensus) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	if m.VerifyVoteExtensionFn == nil {
		return &abci.VerifyVoteExtensionResponse{}, nil
	}
	return m.VerifyVoteExtensionFn(ctx, req)
}

func (m *MockAppConnConsensus) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	if m.CommitFn == nil {
		return &abci.CommitResponse{}, nil
	}
	return m.CommitFn(ctx, req)
}

func (m *MockAppConnConsensus) CommitSync(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	return m.Commit(ctx, req)
}

// Helper to create ready-to-use mocks
func NewMockMempool() *MockAppConnMempool {
	return &MockAppConnMempool{
		CheckTxFn: func(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
			return &abci.CheckTxResponse{Code: 0}, nil
		},
	}
}
