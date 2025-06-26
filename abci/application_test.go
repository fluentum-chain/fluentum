package abci

import (
	"context"
	"testing"
	"time"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplication(t *testing.T) {
	app := NewApplication()
	require.NotNil(t, app)
	
	// Test initial state
	assert.Equal(t, int64(0), app.GetHeight())
	assert.Equal(t, []byte("initial_hash"), app.GetAppHash())
	assert.Empty(t, app.GetState())
}

func TestApplicationCheckTx(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test valid transaction
	req := &cmtabci.RequestCheckTx{
		Tx:   []byte("key=value"),
		Type: cmtabci.CheckTxType_New,
	}

	res, err := app.CheckTx(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), res.Code)
	assert.Equal(t, int64(80), res.GasWanted) // 8 bytes * 10
	assert.Equal(t, int64(80), res.GasUsed)
	assert.Equal(t, "transaction is valid", res.Log)

	// Test empty transaction
	reqEmpty := &cmtabci.RequestCheckTx{
		Tx:   []byte{},
		Type: cmtabci.CheckTxType_New,
	}

	resEmpty, err := app.CheckTx(ctx, reqEmpty)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), resEmpty.Code)
	assert.Equal(t, "empty transaction", resEmpty.Log)

	// Test invalid transaction format
	reqInvalid := &cmtabci.RequestCheckTx{
		Tx:   []byte("a"), // Too short
		Type: cmtabci.CheckTxType_New,
	}

	resInvalid, err := app.CheckTx(ctx, reqInvalid)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), resInvalid.Code)
	assert.Equal(t, "invalid transaction format", resInvalid.Log)
}

func TestApplicationFinalizeBlock(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test with valid transactions
	req := &cmtabci.RequestFinalizeBlock{
		Txs: [][]byte{
			[]byte("a=value1"),
			[]byte("b=value2"),
		},
		Height: 1,
	}

	res, err := app.FinalizeBlock(ctx, req)
	require.NoError(t, err)
	assert.Len(t, res.TxResults, 2)
	assert.Equal(t, uint32(0), res.TxResults[0].Code)
	assert.Equal(t, uint32(0), res.TxResults[1].Code)
	assert.Equal(t, "transaction executed successfully", res.TxResults[0].Log)
	assert.Equal(t, "transaction executed successfully", res.TxResults[1].Log)
	assert.NotNil(t, res.AppHash)
	assert.NotNil(t, res.ConsensusParamUpdates)

	// Verify state was updated
	state := app.GetState()
	assert.Equal(t, []byte("value1"), state["a"])
	assert.Equal(t, []byte("value2"), state["b"])
	assert.Equal(t, int64(1), app.GetHeight())

	// Test with invalid transaction
	reqInvalid := &cmtabci.RequestFinalizeBlock{
		Txs: [][]byte{
			[]byte("a=value3"),
			[]byte(""), // Invalid empty transaction
		},
		Height: 2,
	}

	resInvalid, err := app.FinalizeBlock(ctx, reqInvalid)
	require.NoError(t, err)
	assert.Len(t, resInvalid.TxResults, 2)
	assert.Equal(t, uint32(0), resInvalid.TxResults[0].Code) // First tx should succeed
	assert.Equal(t, uint32(1), resInvalid.TxResults[1].Code) // Second tx should fail
	assert.Equal(t, int64(2), app.GetHeight())
}

func TestApplicationCommit(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Set some state first
	app.mtx.Lock()
	app.state["test"] = []byte("value")
	app.appHash = []byte("test_hash")
	app.mtx.Unlock()

	req := &cmtabci.RequestCommit{}
	res, err := app.Commit(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, []byte("test_hash"), res.Data)
}

