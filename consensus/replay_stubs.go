package consensus

import (
	"context"

	cometbftabci "github.com/cometbft/cometbft/abci/types"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/libs/clist"
	mempl "github.com/fluentum-chain/fluentum/mempool"
	tmstate "github.com/fluentum-chain/fluentum/proto/tendermint/state"
	"github.com/fluentum-chain/fluentum/proxy"
	"github.com/fluentum-chain/fluentum/types"
)

//-----------------------------------------------------------------------------

type emptyMempool struct{}

var _ mempl.Mempool = emptyMempool{}

func (emptyMempool) Lock()            {}
func (emptyMempool) Unlock()          {}
func (emptyMempool) Size() int        { return 0 }
func (emptyMempool) SizeBytes() int64 { return 0 }
func (emptyMempool) CheckTx(_ types.Tx, _ func(*abci.Response), _ mempl.TxInfo) error {
	return nil
}

func (txmp emptyMempool) RemoveTxByKey(txKey types.TxKey) error {
	return nil
}

func (emptyMempool) ReapMaxBytesMaxGas(_, _ int64) types.Txs { return types.Txs{} }
func (emptyMempool) ReapMaxTxs(n int) types.Txs              { return types.Txs{} }
func (emptyMempool) Update(
	_ int64,
	_ types.Txs,
	_ []*cometbftabci.ExecTxResult,
	_ mempl.PreCheckFunc,
	_ mempl.PostCheckFunc,
) error {
	return nil
}
func (emptyMempool) Flush()                        {}
func (emptyMempool) FlushAppConn() error           { return nil }
func (emptyMempool) TxsAvailable() <-chan struct{} { return make(chan struct{}) }
func (emptyMempool) EnableTxsAvailable()           {}
func (emptyMempool) TxsBytes() int64               { return 0 }

func (emptyMempool) TxsFront() *clist.CElement    { return nil }
func (emptyMempool) TxsWaitChan() <-chan struct{} { return nil }

func (emptyMempool) InitWAL() error { return nil }
func (emptyMempool) CloseWAL()      {}

//-----------------------------------------------------------------------------
// mockProxyApp uses ABCIResponses to give the right results.
//
// Useful because we don't want to call Commit() twice for the same block on
// the real app.

func newMockProxyApp(appHash []byte, abciResponses *tmstate.ABCIResponses) proxy.AppConnConsensus {
	// TODO: Fix mockProxyApp interface compatibility issues
	// clientCreator := proxy.NewLocalClientCreator(&mockProxyApp{
	// 	appHash:       appHash,
	// 	abciResponses: abciResponses,
	// })
	// cli, _ := clientCreator.NewABCIClient()
	// err := cli.Start()
	// if err != nil {
	// 	panic(err)
	// }
	// return proxy.NewAppConnConsensus(cli)

	// Return a simple mock for now
	return &mockAppConnConsensus{}
}

type mockAppConnConsensus struct{}

func (m *mockAppConnConsensus) FinalizeBlock(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	return &abci.FinalizeBlockResponse{}, nil
}

func (m *mockAppConnConsensus) PrepareProposal(ctx context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	return &abci.PrepareProposalResponse{}, nil
}

func (m *mockAppConnConsensus) ProcessProposal(ctx context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	return &abci.ProcessProposalResponse{}, nil
}

func (m *mockAppConnConsensus) ExtendVote(ctx context.Context, req *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	return &abci.ExtendVoteResponse{}, nil
}

func (m *mockAppConnConsensus) VerifyVoteExtension(ctx context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	return &abci.VerifyVoteExtensionResponse{}, nil
}

func (m *mockAppConnConsensus) Commit(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	return &abci.CommitResponse{}, nil
}

func (m *mockAppConnConsensus) CommitSync(ctx context.Context, req *abci.CommitRequest) (*abci.CommitResponse, error) {
	return &abci.CommitResponse{}, nil
}
