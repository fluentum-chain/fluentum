package consensus

import (
	"errors"
	"fmt"
	"time"

	"github.com/fluentum-chain/fluentum/consensus"
	"github.com/fluentum-chain/fluentum/types"
	"github.com/fluentum-chain/fluentum/state"
	"github.com/fluentum-chain/fluentum/libs/log"
	"github.com/fluentum-chain/fluentum/libs/service"
	"github.com/fluentum-chain/fluentum/p2p"
	"github.com/fluentum-chain/fluentum/privval"
	"github.com/fluentum-chain/fluentum/proxy"
	"github.com/fluentum-chain/fluentum/fluentum/zkproofs"
	"github.com/fluentum-chain/fluentum/fluentum/quantum"
)

// HybridConsensus combines Tendermint's DPoS with ZK-Rollups and quantum-resistant signatures
type HybridConsensus struct {
	*consensus.State
	service.BaseService
	logger log.Logger

	// Core components
	blockExec    *state.BlockExecutor
	privValidator types.PrivValidator
	blockStore   *store.BlockStore
	stateStore   state.Store

	// ZK-Rollup components
	zkProver     *zkproofs.Prover
	zkVerifier   *zkproofs.Verifier
	zkState      *zkproofs.State

	// Quantum-resistant components
	quantumSigner   *quantum.DilithiumSigner
	quantumVerifier *quantum.DilithiumVerifier

	// Configuration
	blockTime time.Duration
	config    *Config
}

// Config holds the configuration for the hybrid consensus
type Config struct {
	BlockTime           time.Duration
	ZKEnabled          bool
	QuantumEnabled     bool
	ZKProverURL        string
	QuantumKeyFile     string
	MaxZKBatchSize     int
	ZKProofTimeout     time.Duration
}

// NewHybridConsensus creates a new hybrid consensus instance
func NewHybridConsensus(
	config *Config,
	blockExec *state.BlockExecutor,
	blockStore *store.BlockStore,
	stateStore state.Store,
	privValidator types.PrivValidator,
	logger log.Logger,
) *HybridConsensus {
	hc := &HybridConsensus{
		State:         consensus.NewState(config, blockExec, blockStore, stateStore, privValidator, logger),
		logger:        logger,
		blockExec:     blockExec,
		blockStore:    blockStore,
		stateStore:    stateStore,
		privValidator: privValidator,
		blockTime:     config.BlockTime,
		config:        config,
	}

	// Initialize ZK components if enabled
	if config.ZKEnabled {
		hc.zkProver = zkproofs.NewProver(config.ZKProverURL)
		hc.zkVerifier = zkproofs.NewVerifier()
		hc.zkState = zkproofs.NewState()
	}

	// Initialize quantum components if enabled
	if config.QuantumEnabled {
		hc.quantumSigner = quantum.NewDilithiumSigner(config.QuantumKeyFile)
		hc.quantumVerifier = quantum.NewDilithiumVerifier()
	}

	hc.BaseService = *service.NewBaseService(logger, "HybridConsensus", hc)
	return hc
}

// FinalizeBlock processes a block through the hybrid consensus mechanism
func (hc *HybridConsensus) FinalizeBlock(block *types.Block) error {
	// 1. Process regular transactions with Tendermint
	if err := hc.State.FinalizeBlock(block); err != nil {
		return fmt.Errorf("tendermint finalization failed: %w", err)
	}

	// 2. Process ZK-Rollup batches if enabled
	if hc.config.ZKEnabled {
		for _, batch := range block.Data.ZKBatches {
			// Verify ZK proof
			if err := hc.verifyZKBatch(batch); err != nil {
				return fmt.Errorf("zk batch verification failed: %w", err)
			}

			// Apply state transition
			if err := hc.applyZKStateTransition(batch); err != nil {
				return fmt.Errorf("zk state transition failed: %w", err)
			}
		}
	}

	// 3. Verify quantum signature if enabled
	if hc.config.QuantumEnabled {
		if err := hc.verifyQuantumSignature(block); err != nil {
			return fmt.Errorf("quantum signature verification failed: %w", err)
		}
	}

	return nil
}

// verifyZKBatch verifies a ZK-Rollup batch
func (hc *HybridConsensus) verifyZKBatch(batch *types.ZKBatch) error {
	// Verify proof
	valid, err := hc.zkVerifier.VerifyProof(batch.Proof)
	if err != nil {
		return fmt.Errorf("zk proof verification error: %w", err)
	}
	if !valid {
		return errors.New("invalid zk proof")
	}

	// Verify batch size
	if len(batch.Transactions) > hc.config.MaxZKBatchSize {
		return fmt.Errorf("zk batch size exceeds maximum: %d > %d", 
			len(batch.Transactions), hc.config.MaxZKBatchSize)
	}

	return nil
}

// applyZKStateTransition applies the state transition from a ZK batch
func (hc *HybridConsensus) applyZKStateTransition(batch *types.ZKBatch) error {
	// Apply state transition
	if err := hc.zkState.ApplyTransition(batch.StateTransition); err != nil {
		return fmt.Errorf("failed to apply zk state transition: %w", err)
	}

	// Update state hash
	if err := hc.zkState.UpdateHash(); err != nil {
		return fmt.Errorf("failed to update zk state hash: %w", err)
	}

	return nil
}

// verifyQuantumSignature verifies the quantum-resistant signature of a block
func (hc *HybridConsensus) verifyQuantumSignature(block *types.Block) error {
	// Get validator's public key
	val, err := hc.stateStore.LoadValidators(block.Height)
	if err != nil {
		return fmt.Errorf("failed to load validators: %w", err)
	}

	proposer := val.GetByAddress(block.ProposerAddress)
	if proposer == nil {
		return errors.New("proposer not found in validator set")
	}

	// Verify signature
	valid, err := hc.quantumVerifier.Verify(
		proposer.PubKey,
		block.Hash(),
		block.QuantumSignature,
	)
	if err != nil {
		return fmt.Errorf("quantum signature verification error: %w", err)
	}
	if !valid {
		return errors.New("invalid quantum signature")
	}

	return nil
}

// OnStart implements service.Service
func (hc *HybridConsensus) OnStart() error {
	// Start base Tendermint consensus
	if err := hc.State.OnStart(); err != nil {
		return err
	}

	// Start ZK components if enabled
	if hc.config.ZKEnabled {
		if err := hc.zkProver.Start(); err != nil {
			return fmt.Errorf("failed to start zk prover: %w", err)
		}
	}

	// Start quantum components if enabled
	if hc.config.QuantumEnabled {
		if err := hc.quantumSigner.Start(); err != nil {
			return fmt.Errorf("failed to start quantum signer: %w", err)
		}
	}

	return nil
}

// OnStop implements service.Service
func (hc *HybridConsensus) OnStop() {
	// Stop base Tendermint consensus
	hc.State.OnStop()

	// Stop ZK components if enabled
	if hc.config.ZKEnabled {
		hc.zkProver.Stop()
	}

	// Stop quantum components if enabled
	if hc.config.QuantumEnabled {
		hc.quantumSigner.Stop()
	}
} 