func TestApplicationInfo(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Set some state
	app.mtx.Lock()
	app.height = 100
	app.appHash = []byte("last_hash")
	app.mtx.Unlock()

	req := &cmtabci.RequestInfo{
		Version: "1.0.0",
	}

	res, err := app.Info(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "tendermint-abci-app", res.Data)
	assert.Equal(t, "1.0.0", res.Version)
	assert.Equal(t, uint64(1), res.AppVersion)
	assert.Equal(t, int64(100), res.LastBlockHeight)
	assert.Equal(t, []byte("last_hash"), res.LastBlockAppHash)
}

func TestApplicationQuery(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Set some state
	app.mtx.Lock()
	app.state["test_key"] = []byte("test_value")
	app.height = 50
	app.mtx.Unlock()

	// Test store query
	reqStore := &cmtabci.RequestQuery{
		Data: []byte("test_key"),
		Path: "/store",
	}

	resStore, err := app.Query(ctx, reqStore)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), resStore.Code)
	assert.Equal(t, []byte("test_value"), resStore.Value)
	assert.Equal(t, "found", resStore.Log)

	// Test height query
	reqHeight := &cmtabci.RequestQuery{
		Path: "/height",
	}

	resHeight, err := app.Query(ctx, reqHeight)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), resHeight.Code)
	assert.Equal(t, []byte("50"), resHeight.Value)
	assert.Equal(t, "current height", resHeight.Log)

	// Test unknown path
	reqUnknown := &cmtabci.RequestQuery{
		Path: "/unknown",
	}

	resUnknown, err := app.Query(ctx, reqUnknown)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), resUnknown.Code)
	assert.Equal(t, "unknown query path", resUnknown.Log)

	// Test store query for non-existent key
	reqNotFound := &cmtabci.RequestQuery{
		Data: []byte("non_existent"),
		Path: "/store",
	}

	resNotFound, err := app.Query(ctx, reqNotFound)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), resNotFound.Code)
	assert.Equal(t, "not found", resNotFound.Log)
}

func TestApplicationInitChain(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	req := &cmtabci.RequestInitChain{
		Time:    time.Now(),
		ChainId: "test_chain",
		AppStateBytes: []byte("genesis_state"),
		Validators: []cmtabci.ValidatorUpdate{
			{
				PubKey: cmtabci.PubKey{Data: []byte("validator1")},
				Power:  100,
			},
			{
				PubKey: cmtabci.PubKey{Data: []byte("validator2")},
				Power:  200,
			},
		},
		InitialHeight: 1,
	}

	res, err := app.InitChain(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, res.ConsensusParams)
	assert.Len(t, res.Validators, 2)
	assert.Equal(t, int64(100), res.Validators[0].Power)
	assert.Equal(t, int64(200), res.Validators[1].Power)
	assert.NotNil(t, res.AppHash)

	// Verify genesis state was stored
	state := app.GetState()
	assert.Equal(t, []byte("genesis_state"), state["genesis"])
}

func TestApplicationPrepareProposal(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	req := &cmtabci.RequestPrepareProposal{
		MaxTxBytes: 100,
		Txs: [][]byte{
			[]byte("a=value1"), // 8 bytes
			[]byte("b=value2"), // 8 bytes
			[]byte("c=value3"), // 8 bytes
			[]byte("d=value4"), // 8 bytes
		},
		Height: 1,
	}

	res, err := app.PrepareProposal(ctx, req)
	require.NoError(t, err)
	
	// Should select transactions that fit within MaxTxBytes
	// Each transaction is 8 bytes, so we can fit 12 transactions (96 bytes)
	// But we only have 4 transactions, so all should be selected
	assert.Len(t, res.Txs, 4)
	assert.Equal(t, []byte("a=value1"), res.Txs[0])
	assert.Equal(t, []byte("b=value2"), res.Txs[1])
	assert.Equal(t, []byte("c=value3"), res.Txs[2])
	assert.Equal(t, []byte("d=value4"), res.Txs[3])

	// Test with transactions that exceed MaxTxBytes
	reqLarge := &cmtabci.RequestPrepareProposal{
		MaxTxBytes: 10, // Very small limit
		Txs: [][]byte{
			[]byte("a=value1"), // 8 bytes
			[]byte("b=value2"), // 8 bytes
		},
		Height: 1,
	}

	resLarge, err := app.PrepareProposal(ctx, reqLarge)
	require.NoError(t, err)
	
	// Should only select one transaction that fits
	assert.Len(t, resLarge.Txs, 1)
	assert.Equal(t, []byte("a=value1"), resLarge.Txs[0])
}

