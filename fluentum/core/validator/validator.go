package validator

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	"github.com/fluentum-chain/fluentum/fluentum/types"
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
	v := &Validator{
		ID:         id,
		UseQuantum: useQuantum,
	}

	// Initialize fallback signer
	fallbackSigner, err := NewDefaultSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to create fallback signer: %w", err)
	}
	v.FallbackSigner = fallbackSigner

	// Try to load quantum signing plugin
	if useQuantum {
		pm := plugin.Instance()
		if pm.IsPluginLoaded() {
			signer, err := pm.GetSigner()
			if err != nil {
				fmt.Printf("Warning: Failed to get quantum signer: %v, falling back to Ed25519\n", err)
				v.UseQuantum = false
			} else {
				v.SignerPlugin = signer
				// Generate quantum key pair
				pk, sk, err := signer.GenerateKeyPair()
				if err != nil {
					return nil, fmt.Errorf("failed to generate quantum key pair: %w", err)
				}
				v.PublicKey = pk
				v.PrivateKey = sk
				fmt.Printf("Quantum signer loaded: %s (Security: %s)\n", signer.AlgorithmName(), signer.SecurityLevel())
			}
		} else {
			fmt.Printf("Warning: No quantum plugin loaded, falling back to Ed25519\n")
			v.UseQuantum = false
		}
	}

	// If not using quantum or quantum failed, use Ed25519
	if !v.UseQuantum {
		v.PublicKey = v.FallbackSigner.publicKey
		v.PrivateKey = v.FallbackSigner.privateKey
		fmt.Printf("Using Ed25519 signer for validator %s\n", id)
	}

	return v, nil
}

// SignBlock signs a block with the appropriate signing algorithm
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

	// Prepare block data for verification
	blockData, err := v.prepareBlockData(block)
	if err != nil {
		return false, fmt.Errorf("failed to prepare block data: %w", err)
	}

	// Check if this is a quantum signature
	if len(block.QuantumSignature) > 0 && v.SignerPlugin != nil {
		// Verify quantum signature
		valid, err := v.SignerPlugin.Verify(block.ValidatorPubKey, blockData, block.QuantumSignature)
		if err != nil {
			return false, fmt.Errorf("quantum verification failed: %w", err)
		}
		return valid, nil
	}

	// Verify Ed25519 signature
	if len(block.Signature) == 0 {
		return false, fmt.Errorf("no signature found")
	}

	valid := ed25519.Verify(block.ValidatorPubKey, blockData, block.Signature)
	return valid, nil
}

// VerifyBlockAsync verifies a block signature asynchronously
func (v *Validator) VerifyBlockAsync(ctx context.Context, block *Block) (bool, error) {
	if len(block.QuantumSignature) > 0 && v.SignerPlugin != nil {
		blockData, err := v.prepareBlockData(block)
		if err != nil {
			return false, fmt.Errorf("failed to prepare block data: %w", err)
		}

		valid, err := v.SignerPlugin.VerifyAsync(ctx, block.ValidatorPubKey, blockData, block.QuantumSignature)
		if err != nil {
			return false, fmt.Errorf("async quantum verification failed: %w", err)
		}
		return valid, nil
	}

	// Fallback to synchronous verification for Ed25519
	return v.VerifyBlock(block)
}

// BatchVerifyBlocks verifies multiple blocks efficiently
func (v *Validator) BatchVerifyBlocks(blocks []*Block) ([]bool, error) {
	if len(blocks) == 0 {
		return []bool{}, nil
	}

	// Separate quantum and classical blocks
	var quantumBlocks []*Block
	var classicalBlocks []*Block

	for _, block := range blocks {
		if len(block.QuantumSignature) > 0 {
			quantumBlocks = append(quantumBlocks, block)
		} else {
			classicalBlocks = append(classicalBlocks, block)
		}
	}

	results := make([]bool, len(blocks))
	blockIndex := 0

	// Verify quantum blocks in batch if possible
	if len(quantumBlocks) > 0 && v.SignerPlugin != nil {
		publicKeys := make([][]byte, len(quantumBlocks))
		messages := make([][]byte, len(quantumBlocks))
		signatures := make([][]byte, len(quantumBlocks))

		for i, block := range quantumBlocks {
			blockData, err := v.prepareBlockData(block)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare block data for block %d: %w", i, err)
			}

			publicKeys[i] = block.ValidatorPubKey
			messages[i] = blockData
			signatures[i] = block.QuantumSignature
		}

		quantumResults, err := v.SignerPlugin.BatchVerify(publicKeys, messages, signatures)
		if err != nil {
			return nil, fmt.Errorf("batch quantum verification failed: %w", err)
		}

		// Map results back to original block indices
		for i, valid := range quantumResults {
			results[blockIndex] = valid
			blockIndex++
		}
	}

	// Verify classical blocks individually
	for _, block := range classicalBlocks {
		valid, err := v.VerifyBlock(block)
		if err != nil {
			return nil, fmt.Errorf("classical verification failed: %w", err)
		}
		results[blockIndex] = valid
		blockIndex++
	}

	return results, nil
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
	data, err := types.Encode(blockCopy)
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

	pm := plugin.Instance()
	if !pm.IsPluginLoaded() {
		return fmt.Errorf("no quantum plugin loaded")
	}

	signer, err := pm.GetSigner()
	if err != nil {
		return fmt.Errorf("failed to get quantum signer: %w", err)
	}

	// Generate new quantum key pair
	pk, sk, err := signer.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate quantum key pair: %w", err)
	}

	v.SignerPlugin = signer
	v.PublicKey = pk
	v.PrivateKey = sk
	v.UseQuantum = true

	return nil
}

// SwitchToClassical switches the validator to use classical Ed25519 signing
func (v *Validator) SwitchToClassical() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Generate new Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	v.FallbackSigner = &DefaultSigner{
		publicKey:  publicKey,
		privateKey: privateKey,
	}
	v.PublicKey = publicKey
	v.PrivateKey = privateKey
	v.UseQuantum = false
	v.SignerPlugin = nil

	return nil
} 
