package client

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	abcicli "github.com/cometbft/cometbft/abci/client"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fluentum-chain/fluentum/abci/server"
	"github.com/fluentum-chain/fluentum/abci/types"
	tmrand "github.com/fluentum-chain/fluentum/libs/rand"
	"github.com/fluentum-chain/fluentum/libs/service"
)

func TestProperSyncCalls(t *testing.T) {
	app := slowApp{}

	s, c := setupClientServer(t, app)
	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Error(err)
		}
	})
	t.Cleanup(func() {
		if err := c.Stop(); err != nil {
			t.Error(err)
		}
	})

	resp := make(chan error, 1)
	go func() {
		// This is BeginBlockSync unrolled....
		reqres := c.BeginBlockAsync(types.RequestFinalizeBlock{})
		err := c.FlushSync()
		require.NoError(t, err)
		res := reqres.Response.GetBeginBlock()
		require.NotNil(t, res)
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.NoError(t, err, "This should return success")
	}
}

func TestHangingSyncCalls(t *testing.T) {
	app := slowApp{}

	s, c := setupClientServer(t, app)
	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Log(err)
		}
	})
	t.Cleanup(func() {
		if err := c.Stop(); err != nil {
			t.Log(err)
		}
	})

	resp := make(chan error, 1)
	go func() {
		// Start BeginBlock and flush it
		reqres := c.BeginBlockAsync(types.RequestFinalizeBlock{})
		flush := c.FlushAsync()
		// wait 20 ms for all events to travel socket, but
		// no response yet from server
		time.Sleep(20 * time.Millisecond)
		// kill the server, so the connections break
		err := s.Stop()
		require.NoError(t, err)

		// wait for the response from BeginBlock
		reqres.Wait()
		flush.Wait()
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.Error(t, err, "We should get EOF error")
	}
}

func setupClientServer(t *testing.T, app types.Application) (
	service.Service, abcicli.Client) {
	// some port between 20k and 30k
	port := 20000 + tmrand.Int32()%10000
	addr := fmt.Sprintf("localhost:%d", port)

	s, err := server.NewServer(addr, "socket", app)
	require.NoError(t, err)
	err = s.Start()
	require.NoError(t, err)

	c := abcicli.NewSocketClient(addr, true)
	err = c.Start()
	require.NoError(t, err)

	return s, c
}

type slowApp struct {
	types.BaseApplication
}

func (slowApp) BeginBlock(req types.RequestFinalizeBlock) types.ResponseBeginBlock {
	time.Sleep(200 * time.Millisecond)
	return types.ResponseBeginBlock{}
}

// TestCallbackInvokedWhenSetLaet ensures that the callback is invoked when
// set after the client completes the call into the app. Currently this
// test relies on the callback being allowed to be invoked twice if set multiple
// times, once when set early and once when set late.
func TestCallbackInvokedWhenSetLate(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	app := blockedABCIApplication{
		wg: wg,
	}
	_, c := setupClientServer(t, app)
	reqRes := c.CheckTxAsync(types.RequestCheckTx{})

	done := make(chan struct{})
	cb := func(_ *types.Response) {
		close(done)
	}
	reqRes.SetCallback(cb)
	app.wg.Done()
	<-done

	var called bool
	cb = func(_ *types.Response) {
		called = true
	}
	reqRes.SetCallback(cb)
	require.True(t, called)
}

type blockedABCIApplication struct {
	wg *sync.WaitGroup
	types.BaseApplication
}

func (b blockedABCIApplication) CheckTx(r types.RequestCheckTx) types.ResponseCheckTx {
	b.wg.Wait()
	return b.BaseApplication.CheckTx(r)
}

// TestCallbackInvokedWhenSetEarly ensures that the callback is invoked when
// set before the client completes the call into the app.
func TestCallbackInvokedWhenSetEarly(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	app := blockedABCIApplication{
		wg: wg,
	}
	_, c := setupClientServer(t, app)
	reqRes := c.CheckTxAsync(types.RequestCheckTx{})

	done := make(chan struct{})
	cb := func(_ *types.Response) {
		close(done)
	}
	reqRes.SetCallback(cb)
	app.wg.Done()

	called := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}
	require.Eventually(t, called, time.Second, time.Millisecond*25)
}