func TestApplicationProcessProposal(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test with valid transactions
	reqValid := &cmtabci.RequestProcessProposal{
		Txs: [][]byte{
			[]byte("a=value1"),
			[]byte("b=value2"),
		},
		Height: 1,
	}

	resValid, err := app.ProcessProposal(ctx, reqValid)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseProcessProposal_ACCEPT, resValid.Status)

	// Test with invalid transaction
	reqInvalid := &cmtabci.RequestProcessProposal{
		Txs: [][]byte{
			[]byte("a=value1"),
			[]byte(""), // Invalid empty transaction
		},
		Height: 1,
	}

	resInvalid, err := app.ProcessProposal(ctx, reqInvalid)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseProcessProposal_REJECT, resInvalid.Status)
}

func TestApplicationExtendVote(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	req := &cmtabci.RequestExtendVote{
		Hash:   []byte("block_hash"),
		Height: 100,
	}

	res, err := app.ExtendVote(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, []byte("height_100"), res.VoteExtension)
}

func TestApplicationVerifyVoteExtension(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test valid vote extension
	reqValid := &cmtabci.RequestVerifyVoteExtension{
		Hash:              []byte("block_hash"),
		ValidatorProTxHash: []byte("validator_hash"),
		Height:            100,
		VoteExtension:     []byte("height_100"),
	}

	resValid, err := app.VerifyVoteExtension(ctx, reqValid)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseVerifyVoteExtension_ACCEPT, resValid.Status)

	// Test invalid vote extension
	reqInvalid := &cmtabci.RequestVerifyVoteExtension{
		Hash:              []byte("block_hash"),
		ValidatorProTxHash: []byte("validator_hash"),
		Height:            100,
		VoteExtension:     []byte("invalid_extension"),
	}

	resInvalid, err := app.VerifyVoteExtension(ctx, reqInvalid)
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseVerifyVoteExtension_REJECT, resInvalid.Status)
}

func TestApplicationSnapshotMethods(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test ListSnapshots
	resList, err := app.ListSnapshots(ctx, &cmtabci.RequestListSnapshots{})
	require.NoError(t, err)
	assert.Empty(t, resList.Snapshots)

	// Test OfferSnapshot
	resOffer, err := app.OfferSnapshot(ctx, &cmtabci.RequestOfferSnapshot{
		Snapshot: &cmtabci.Snapshot{
			Height: 100,
			Format: 1,
			Chunks: 10,
			Hash:   []byte("snapshot_hash"),
		},
		AppHash: []byte("app_hash"),
	})
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseOfferSnapshot_REJECT, resOffer.Result)

	// Test LoadSnapshotChunk
	resLoad, err := app.LoadSnapshotChunk(ctx, &cmtabci.RequestLoadSnapshotChunk{
		Height: 100,
		Format: 1,
		Chunk:  0,
	})
	require.NoError(t, err)
	assert.Empty(t, resLoad.Chunk)

	// Test ApplySnapshotChunk
	resApply, err := app.ApplySnapshotChunk(ctx, &cmtabci.RequestApplySnapshotChunk{
		Index:  0,
		Chunk:  []byte("chunk_data"),
		Sender: "sender1",
	})
	require.NoError(t, err)
	assert.Equal(t, cmtabci.ResponseApplySnapshotChunk_REJECT, resApply.Result)
}

