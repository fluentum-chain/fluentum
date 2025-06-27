package validator

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
)

// Validator represents a validator node that can sign and verify blocks
type Validator struct {
	ID              string
	PublicKey       []byte
	PrivateKey      []byte
	SignerPlugin    plugin.SignerPlugin
	FallbackSigner  *DefaultSigner
	UseQuantum      bool
	mu              sync.RWMutex
	lastBlockHeight int64
	lastBlockTime   time.Time
}

// Block represents a blockchain block
type Block struct {
	Height           int64     `json:"height"`
	Timestamp        time.Time `json:"timestamp"`
	Data             []byte    `json:"data"`
	ValidatorID      string    `json:"validator_id"`
	ValidatorPubKey  []byte    `json:"validator_pub_key"`
	Signature        []byte    `json:"signature"`
	PreviousHash     []byte    `json:"previous_hash"`
	Hash             []byte    `json:"hash"`
	QuantumSignature []byte    `json:"quantum_signature,omitempty"`
}

// DefaultSigner provides fallback Ed25519 signing
type DefaultSigner struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// NewDefaultSigner creates a new Ed25519 signer
func NewDefaultSigner() (*DefaultSigner, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	return &DefaultSigner{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

// NewValidator creates a new validator with quantum signing capability
func NewValidator(id string, useQuantum bool) (*Validator, error) {
	// Initialize fallback signer
	fallbackSigner, err := NewDefaultSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback signer: %w", err)
	}

	v := &Validator{
		ID:             id,
		PublicKey:      fallbackSigner.publicKey,
		PrivateKey:     fallbackSigner.privateKey,
		FallbackSigner: fallbackSigner,
		UseQuantum:     useQuantum,
	}

	// Try to load quantum signer if requested
	if useQuantum {
		pm := plugin.Instance()
		if pm.GetPluginCount() > 0 {
			quantumSigner, err := pm.GetSigner()
			if err == nil {
				v.SignerPlugin = quantumSigner
				fmt.Printf("Validator initialized with quantum signing: %s\n", quantumSigner.AlgorithmName())
			} else {
				fmt.Printf("Warning: Failed to load quantum signer, falling back to Ed25519: %v\n", err)
				v.UseQuantum = false
			}
		} else {
			fmt.Printf("Warning: No plugins available, falling back to Ed25519\n")
			v.UseQuantum = false
		}
	}

	return v, nil
}

// SignBlock signs a block using the appropriate signing algorithm
func (v *Validator) SignBlock(block *Block) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Validate block
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}

	if block.Height <= v.lastBlockHeight {
		return fmt.Errorf("block height must be greater than last signed block")
	}

	// Prepare block data for signing (exclude signature fields)
	blockData, err := v.prepareBlockData(block)
	if err != nil {
		return fmt.Errorf("failed to prepare block data: %w", err)
	}

	// Sign with quantum algorithm if available
	if v.UseQuantum && v.SignerPlugin != nil {
		signature, err := v.SignerPlugin.Sign(v.PrivateKey, blockData)
		if err != nil {
			return fmt.Errorf("quantum signing failed: %w", err)
		}
		block.QuantumSignature = signature
		block.Signature = signature // For backward compatibility
	} else {
		// Fallback to Ed25519
		signature := ed25519.Sign(v.FallbackSigner.privateKey, blockData)
		block.Signature = signature
	}

	// Update block metadata
	block.ValidatorID = v.ID
	block.ValidatorPubKey = v.PublicKey
	block.Timestamp = time.Now()

	// Update validator state
	v.lastBlockHeight = block.Height
	v.lastBlockTime = block.Timestamp

	return nil
}

// SignBlockAsync signs a block asynchronously
func (v *Validator) SignBlockAsync(ctx context.Context, block *Block) error {
	if v.UseQuantum && v.SignerPlugin != nil {
		blockData, err := v.prepareBlockData(block)
		if err != nil {
			return fmt.Errorf("failed to prepare block data: %w", err)
		}

		signature, err := v.SignerPlugin.SignAsync(ctx, v.PrivateKey, blockData)
		if err != nil {
			return fmt.Errorf("async quantum signing failed: %w", err)
		}

		block.QuantumSignature = signature
		block.Signature = signature
		block.ValidatorID = v.ID
		block.ValidatorPubKey = v.PublicKey
		block.Timestamp = time.Now()

		return nil
	}

	// Fallback to synchronous signing for Ed25519
	return v.SignBlock(block)
}

