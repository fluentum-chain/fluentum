package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fluentum-chain/fluentum/quantum"
	"github.com/stretchr/testify/require"

	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/consensus"
	"github.com/fluentum-chain/fluentum/types"
	"github.com/fluentum-chain/fluentum/zkprover"
)

// TestHybridConsensus tests the hybrid consensus mechanism
func TestHybridConsensus(t *testing.T) {
	// Setup testnet
	cfg := ResetConfig("hybrid_consensus_test")
	defer os.RemoveAll(cfg.RootDir)

	// Initialize consensus state
	consensusState, err := consensus.NewState(cfg)
	require.NoError(t, err)
	defer consensusState.Stop()

	// Create and validate ZK batch
	batch := createValidZKBatch(t)
	err = consensusState.HandleZKBatch(context.Background(), batch)
	require.NoError(t, err)

	// Create test block with ZK batch
	block := createTestBlock(t, batch)
	require.NotNil(t, block)

	// Process block
	err = consensusState.FinalizeBlock(context.Background(), block)
	require.NoError(t, err)

	// Verify state changes
	newState := getAppState(t, consensusState)
	require.Equal(t, batch.ExpectedStateRoot, newState)

	// Verify block finalization
	require.True(t, consensusState.IsBlockFinalized(block.Height))
}

// TestQuantumSignatures tests quantum-resistant signatures
func TestQuantumSignatures(t *testing.T) {
	// Generate key pair
	privKey, pubKey, err := quantum.GenerateKeyPair()
	require.NoError(t, err)

	// Test message
	msg := []byte("test message")

	// Sign message
	sig, err := quantum.Sign(privKey, msg)
	require.NoError(t, err)
	require.NotNil(t, sig)

	// Verify valid signature
	valid, err := quantum.Verify(pubKey, msg, sig)
	require.NoError(t, err)
	require.True(t, valid)

	// Test tampered message
	tamperedMsg := []byte("tampered message")
	valid, err = quantum.Verify(pubKey, tamperedMsg, sig)
	require.NoError(t, err)
	require.False(t, valid)

	// Test invalid signature
	invalidSig := make([]byte, len(sig))
	copy(invalidSig, sig)
	invalidSig[0] ^= 0xFF // Flip some bits
	valid, err = quantum.Verify(pubKey, msg, invalidSig)
	require.NoError(t, err)
	require.False(t, valid)
}

// TestConsensusStateTransition tests state transitions in consensus
func TestConsensusStateTransition(t *testing.T) {
	cfg := ResetConfig("consensus_state_test")
	defer os.RemoveAll(cfg.RootDir)

	consensusState, err := consensus.NewState(cfg)
	require.NoError(t, err)
	defer consensusState.Stop()

	// Test initial state
	require.Equal(t, consensus.RoundStepNewHeight, consensusState.Step())

	// Create and process ZK batch
	batch := createValidZKBatch(t)
	err = consensusState.HandleZKBatch(context.Background(), batch)
	require.NoError(t, err)

	// Verify state transition
	require.Equal(t, consensus.RoundStepPropose, consensusState.Step())

	// Create and process block
	block := createTestBlock(t, batch)
	err = consensusState.FinalizeBlock(context.Background(), block)
	require.NoError(t, err)

	// Verify final state
	require.Equal(t, consensus.RoundStepCommit, consensusState.Step())
}

// Helper functions

func createValidZKBatch(t *testing.T) *types.ZKBatch {
	// Create test transactions
	txs := []types.Tx{
		{
			Type:     types.TxTypeTransfer,
			From:     "0x123",
			To:       "0x456",
			Amount:   1000000000,
			Nonce:    1,
			Gas:      21000,
			GasPrice: 20000000000,
		},
	}

	// Generate ZK proof
	prover := zkprover.NewProver()
	proof, err := prover.ProveBatch(txs)
	require.NoError(t, err)

	return &types.ZKBatch{
		Transactions: txs,
		Proof:        proof,
		StateRoot:    []byte("test_state_root"),
		Timestamp:    time.Now().Unix(),
	}
}

func createTestBlock(t *testing.T, batch *types.ZKBatch) *types.Block {
	return &types.Block{
		Header: types.Header{
			Height:    1,
			Time:      time.Now(),
			ChainID:   "test_chain",
			StateRoot: batch.StateRoot,
		},
		Data: types.Data{
			Txs: batch.Transactions,
		},
		Evidence: types.EvidenceData{},
		LastCommit: &types.Commit{
			Height:     0,
			Round:      0,
			BlockID:    types.BlockID{},
			Signatures: []types.CommitSig{},
		},
	}
}

func getAppState(t *testing.T, consensusState *consensus.State) []byte {
	state, err := consensusState.GetAppState()
	require.NoError(t, err)
	return state
}

func ResetConfig(testName string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.RootDir = os.TempDir() + "/" + testName
	cfg.P2P.ListenAddress = "tcp://127.0.0.1:0"
	cfg.RPC.ListenAddress = "tcp://127.0.0.1:0"
	cfg.Consensus.CreateEmptyBlocksInterval = 0
	return cfg
}