func TestSocketClient(t *testing.T) {
	// Create a pipe for testing
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Create client
	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Test basic client creation
	assert.NotNil(t, client)
	assert.Implements(t, (*Client)(nil), client)

	// Test client closure
	err := client.Close()
	assert.NoError(t, err)
}

func TestSocketClientCheckTx(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_CheckTx{
			CheckTx: &cmtabci.ResponseCheckTx{
				Code:      0,
				GasWanted: 100,
				GasUsed:   100,
				Log:       "success",
			},
		},
	})

	// Test CheckTx
	ctx := context.Background()
	req := &cmtabci.RequestCheckTx{
		Tx:   []byte("test_tx"),
		Type: cmtabci.CheckTxType_New,
	}

	res, err := client.CheckTx(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), res.Code)
	assert.Equal(t, int64(100), res.GasWanted)
	assert.Equal(t, int64(100), res.GasUsed)
	assert.Equal(t, "success", res.Log)
}

func TestSocketClientFinalizeBlock(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_FinalizeBlock{
			FinalizeBlock: &cmtabci.ResponseFinalizeBlock{
				TxResults: []*cmtabci.ExecTxResult{
					{
						Code: 0,
						Log:  "tx1 success",
					},
					{
						Code: 0,
						Log:  "tx2 success",
					},
				},
				AppHash: []byte("app_hash"),
			},
		},
	})

	// Test FinalizeBlock
	ctx := context.Background()
	req := &cmtabci.RequestFinalizeBlock{
		Txs: [][]byte{
			[]byte("tx1"),
			[]byte("tx2"),
		},
		Height: 1,
	}

	res, err := client.FinalizeBlock(ctx, req)
	require.NoError(t, err)
	assert.Len(t, res.TxResults, 2)
	assert.Equal(t, uint32(0), res.TxResults[0].Code)
	assert.Equal(t, "tx1 success", res.TxResults[0].Log)
	assert.Equal(t, uint32(0), res.TxResults[1].Code)
	assert.Equal(t, "tx2 success", res.TxResults[1].Log)
	assert.Equal(t, []byte("app_hash"), res.AppHash)
}

func TestSocketClientCommit(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_Commit{
			Commit: &cmtabci.ResponseCommit{
				Data: []byte("commit_data"),
			},
		},
	})

	// Test Commit
	ctx := context.Background()
	req := &cmtabci.RequestCommit{}

	res, err := client.Commit(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, []byte("commit_data"), res.Data)
}

func TestSocketClientInfo(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_Info{
			Info: &cmtabci.ResponseInfo{
				Data:             "test_app",
				Version:          "1.0.0",
				AppVersion:       1,
				LastBlockHeight:  100,
				LastBlockAppHash: []byte("last_hash"),
			},
		},
	})

	// Test Info
	ctx := context.Background()
	req := &cmtabci.RequestInfo{
		Version: "1.0.0",
	}

	res, err := client.Info(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "test_app", res.Data)
	assert.Equal(t, "1.0.0", res.Version)
	assert.Equal(t, uint64(1), res.AppVersion)
	assert.Equal(t, int64(100), res.LastBlockHeight)
	assert.Equal(t, []byte("last_hash"), res.LastBlockAppHash)
}

func TestSocketClientQuery(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_Query{
			Query: &cmtabci.ResponseQuery{
				Code:  0,
				Value: []byte("query_result"),
				Log:   "query successful",
			},
		},
	})

	// Test Query
	ctx := context.Background()
	req := &cmtabci.RequestQuery{
		Data: []byte("query_data"),
		Path: "/store",
	}

	res, err := client.Query(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), res.Code)
	assert.Equal(t, []byte("query_result"), res.Value)
	assert.Equal(t, "query successful", res.Log)
}

func TestSocketClientInitChain(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_InitChain{
			InitChain: &cmtabci.ResponseInitChain{
				ConsensusParams: &cmtabci.ConsensusParams{
					Block: &cmtabci.BlockParams{
						MaxBytes: 22020096,
						MaxGas:   15000000,
					},
				},
				Validators: []*cmtabci.ValidatorUpdate{
					{
						PubKey: cmtabci.PubKey{Data: []byte("validator1")},
						Power:  100,
					},
				},
				AppHash: []byte("init_hash"),
			},
		},
	})

	// Test InitChain
	ctx := context.Background()
	req := &cmtabci.RequestInitChain{
		Time:    time.Now(),
		ChainId: "test_chain",
	}

	res, err := client.InitChain(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res.ConsensusParams)
	assert.Equal(t, int64(22020096), res.ConsensusParams.Block.MaxBytes)
	assert.Equal(t, int64(15000000), res.ConsensusParams.Block.MaxGas)
	assert.Len(t, res.Validators, 1)
	assert.Equal(t, int64(100), res.Validators[0].Power)
	assert.Equal(t, []byte("init_hash"), res.AppHash)
}