// VerifyBlock verifies a block signature
func (v *Validator) VerifyBlock(block *Block) (bool, error) {
	if block == nil {
		return false, fmt.Errorf("block cannot be nil")
	}

	// Prepare block data for verification (exclude signature fields)
	blockData, err := v.prepareBlockData(block)
	if err != nil {
		return false, fmt.Errorf("failed to prepare block data: %w", err)
	}

	// Verify quantum signature if present
	if len(block.QuantumSignature) > 0 && v.SignerPlugin != nil {
		valid, err := v.SignerPlugin.Verify(block.ValidatorPubKey, blockData, block.QuantumSignature)
		if err != nil {
			return false, fmt.Errorf("quantum signature verification failed: %w", err)
		}
		return valid, nil
	}

	// Fallback to Ed25519 verification
	valid := ed25519.Verify(block.ValidatorPubKey, blockData, block.Signature)
	return valid, nil
}

// VerifyBlockAsync verifies a block signature asynchronously
func (v *Validator) VerifyBlockAsync(ctx context.Context, block *Block) (bool, error) {
	if block == nil {
		return false, fmt.Errorf("block cannot be nil")
	}

	// Prepare block data for verification
	blockData, err := v.prepareBlockData(block)
	if err != nil {
		return false, fmt.Errorf("failed to prepare block data: %w", err)
	}

	// Verify quantum signature if present
	if len(block.QuantumSignature) > 0 && v.SignerPlugin != nil {
		valid, err := v.SignerPlugin.VerifyAsync(ctx, block.ValidatorPubKey, blockData, block.QuantumSignature)
		if err != nil {
			return false, fmt.Errorf("async quantum signature verification failed: %w", err)
		}
		return valid, nil
	}

	// Fallback to synchronous verification for Ed25519
	return v.VerifyBlock(block)
}

// prepareBlockData prepares block data for signing/verification
func (v *Validator) prepareBlockData(block *Block) ([]byte, error) {
	// Create a copy of the block without signature fields
	blockCopy := &Block{
		Height:          block.Height,
		Timestamp:       block.Timestamp,
		Data:            block.Data,
		ValidatorID:     block.ValidatorID,
		ValidatorPubKey: block.ValidatorPubKey,
		PreviousHash:    block.PreviousHash,
		Hash:            block.Hash,
	}

	// Serialize the block data
	data, err := json.Marshal(blockCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block: %w", err)
	}

	return data, nil
}

// GetPublicKey returns the validator's public key
func (v *Validator) GetPublicKey() []byte {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.PublicKey
}

// GetPublicKeyHex returns the validator's public key as a hex string
func (v *Validator) GetPublicKeyHex() string {
	return hex.EncodeToString(v.GetPublicKey())
}

// GetSignerInfo returns information about the current signer
func (v *Validator) GetSignerInfo() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	info := map[string]interface{}{
		"validator_id": v.ID,
		"use_quantum":  v.UseQuantum,
		"public_key":   hex.EncodeToString(v.PublicKey),
	}

	if v.UseQuantum && v.SignerPlugin != nil {
		info["algorithm"] = v.SignerPlugin.AlgorithmName()
		info["security_level"] = v.SignerPlugin.SecurityLevel()
		info["quantum_resistant"] = v.SignerPlugin.IsQuantumResistant()
		info["signature_size"] = v.SignerPlugin.SignatureSize()
		info["public_key_size"] = v.SignerPlugin.PublicKeySize()
		info["performance_metrics"] = v.SignerPlugin.PerformanceMetrics()
	} else {
		info["algorithm"] = "Ed25519"
		info["security_level"] = "128-bit"
		info["quantum_resistant"] = false
		info["signature_size"] = 64
		info["public_key_size"] = 32
	}

	return info
}

// SwitchToQuantum switches the validator to use quantum signing
func (v *Validator) SwitchToQuantum() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.UseQuantum {
		return fmt.Errorf("validator already using quantum signing")
	}

	pm := plugin.Instance()
	if pm.GetPluginCount() > 0 {
		quantumSigner, err := pm.GetSigner()
		if err != nil {
			return fmt.Errorf("failed to get quantum signer: %w", err)
		}

		v.SignerPlugin = quantumSigner
		v.UseQuantum = true
		return nil
	}

	return fmt.Errorf("no quantum signer plugin available")
}

// SwitchToClassical switches the validator to use classical signing
func (v *Validator) SwitchToClassical() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.UseQuantum = false
	v.SignerPlugin = nil
}

// GetLastBlockInfo returns information about the last signed block
func (v *Validator) GetLastBlockInfo() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return map[string]interface{}{
		"last_block_height": v.lastBlockHeight,
		"last_block_time":   v.lastBlockTime,
		"validator_id":      v.ID,
	}
}

// ResetLastBlockInfo resets the last block information
func (v *Validator) ResetLastBlockInfo() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.lastBlockHeight = 0
	v.lastBlockTime = time.Time{}
}
