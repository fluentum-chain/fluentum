package abci

import (
	"context"
	"testing"

	"github.com/fluentum-chain/fluentum/abci/types"
	"github.com/stretchr/testify/require"
)

// Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	var _ types.Application = (*MyApp)(nil)
}

func TestNewMyApp(t *testing.T) {
	app := NewMyApp("test-chain")
	require.NotNil(t, app)
	require.Equal(t, "test-chain", app.chainID)
	require.Equal(t, int64(0), app.height)
	require.NotNil(t, app.state)
	require.NotNil(t, app.gasMeter)
}

func TestMyApp_Info(t *testing.T) {
	app := NewMyApp("test-chain")
	
	res, err := app.Info(context.Background(), &types.RequestInfo{})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Contains(t, res.Data, "MyApp v1.0.0")
	require.Contains(t, res.Data, "test-chain")
	require.Equal(t, "1.0.0", res.Version)
	require.Equal(t, uint64(1), res.AppVersion)
	require.Equal(t, int64(0), res.LastBlockHeight)
}

func TestMyApp_CheckTx(t *testing.T) {
	app := NewMyApp("test-chain")
	
	// Test valid transaction
	validTx := []byte("SET key=value")
	res, err := app.CheckTx(context.Background(), &types.RequestCheckTx{
		Tx:   validTx,
		Type: types.CheckTxType_New,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, types.CodeTypeOK, res.Code)
	require.Equal(t, []byte("valid"), res.Data)
	require.Equal(t, int64(len(validTx)), res.GasWanted)
	
	// Test empty transaction
	emptyRes, err := app.CheckTx(context.Background(), &types.RequestCheckTx{
		Tx:   []byte{},
		Type: types.CheckTxType_New,
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeEncodingError, emptyRes.Code)
	require.Contains(t, emptyRes.Log, "empty transaction")
	
	// Test short transaction
	shortRes, err := app.CheckTx(context.Background(), &types.RequestCheckTx{
		Tx:   []byte("abc"),
		Type: types.CheckTxType_New,
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeEncodingError, shortRes.Code)
	require.Contains(t, shortRes.Log, "transaction too short")
}

func TestMyApp_FinalizeBlock(t *testing.T) {
	app := NewMyApp("test-chain")
	
	// Test empty block
	res, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.TxResults, 0)
	require.Equal(t, int64(1), app.height)
	require.Len(t, res.Events, 1) // block event
	
	// Test block with transactions
	txs := [][]byte{
		[]byte("SET key1=value1"),
		[]byte("SET key2=value2"),
		[]byte("GET key1"),
	}
	
	res2, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 2,
		Txs:    txs,
	})
	require.NoError(t, err)
	require.NotNil(t, res2)
	require.Len(t, res2.TxResults, 3)
	require.Equal(t, int64(2), app.height)
	require.Len(t, res2.Events, 4) // 3 tx events + 1 block event
	
	// Check that transactions were processed correctly
	require.Equal(t, types.CodeTypeOK, res2.TxResults[0].Code) // SET key1=value1
	require.Equal(t, types.CodeTypeOK, res2.TxResults[1].Code) // SET key2=value2
	require.Equal(t, types.CodeTypeOK, res2.TxResults[2].Code) // GET key1
}

func TestMyApp_Commit(t *testing.T) {
	app := NewMyApp("test-chain")
	app.height = 100
	
	res, err := app.Commit(context.Background(), &types.RequestCommit{})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Data) // app hash
	require.Equal(t, int64(100), res.RetainHeight)
}

func TestMyApp_InitChain(t *testing.T) {
	app := NewMyApp("test-chain")
	
	req := &types.RequestInitChain{
		ChainId:       "new-chain",
		InitialHeight: 1,
		AppStateBytes: []byte("genesis state"),
	}
	
	res, err := app.InitChain(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "new-chain", app.chainID)
	require.Equal(t, int64(1), app.height)
	require.Equal(t, []byte("genesis state"), app.state["genesis"])
}

func TestMyApp_Query(t *testing.T) {
	app := NewMyApp("test-chain")
	app.state["test_key"] = []byte("test_value")
	
	// Test successful query
	res, err := app.Query(context.Background(), &types.RequestQuery{
		Path: "state",
		Data: []byte("test_key"),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, types.CodeTypeOK, res.Code)
	require.Equal(t, []byte("test_value"), res.Value)
	
	// Test query for non-existent key
	res2, err := app.Query(context.Background(), &types.RequestQuery{
		Path: "state",
		Data: []byte("non_existent_key"),
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeUnknownAddress, res2.Code)
	require.Contains(t, res2.Log, "key not found")
	
	// Test unknown query path
	res3, err := app.Query(context.Background(), &types.RequestQuery{
		Path: "unknown",
		Data: []byte("test"),
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeUnknownRequest, res3.Code)
	require.Contains(t, res3.Log, "unknown query path")
}

func TestMyApp_TransactionProcessing(t *testing.T) {
	app := NewMyApp("test-chain")
	
	// Test SET transaction
	setTx := []byte("SET mykey=myvalue")
	res, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 1,
		Txs:    [][]byte{setTx},
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeOK, res.TxResults[0].Code)
	require.Equal(t, []byte("set"), res.TxResults[0].Data)
	
	// Test GET transaction
	getTx := []byte("GET mykey")
	res2, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 2,
		Txs:    [][]byte{getTx},
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeOK, res2.TxResults[0].Code)
	require.Equal(t, []byte("myvalue"), res2.TxResults[0].Data)
	
	// Test GET for non-existent key
	getNonExistentTx := []byte("GET nonexistent")
	res3, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 3,
		Txs:    [][]byte{getNonExistentTx},
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeUnknownAddress, res3.TxResults[0].Code)
	
	// Test invalid SET format
	invalidSetTx := []byte("SET invalid_format")
	res4, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 4,
		Txs:    [][]byte{invalidSetTx},
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeEncodingError, res4.TxResults[0].Code)
	
	// Test unknown command
	unknownTx := []byte("UNKN command")
	res5, err := app.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{
		Height: 5,
		Txs:    [][]byte{unknownTx},
	})
	require.NoError(t, err)
	require.Equal(t, types.CodeTypeUnknownRequest, res5.TxResults[0].Code)
}

func TestSimpleGasMeter(t *testing.T) {
	gm := NewSimpleGasMeter(1000)
	
	// Test initial state
	require.Equal(t, int64(0), gm.GasConsumed())
	require.Equal(t, int64(1000), gm.GasLimit())
	require.False(t, gm.IsOutOfGas())
	
	// Test consuming gas
	err := gm.ConsumeGas(500, "test")
	require.NoError(t, err)
	require.Equal(t, int64(500), gm.GasConsumed())
	require.False(t, gm.IsOutOfGas())
	
	// Test consuming more gas
	err = gm.ConsumeGas(400, "test2")
	require.NoError(t, err)
	require.Equal(t, int64(900), gm.GasConsumed())
	require.False(t, gm.IsOutOfGas())
	
	// Test out of gas
	err = gm.ConsumeGas(200, "test3")
	require.Error(t, err)
	require.Contains(t, err.Error(), "out of gas")
	require.Equal(t, int64(900), gm.GasConsumed()) // Should not change
	
	// Test refunding gas
	gm.RefundGas(100, "refund")
	require.Equal(t, int64(800), gm.GasConsumed())
	
	// Test reset
	gm.Reset()
	require.Equal(t, int64(0), gm.GasConsumed())
} 