func TestApplicationExecuteTransaction(t *testing.T) {
	app := NewApplication()

	// Test valid transaction
	tx := []byte("key=value")
	res, err := app.executeTransaction(tx)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), res.Code)
	assert.Equal(t, []byte("stored: key=value"), res.Data)
	assert.Equal(t, "transaction executed successfully", res.Log)
	assert.Len(t, res.Events, 1)
	assert.Equal(t, "store", res.Events[0].Type)
	assert.Len(t, res.Events[0].Attributes, 2)

	// Test transaction without equals sign
	txNoEquals := []byte("simple_key")
	res2, err := app.executeTransaction(txNoEquals)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), res2.Code)
	assert.Equal(t, []byte("stored: simple_key=default"), res2.Data)

	// Test empty transaction
	txEmpty := []byte{}
	_, err = app.executeTransaction(txEmpty)
	assert.Error(t, err)
	assert.Equal(t, "empty transaction", err.Error())

	// Verify state was updated
	state := app.GetState()
	assert.Equal(t, []byte("value"), state["key"])
	assert.Equal(t, []byte("default"), state["simple_key"])
}

func TestApplicationCalculateAppHash(t *testing.T) {
	app := NewApplication()

	// Test initial hash
	hash1 := app.calculateAppHash()
	assert.Equal(t, []byte("hash_0_0"), hash1)

	// Update state and height
	app.mtx.Lock()
	app.state["test"] = []byte("value")
	app.height = 1
	app.mtx.Unlock()

	hash2 := app.calculateAppHash()
	assert.Equal(t, []byte("hash_1_1"), hash2)
}

func TestApplicationConcurrentAccess(t *testing.T) {
	app := NewApplication()
	ctx := context.Background()

	// Test concurrent CheckTx calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			req := &cmtabci.RequestCheckTx{
				Tx:   []byte(fmt.Sprintf("key%d=value%d", id, id)),
				Type: cmtabci.CheckTxType_New,
			}
			_, err := app.CheckTx(ctx, req)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent FinalizeBlock calls
	done2 := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			req := &cmtabci.RequestFinalizeBlock{
				Txs: [][]byte{
					[]byte(fmt.Sprintf("tx%d=value%d", id, id)),
				},
				Height: int64(id + 1),
			}
			_, err := app.FinalizeBlock(ctx, req)
			assert.NoError(t, err)
			done2 <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done2
	}
}

func TestApplicationStateManagement(t *testing.T) {
	app := NewApplication()

	// Test GetState
	state := app.GetState()
	assert.Empty(t, state)

	// Test GetHeight
	assert.Equal(t, int64(0), app.GetHeight())

	// Test GetAppHash
	assert.Equal(t, []byte("initial_hash"), app.GetAppHash())

	// Update state and verify
	app.mtx.Lock()
	app.state["test"] = []byte("value")
	app.height = 100
	app.appHash = []byte("new_hash")
	app.mtx.Unlock()

	state = app.GetState()
	assert.Equal(t, []byte("value"), state["test"])
	assert.Equal(t, int64(100), app.GetHeight())
	assert.Equal(t, []byte("new_hash"), app.GetAppHash())
}

// Benchmark tests

func BenchmarkCheckTx(b *testing.B) {
	app := NewApplication()
	ctx := context.Background()
	req := &cmtabci.RequestCheckTx{
		Tx:   []byte("key=value"),
		Type: cmtabci.CheckTxType_New,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.CheckTx(ctx, req)
	}
}

func BenchmarkFinalizeBlock(b *testing.B) {
	app := NewApplication()
	ctx := context.Background()
	req := &cmtabci.RequestFinalizeBlock{
		Txs: [][]byte{
			[]byte("a=value1"),
			[]byte("b=value2"),
			[]byte("c=value3"),
		},
		Height: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.FinalizeBlock(ctx, req)
	}
}

func BenchmarkQuery(b *testing.B) {
	app := NewApplication()
	ctx := context.Background()

	// Set up some state
	app.mtx.Lock()
	app.state["test_key"] = []byte("test_value")
	app.mtx.Unlock()

	req := &cmtabci.RequestQuery{
		Data: []byte("test_key"),
		Path: "/store",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Query(ctx, req)
	}
} 