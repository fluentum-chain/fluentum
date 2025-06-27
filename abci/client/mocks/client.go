// Manual mock for ABCI Client interface
package client

import (
	"context"

	abcicli "github.com/fluentum-chain/fluentum/abci/client"
	tmlog "github.com/fluentum-chain/fluentum/libs/log"

	mock "github.com/stretchr/testify/mock"

	cmtabci "github.com/cometbft/cometbft/abci/types"
)

// MockClient is a mock implementation of the Client interface
type MockClient struct {
	mock.Mock
}

// Echo method for testing
func (m *MockClient) Echo(ctx context.Context, msg string) (*cmtabci.ResponseEcho, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(*cmtabci.ResponseEcho), args.Error(1)
}

// CheckTx method
func (m *MockClient) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) (*cmtabci.ResponseCheckTx, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseCheckTx), args.Error(1)
}

// CheckTxAsync method
func (m *MockClient) CheckTxAsync(ctx context.Context, req *cmtabci.RequestCheckTx) *abcicli.ReqRes {
	args := m.Called(ctx, req)
	return args.Get(0).(*abcicli.ReqRes)
}

// Flush method
func (m *MockClient) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// FinalizeBlock method
func (m *MockClient) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) (*cmtabci.ResponseFinalizeBlock, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseFinalizeBlock), args.Error(1)
}

// FinalizeBlockAsync method
func (m *MockClient) FinalizeBlockAsync(ctx context.Context, req *cmtabci.RequestFinalizeBlock) *abcicli.ReqRes {
	args := m.Called(ctx, req)
	return args.Get(0).(*abcicli.ReqRes)
}

// PrepareProposal method
func (m *MockClient) PrepareProposal(ctx context.Context, req *cmtabci.RequestPrepareProposal) (*cmtabci.ResponsePrepareProposal, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponsePrepareProposal), args.Error(1)
}

// ProcessProposal method
func (m *MockClient) ProcessProposal(ctx context.Context, req *cmtabci.RequestProcessProposal) (*cmtabci.ResponseProcessProposal, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseProcessProposal), args.Error(1)
}

// ExtendVote method
func (m *MockClient) ExtendVote(ctx context.Context, req *cmtabci.RequestExtendVote) (*cmtabci.ResponseExtendVote, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseExtendVote), args.Error(1)
}

// VerifyVoteExtension method
func (m *MockClient) VerifyVoteExtension(ctx context.Context, req *cmtabci.RequestVerifyVoteExtension) (*cmtabci.ResponseVerifyVoteExtension, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseVerifyVoteExtension), args.Error(1)
}

// Commit method
func (m *MockClient) Commit(ctx context.Context) (*cmtabci.ResponseCommit, error) {
	args := m.Called(ctx)
	return args.Get(0).(*cmtabci.ResponseCommit), args.Error(1)
}

// CommitAsync method
func (m *MockClient) CommitAsync(ctx context.Context) *abcicli.ReqRes {
	args := m.Called(ctx)
	return args.Get(0).(*abcicli.ReqRes)
}

// InitChain method
func (m *MockClient) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) (*cmtabci.ResponseInitChain, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseInitChain), args.Error(1)
}

// Info method
func (m *MockClient) Info(ctx context.Context, req *cmtabci.RequestInfo) (*cmtabci.ResponseInfo, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseInfo), args.Error(1)
}

// Query method
func (m *MockClient) Query(ctx context.Context, req *cmtabci.RequestQuery) (*cmtabci.ResponseQuery, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseQuery), args.Error(1)
}

// ListSnapshots method
func (m *MockClient) ListSnapshots(ctx context.Context, req *cmtabci.RequestListSnapshots) (*cmtabci.ResponseListSnapshots, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseListSnapshots), args.Error(1)
}

// OfferSnapshot method
func (m *MockClient) OfferSnapshot(ctx context.Context, req *cmtabci.RequestOfferSnapshot) (*cmtabci.ResponseOfferSnapshot, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseOfferSnapshot), args.Error(1)
}

// LoadSnapshotChunk method
func (m *MockClient) LoadSnapshotChunk(ctx context.Context, req *cmtabci.RequestLoadSnapshotChunk) (*cmtabci.ResponseLoadSnapshotChunk, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseLoadSnapshotChunk), args.Error(1)
}

// ApplySnapshotChunk method
func (m *MockClient) ApplySnapshotChunk(ctx context.Context, req *cmtabci.RequestApplySnapshotChunk) (*cmtabci.ResponseApplySnapshotChunk, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*cmtabci.ResponseApplySnapshotChunk), args.Error(1)
}

// Error method
func (m *MockClient) Error() error {
	args := m.Called()
	return args.Error(0)
}

// SetResponseCallback method
func (m *MockClient) SetResponseCallback(cb abcicli.Callback) {
	m.Called(cb)
}

// SetLogger method
func (m *MockClient) SetLogger(logger tmlog.Logger) {
	m.Called(logger)
}

// Close method
func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
