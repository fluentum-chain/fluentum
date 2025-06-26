package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckTxType(t *testing.T) {
	// Test CheckTxType constants
	require.Equal(t, CheckTxType(0), CheckTxType_New)
	require.Equal(t, CheckTxType(1), CheckTxType_Recheck)

	// Test String method
	require.Equal(t, "NEW", CheckTxType_New.String())
	require.Equal(t, "RECHECK", CheckTxType_Recheck.String())
	require.Equal(t, "UNKNOWN", CheckTxType(99).String())
}

func TestRequestCheckTx(t *testing.T) {
	req := &RequestCheckTx{
		Tx:   []byte("test_tx"),
		Type: CheckTxType_New,
	}

	require.Equal(t, []byte("test_tx"), req.Tx)
	require.Equal(t, CheckTxType_New, req.Type)
}

func TestResponseCheckTx(t *testing.T) {
	events := []Event{
		{
			Type: "test_event",
			Attributes: []EventAttribute{
				{Key: "key1", Value: "value1", Index: true},
				{Key: "key2", Value: "value2", Index: false},
			},
		},
	}

	res := &ResponseCheckTx{
		Code:      0,
		Data:      []byte("success"),
		Log:       "transaction processed",
		Info:      "info",
		GasWanted: 1000,
		GasUsed:   500,
		Events:    events,
		Codespace: "test",
	}

	require.Equal(t, uint32(0), res.Code)
	require.Equal(t, []byte("success"), res.Data)
	require.Equal(t, "transaction processed", res.Log)
	require.Equal(t, "info", res.Info)
	require.Equal(t, int64(1000), res.GasWanted)
	require.Equal(t, int64(500), res.GasUsed)
	require.Len(t, res.Events, 1)
	require.Equal(t, "test", res.Codespace)
}

func TestRequestFinalizeBlock(t *testing.T) {
	header := &Header{
		ChainID: "test-chain",
		Height:  100,
		Time:    time.Now(),
	}

	req := &RequestFinalizeBlock{
		Height: 100,
		Txs:    [][]byte{[]byte("tx1"), []byte("tx2")},
		Hash:   []byte("block_hash"),
		Header: header,
	}

	require.Equal(t, int64(100), req.Height)
	require.Len(t, req.Txs, 2)
	require.Equal(t, []byte("block_hash"), req.Hash)
	require.Equal(t, header, req.Header)
}

func TestResponseFinalizeBlock(t *testing.T) {
	txResults := []*ExecTxResult{
		{
			Code:      0,
			Data:      []byte("result1"),
			Log:       "success",
			GasUsed:   100,
			GasWanted: 200,
		},
		{
			Code:      1,
			Data:      []byte("result2"),
			Log:       "error",
			GasUsed:   50,
			GasWanted: 100,
		},
	}

	validatorUpdates := []ValidatorUpdate{
		{
			PubKey: PubKey{Type: "ed25519", Data: []byte("pubkey1")},
			Power:  100,
		},
	}

	events := []Event{
		{
			Type: "finalize_block",
			Attributes: []EventAttribute{
				{Key: "height", Value: "100", Index: true},
			},
		},
	}

	res := &ResponseFinalizeBlock{
		TxResults:        txResults,
		ValidatorUpdates: validatorUpdates,
		AppHash:          []byte("app_hash"),
		Events:           events,
	}

	require.Len(t, res.TxResults, 2)
	require.Len(t, res.ValidatorUpdates, 1)
	require.Equal(t, []byte("app_hash"), res.AppHash)
	require.Len(t, res.Events, 1)
}

func TestExecTxResult(t *testing.T) {
	events := []Event{
		{
			Type: "exec_tx",
			Attributes: []EventAttribute{
				{Key: "status", Value: "success", Index: true},
			},
		},
	}

	result := &ExecTxResult{
		Code:      0,
		Data:      []byte("execution_result"),
		Log:       "execution successful",
		Info:      "info",
		Events:    events,
		GasUsed:   150,
		GasWanted: 300,
	}

	require.Equal(t, uint32(0), result.Code)
	require.Equal(t, []byte("execution_result"), result.Data)
	require.Equal(t, "execution successful", result.Log)
	require.Equal(t, "info", result.Info)
	require.Len(t, result.Events, 1)
	require.Equal(t, int64(150), result.GasUsed)
	require.Equal(t, int64(300), result.GasWanted)
}

func TestEventAndEventAttribute(t *testing.T) {
	attr1 := EventAttribute{
		Key:   "module",
		Value: "bank",
		Index: true,
	}

	attr2 := EventAttribute{
		Key:   "action",
		Value: "transfer",
		Index: false,
	}

	event := Event{
		Type:       "transfer",
		Attributes: []EventAttribute{attr1, attr2},
	}

	require.Equal(t, "transfer", event.Type)
	require.Len(t, event.Attributes, 2)
	require.Equal(t, "module=bank", attr1.String())
	require.Equal(t, "transfer(module=bank, action=transfer)", event.String())
}

func TestValidatorUpdate(t *testing.T) {
	validator := ValidatorUpdate{
		PubKey: PubKey{
			Type: "ed25519",
			Data: []byte("validator_pubkey"),
		},
		Power: 1000,
	}

	require.Equal(t, "ed25519", validator.PubKey.Type)
	require.Equal(t, []byte("validator_pubkey"), validator.PubKey.Data)
	require.Equal(t, int64(1000), validator.Power)
}

func TestSnapshot(t *testing.T) {
	snapshot := &Snapshot{
		Height:   1000,
		Format:   1,
		Chunks:   10,
		Hash:     []byte("snapshot_hash"),
		Metadata: []byte("metadata"),
	}

	require.Equal(t, uint64(1000), snapshot.Height)
	require.Equal(t, uint32(1), snapshot.Format)
	require.Equal(t, uint32(10), snapshot.Chunks)
	require.Equal(t, []byte("snapshot_hash"), snapshot.Hash)
	require.Equal(t, []byte("metadata"), snapshot.Metadata)
}

func TestResponseOfferSnapshotResult(t *testing.T) {
	require.Equal(t, ResponseOfferSnapshot_Result(0), ResponseOfferSnapshot_UNKNOWN)
	require.Equal(t, ResponseOfferSnapshot_Result(1), ResponseOfferSnapshot_ACCEPT)
	require.Equal(t, ResponseOfferSnapshot_Result(2), ResponseOfferSnapshot_ABORT)
	require.Equal(t, ResponseOfferSnapshot_Result(3), ResponseOfferSnapshot_REJECT)
	require.Equal(t, ResponseOfferSnapshot_Result(4), ResponseOfferSnapshot_REJECT_FORMAT)
	require.Equal(t, ResponseOfferSnapshot_Result(5), ResponseOfferSnapshot_REJECT_SENDER)
}

func TestResponseApplySnapshotChunkResult(t *testing.T) {
	require.Equal(t, ResponseApplySnapshotChunk_Result(0), ResponseApplySnapshotChunk_UNKNOWN)
	require.Equal(t, ResponseApplySnapshotChunk_Result(1), ResponseApplySnapshotChunk_ACCEPT)
	require.Equal(t, ResponseApplySnapshotChunk_Result(2), ResponseApplySnapshotChunk_ABORT)
	require.Equal(t, ResponseApplySnapshotChunk_Result(3), ResponseApplySnapshotChunk_RETRY)
	require.Equal(t, ResponseApplySnapshotChunk_Result(4), ResponseApplySnapshotChunk_RETRY_SNAPSHOT)
	require.Equal(t, ResponseApplySnapshotChunk_Result(5), ResponseApplySnapshotChunk_REJECT_SENDER)
} 