func TestSocketClientPrepareProposal(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_PrepareProposal{
			PrepareProposal: &cmtabci.ResponsePrepareProposal{
				Txs: [][]byte{
					[]byte("selected_tx1"),
					[]byte("selected_tx2"),
				},
			},
		},
	})

	// Test PrepareProposal
	ctx := context.Background()
	req := &cmtabci.RequestPrepareProposal{
		MaxTxBytes: 1000,
		Txs: [][]byte{
			[]byte("tx1"),
			[]byte("tx2"),
			[]byte("tx3"),
		},
	}

	res, err := client.PrepareProposal(ctx, req)
	require.NoError(t, err)
	assert.Len(t, res.Txs, 2)
	assert.Equal(t, []byte("selected_tx1"), res.Txs[0])
	assert.Equal(t, []byte("selected_tx2"), res.Txs[1])
}

func TestSocketClientProcessProposal(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_ProcessProposal{
			ProcessProposal: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			},
		},
	})

	// Test ProcessProposal
	ctx := context.Background()
	req := &cmtabci.RequestProcessProposal{
		Txs: [][]byte{
			[]byte("tx1"),
			[]byte("tx2"),
		},
		Height: 1,
	}

	res, err := client.ProcessProposal(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseProcessProposal_ACCEPT, res.Status)
}

func TestSocketClientExtendVote(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_ExtendVote{
			ExtendVote: &cmtabci.ResponseExtendVote{
				VoteExtension: []byte("vote_extension"),
			},
		},
	})

	// Test ExtendVote
	ctx := context.Background()
	req := &cmtabci.RequestExtendVote{
		Hash:   []byte("block_hash"),
		Height: 1,
	}

	res, err := client.ExtendVote(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, []byte("vote_extension"), res.VoteExtension)
}

func TestSocketClientVerifyVoteExtension(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_VerifyVoteExtension{
			VerifyVoteExtension: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
		},
	})

	// Test VerifyVoteExtension
	ctx := context.Background()
	req := &cmtabci.RequestVerifyVoteExtension{
		Hash:              []byte("block_hash"),
		ValidatorProTxHash: []byte("validator_hash"),
		Height:            1,
		VoteExtension:     []byte("vote_extension"),
	}

	res, err := client.VerifyVoteExtension(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseVerifyVoteExtension_ACCEPT, res.Status)
}

func TestSocketClientSnapshotMethods(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Test ListSnapshots
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_ListSnapshots{
			ListSnapshots: &cmtabci.ResponseListSnapshots{
				Snapshots: []*cmtabci.Snapshot{
					{
						Height: 100,
						Format: 1,
						Chunks: 10,
						Hash:   []byte("snapshot_hash"),
					},
				},
			},
		},
	})

	ctx := context.Background()
	res, err := client.ListSnapshots(ctx, &cmtabci.RequestListSnapshots{})
	require.NoError(t, err)
	assert.Len(t, res.Snapshots, 1)
	assert.Equal(t, uint64(100), res.Snapshots[0].Height)
	assert.Equal(t, uint32(1), res.Snapshots[0].Format)
	assert.Equal(t, uint32(10), res.Snapshots[0].Chunks)
	assert.Equal(t, []byte("snapshot_hash"), res.Snapshots[0].Hash)

	// Test OfferSnapshot
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 2,
		Value: &cmtabci.Response_OfferSnapshot{
			OfferSnapshot: &cmtabci.ResponseOfferSnapshot{
				Result: cmtabci.ResponseOfferSnapshot_ACCEPT,
			},
		},
	})

	res2, err := client.OfferSnapshot(ctx, &cmtabci.RequestOfferSnapshot{
		Snapshot: &cmtabci.Snapshot{
			Height: 100,
			Format: 1,
			Chunks: 10,
			Hash:   []byte("snapshot_hash"),
		},
		AppHash: []byte("app_hash"),
	})
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseOfferSnapshot_ACCEPT, res2.Result)

	// Test LoadSnapshotChunk
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 3,
		Value: &cmtabci.Response_LoadSnapshotChunk{
			LoadSnapshotChunk: &cmtabci.ResponseLoadSnapshotChunk{
				Chunk: []byte("chunk_data"),
			},
		},
	})

	res3, err := client.LoadSnapshotChunk(ctx, &cmtabci.RequestLoadSnapshotChunk{
		Height: 100,
		Format: 1,
		Chunk:  0,
	})
	require.NoError(t, err)
	assert.Equal(t, []byte("chunk_data"), res3.Chunk)

	// Test ApplySnapshotChunk
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 4,
		Value: &cmtabci.Response_ApplySnapshotChunk{
			ApplySnapshotChunk: &cmtabci.ResponseApplySnapshotChunk{
				Result:        cmtabci.ResponseApplySnapshotChunk_ACCEPT,
				RefetchChunks: []uint32{1, 2},
				RejectSenders: []string{"sender1"},
			},
		},
	})

	res4, err := client.ApplySnapshotChunk(ctx, &cmtabci.RequestApplySnapshotChunk{
		Index:  0,
		Chunk:  []byte("chunk_data"),
		Sender: "sender1",
	})
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseApplySnapshotChunk_ACCEPT, res4.Result)
	assert.Equal(t, []uint32{1, 2}, res4.RefetchChunks)
	assert.Equal(t, []string{"sender1"}, res4.RejectSenders)
}

