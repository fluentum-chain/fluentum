package client

import (
	"context"
	"testing"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
)

// Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	// Test local client
	var _ Client = (*localClient)(nil)
	
	// Test gRPC client
	var _ Client = (*grpcClient)(nil)
	
	// Test socket client
	var _ Client = (*socketClient)(nil)
}

// Mock application for testing
type mockApplication struct {
	cmtabci.Application
}

func (app *mockApplication) CheckTx(ctx context.Context, req *cmtabci.RequestCheckTx) *cmtabci.ResponseCheckTx {
	return &cmtabci.ResponseCheckTx{
		Code: 0,
		Data: []byte("ok"),
		Log:  "success",
	}
}

func (app *mockApplication) FinalizeBlock(ctx context.Context, req *cmtabci.RequestFinalizeBlock) *cmtabci.ResponseFinalizeBlock {
	return &cmtabci.ResponseFinalizeBlock{
		TxResults: []*cmtabci.ExecTxResult{
			{
				Code: 0,
				Data: []byte("ok"),
				Log:  "success",
			},
		},
		AppHash: []byte("app_hash"),
	}
}

func (app *mockApplication) Commit(ctx context.Context, req *cmtabci.RequestCommit) *cmtabci.ResponseCommit {
	return &cmtabci.ResponseCommit{
		Data: []byte("commit_data"),
	}
}

func (app *mockApplication) InitChain(ctx context.Context, req *cmtabci.RequestInitChain) *cmtabci.ResponseInitChain {
	return &cmtabci.ResponseInitChain{
		AppHash: []byte("init_hash"),
	}
}

func (app *mockApplication) Info(ctx context.Context, req *cmtabci.RequestInfo) *cmtabci.ResponseInfo {
	return &cmtabci.ResponseInfo{
		Data: "mock_app",
	}
}

func (app *mockApplication) Query(ctx context.Context, req *cmtabci.RequestQuery) *cmtabci.ResponseQuery {
	return &cmtabci.ResponseQuery{
		Code: 0,
		Value: []byte("query_result"),
	}
}

// Test local client
func TestLocalClient(t *testing.T) {
	app := &mockApplication{}
	client := NewLocalClient(app)

	// Test CheckTx
	res, err := client.CheckTx(context.Background(), &cmtabci.RequestCheckTx{
		Tx: []byte("test_tx"),
	})
	require.NoError(t, err)
	require.Equal(t, uint32(0), res.Code)
	require.Equal(t, []byte("ok"), res.Data)

	// Test FinalizeBlock
	res2, err := client.FinalizeBlock(context.Background(), &cmtabci.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{[]byte("test")},
	})
	require.NoError(t, err)
	require.Len(t, res2.TxResults, 1)
	require.Equal(t, uint32(0), res2.TxResults[0].Code)

	// Test Commit
	res3, err := client.Commit(context.Background(), &cmtabci.RequestCommit{})
	require.NoError(t, err)
	require.Equal(t, []byte("commit_data"), res3.Data)

	// Test InitChain
	res4, err := client.InitChain(context.Background(), &cmtabci.RequestInitChain{})
	require.NoError(t, err)
	require.Equal(t, []byte("init_hash"), res4.AppHash)

	// Test Info
	res5, err := client.Info(context.Background(), &cmtabci.RequestInfo{})
	require.NoError(t, err)
	require.Equal(t, "mock_app", res5.Data)

	// Test Query
	res6, err := client.Query(context.Background(), &cmtabci.RequestQuery{})
	require.NoError(t, err)
	require.Equal(t, uint32(0), res6.Code)
	require.Equal(t, []byte("query_result"), res6.Value)
}

// Test validation helpers
func TestValidationHelpers(t *testing.T) {
	// Test validateBlockHeight
	err := validateBlockHeight(0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid block height")

	err = validateBlockHeight(-1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid block height")

	err = validateBlockHeight(1)
	require.NoError(t, err)

	// Test validateTxData
	err = validateTxData(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty transaction data")

	err = validateTxData([]byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty transaction data")

	err = validateTxData([]byte("valid_tx"))
	require.NoError(t, err)
}

// Test ReqRes functionality
func TestReqRes(t *testing.T) {
	req := &cmtabci.RequestCheckTx{Tx: []byte("test")}
	reqRes := NewReqRes(&cmtabci.Request{Value: &cmtabci.Request_CheckTx{CheckTx: req}})

	require.NotNil(t, reqRes.Request)
	require.NotNil(t, reqRes.DoneCh)
	require.NotNil(t, reqRes.ResponseCh)
	require.NotNil(t, reqRes.ErrorCh)

	// Test Done
	go func() {
		reqRes.Done()
	}()
	reqRes.Wait()

	// Test InvokeCallback
	called := false
	reqRes.ResponseCb = func(res *cmtabci.Response, err error) {
		called = true
	}
	reqRes.InvokeCallback()
	require.True(t, called)
} 