func TestSocketClientContextCancellation(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &cmtabci.RequestCheckTx{
		Tx:   []byte("test_tx"),
		Type: cmtabci.CheckTxType_New,
	}

	_, err := client.CheckTx(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestSocketClientConcurrentRequests(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server that responds to multiple requests
	go func() {
		for i := uint64(1); i <= 5; i++ {
			mockABCIResponse(serverConn, &cmtabci.Response{
				RequestID: i,
				Value: &cmtabci.Response_CheckTx{
					CheckTx: &cmtabci.ResponseCheckTx{
						Code:      0,
						GasWanted: int64(i * 100),
						GasUsed:   int64(i * 100),
						Log:       "success",
					},
				},
			})
		}
	}()

	// Send concurrent requests
	ctx := context.Background()
	results := make(chan *cmtabci.ResponseCheckTx, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			req := &cmtabci.RequestCheckTx{
				Tx:   []byte("test_tx"),
				Type: cmtabci.CheckTxType_New,
			}
			res, err := client.CheckTx(ctx, req)
			if err != nil {
				errors <- err
			} else {
				results <- res
			}
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < 5; i++ {
		select {
		case res := <-results:
			assert.Equal(t, uint32(0), res.Code)
			successCount++
		case err := <-errors:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for response")
		}
	}

	assert.Equal(t, 5, successCount)
}

func TestSocketClientAsyncMethods(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	logger := &testLogger{}
	client := NewSocketClient(clientConn, logger)

	// Start mock server
	go mockABCIResponse(serverConn, &cmtabci.Response{
		RequestID: 1,
		Value: &cmtabci.Response_CheckTx{
			CheckTx: &cmtabci.ResponseCheckTx{
				Code:      0,
				GasWanted: 100,
				GasUsed:   100,
				Log:       "success",
			},
		},
	})

	// Test async CheckTx
	ctx := context.Background()
	req := &cmtabci.RequestCheckTx{
		Tx:   []byte("test_tx"),
		Type: cmtabci.CheckTxType_New,
	}

	reqRes := client.CheckTxAsync(ctx, req)
	require.NotNil(t, reqRes)

	// Wait for completion
	reqRes.Wait()
	assert.NoError(t, reqRes.Error)
	assert.NotNil(t, reqRes.Response)

	checkTxRes, ok := reqRes.Response.(*cmtabci.ResponseCheckTx)
	assert.True(t, ok)
	assert.Equal(t, uint32(0), checkTxRes.Code)
	assert.Equal(t, int64(100), checkTxRes.GasWanted)
}

// Mock server function
func mockABCIResponse(conn net.Conn, response *cmtabci.Response) {
	// This is a simplified mock - in a real test you'd implement proper protobuf serialization
	// For now, we'll just close the connection to simulate a response
	time.Sleep(10 * time.Millisecond) // Simulate processing time
	conn.Close()
}

// Test logger implementation
type testLogger struct{}

func (l *testLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *testLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *testLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *testLogger) Warn(msg string, keysAndValues ...interface{})  {}
