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
)

// HybridSigner provides dual classical and quantum signing capabilities
type HybridSigner struct {
	classicSigner *DefaultSigner
	quantumSigner plugin.SignerPlugin
	useQuantum    bool
	mu            sync.RWMutex
	stats         *HybridSignerStats
}

// HybridSignerStats tracks performance metrics for hybrid signing
type HybridSignerStats struct {
	ClassicSignCount   int64         `json:"classic_sign_count"`
	QuantumSignCount   int64         `json:"quantum_sign_count"`
	ClassicVerifyCount int64         `json:"classic_verify_count"`
	QuantumVerifyCount int64         `json:"quantum_verify_count"`
	TotalClassicTime   time.Duration `json:"total_classic_time"`
	TotalQuantumTime   time.Duration `json:"total_quantum_time"`
	LastReset          time.Time     `json:"last_reset"`
}

// HybridSignature contains both classical and quantum signatures
type HybridSignature struct {
	ClassicSignature []byte `json:"classic_signature"`
	QuantumSignature []byte `json:"quantum_signature"`
	Timestamp        time.Time `json:"timestamp"`
	Mode             string   `json:"mode"` // "dual", "classic", "quantum"
}

// NewHybridSigner creates a new hybrid signer
func NewHybridSigner() (*HybridSigner, error) {
	// Initialize classical signer
	classicSigner, err := NewDefaultSigner()
	if err != nil {
		return nil, fmt.Errorf("failed to create classical signer: %w", err)
	}

	hs := &HybridSigner{
		classicSigner: classicSigner,
		useQuantum:    false,
		stats: &HybridSignerStats{
			LastReset: time.Now(),
		},
	}

	// Try to load quantum signer
	pm := plugin.Instance()
	if pm.IsPluginLoaded() {
		quantumSigner, err := pm.GetSigner()
		if err == nil {
			hs.quantumSigner = quantumSigner
			fmt.Printf("Hybrid signer initialized with quantum support: %s\n", quantumSigner.AlgorithmName())
		} else {
			fmt.Printf("Warning: Failed to load quantum signer: %v\n", err)
		}
	}

	return hs, nil
}

// Sign creates a hybrid signature
func (hs *HybridSigner) Sign(privateKey []byte, message []byte) (*HybridSignature, error) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	signature := &HybridSignature{
		Timestamp: time.Now(),
	}

	// Sign with classical algorithm
	start := time.Now()
	classicSig := ed25519.Sign(hs.classicSigner.privateKey, message)
	hs.stats.TotalClassicTime += time.Since(start)
	hs.stats.ClassicSignCount++
	signature.ClassicSignature = classicSig

	// Sign with quantum algorithm if available
	if hs.quantumSigner != nil && hs.useQuantum {
		start := time.Now()
		quantumSig, err := hs.quantumSigner.Sign(privateKey, message)
		if err != nil {
			return nil, fmt.Errorf("quantum signing failed: %w", err)
		}
		hs.stats.TotalQuantumTime += time.Since(start)
		hs.stats.QuantumSignCount++
		signature.QuantumSignature = quantumSig
		signature.Mode = "dual"
	} else {
		signature.Mode = "classic"
	}

	return signature, nil
}

// SignAsync creates a hybrid signature asynchronously
func (hs *HybridSigner) SignAsync(ctx context.Context, privateKey []byte, message []byte) (*HybridSignature, error) {
	// Create channels for results
	classicChan := make(chan []byte, 1)
	quantumChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	// Start classical signing
	go func() {
		start := time.Now()
		sig := ed25519.Sign(hs.classicSigner.privateKey, message)
		hs.mu.Lock()
		hs.stats.TotalClassicTime += time.Since(start)
		hs.stats.ClassicSignCount++
		hs.mu.Unlock()
		classicChan <- sig
	}()

	// Start quantum signing if available
	if hs.quantumSigner != nil && hs.useQuantum {
		go func() {
			start := time.Now()
			sig, err := hs.quantumSigner.SignAsync(ctx, privateKey, message)
			if err != nil {
				errChan <- err
				return
			}
			hs.mu.Lock()
			hs.stats.TotalQuantumTime += time.Since(start)
			hs.stats.QuantumSignCount++
			hs.mu.Unlock()
			quantumChan <- sig
		}()
	}

	// Wait for results
	signature := &HybridSignature{
		Timestamp: time.Now(),
	}

	// Get classical signature
	select {
	case classicSig := <-classicChan:
		signature.ClassicSignature = classicSig
	case <-ctx.Done():
		return nil, fmt.Errorf("classical signing timed out: %w", ctx.Err())
	}

	// Get quantum signature if available
	if hs.quantumSigner != nil && hs.useQuantum {
		select {
		case quantumSig := <-quantumChan:
			signature.QuantumSignature = quantumSig
			signature.Mode = "dual"
		case err := <-errChan:
			return nil, fmt.Errorf("quantum signing failed: %w", err)
		case <-ctx.Done():
			// Quantum signing timed out, but we still have classical signature
			signature.Mode = "classic"
		}
	} else {
		signature.Mode = "classic"
	}

	return signature, nil
}

// Verify verifies a hybrid signature
func (hs *HybridSigner) Verify(publicKey []byte, message []byte, signature *HybridSignature) (bool, error) {
	if signature == nil {
		return false, fmt.Errorf("signature cannot be nil")
	}

	// Verify classical signature
	if len(signature.ClassicSignature) == 0 {
		return false, fmt.Errorf("no classical signature found")
	}

	start := time.Now()
	classicValid := ed25519.Verify(hs.classicSigner.publicKey, message, signature.ClassicSignature)
	hs.mu.Lock()
	hs.stats.TotalClassicTime += time.Since(start)
	hs.stats.ClassicVerifyCount++
	hs.mu.Unlock()

	if !classicValid {
		return false, fmt.Errorf("classical signature verification failed")
	}

	// Verify quantum signature if present
	if len(signature.QuantumSignature) > 0 && hs.quantumSigner != nil {
		start := time.Now()
		quantumValid, err := hs.quantumSigner.Verify(publicKey, message, signature.QuantumSignature)
		if err != nil {
			return false, fmt.Errorf("quantum signature verification failed: %w", err)
		}
		hs.mu.Lock()
		hs.stats.TotalQuantumTime += time.Since(start)
		hs.stats.QuantumVerifyCount++
		hs.mu.Unlock()

		if !quantumValid {
			return false, fmt.Errorf("quantum signature verification failed")
		}

		return true, nil
	}

	// Only classical signature was provided
	return true, nil
}

// VerifyAsync verifies a hybrid signature asynchronously
func (hs *HybridSigner) VerifyAsync(ctx context.Context, publicKey []byte, message []byte, signature *HybridSignature) (bool, error) {
	if signature == nil {
		return false, fmt.Errorf("signature cannot be nil")
	}

	// Create channels for results
	classicChan := make(chan bool, 1)
	quantumChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	// Start classical verification
	go func() {
		start := time.Now()
		valid := ed25519.Verify(hs.classicSigner.publicKey, message, signature.ClassicSignature)
		hs.mu.Lock()
		hs.stats.TotalClassicTime += time.Since(start)
		hs.stats.ClassicVerifyCount++
		hs.mu.Unlock()
		classicChan <- valid
	}()

	// Start quantum verification if signature present
	if len(signature.QuantumSignature) > 0 && hs.quantumSigner != nil {
		go func() {
			start := time.Now()
			valid, err := hs.quantumSigner.VerifyAsync(ctx, publicKey, message, signature.QuantumSignature)
			if err != nil {
				errChan <- err
				return
			}
			hs.mu.Lock()
			hs.stats.TotalQuantumTime += time.Since(start)
			hs.stats.QuantumVerifyCount++
			hs.mu.Unlock()
			quantumChan <- valid
		}()
	}

	// Wait for classical verification
	var classicValid bool
	select {
	case valid := <-classicChan:
		classicValid = valid
	case <-ctx.Done():
		return false, fmt.Errorf("classical verification timed out: %w", ctx.Err())
	}

	if !classicValid {
		return false, fmt.Errorf("classical signature verification failed")
	}

	// Wait for quantum verification if needed
	if len(signature.QuantumSignature) > 0 && hs.quantumSigner != nil {
		select {
		case quantumValid := <-quantumChan:
			if !quantumValid {
				return false, fmt.Errorf("quantum signature verification failed")
			}
		case err := <-errChan:
			return false, fmt.Errorf("quantum verification failed: %w", err)
		case <-ctx.Done():
			return false, fmt.Errorf("quantum verification timed out: %w", ctx.Err())
		}
	}

	return true, nil
}

// EnableQuantum enables quantum signing
func (hs *HybridSigner) EnableQuantum() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.quantumSigner == nil {
		return fmt.Errorf("no quantum signer available")
	}

	hs.useQuantum = true
	return nil
}

// DisableQuantum disables quantum signing
func (hs *HybridSigner) DisableQuantum() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.useQuantum = false
}

// IsQuantumEnabled returns true if quantum signing is enabled
func (hs *HybridSigner) IsQuantumEnabled() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.useQuantum && hs.quantumSigner != nil
}

// GetStats returns performance statistics
func (hs *HybridSigner) GetStats() *HybridSignerStats {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	stats := *hs.stats // Copy to avoid race conditions
	return &stats
}

// ResetStats resets performance statistics
func (hs *HybridSigner) ResetStats() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.stats = &HybridSignerStats{
		LastReset: time.Now(),
	}
}

// GetPerformanceMetrics returns detailed performance metrics
func (hs *HybridSigner) GetPerformanceMetrics() map[string]float64 {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	metrics := make(map[string]float64)

	// Classical metrics
	if hs.stats.ClassicSignCount > 0 {
		metrics["avg_classic_sign_time_ms"] = float64(hs.stats.TotalClassicTime.Microseconds()) / float64(hs.stats.ClassicSignCount) / 1000
	}
	if hs.stats.ClassicVerifyCount > 0 {
		metrics["avg_classic_verify_time_ms"] = float64(hs.stats.TotalClassicTime.Microseconds()) / float64(hs.stats.ClassicVerifyCount) / 1000
	}

	// Quantum metrics
	if hs.stats.QuantumSignCount > 0 {
		metrics["avg_quantum_sign_time_ms"] = float64(hs.stats.TotalQuantumTime.Microseconds()) / float64(hs.stats.QuantumSignCount) / 1000
	}
	if hs.stats.QuantumVerifyCount > 0 {
		metrics["avg_quantum_verify_time_ms"] = float64(hs.stats.TotalQuantumTime.Microseconds()) / float64(hs.stats.QuantumVerifyCount) / 1000
	}

	// Counts
	metrics["total_classic_sign_count"] = float64(hs.stats.ClassicSignCount)
	metrics["total_quantum_sign_count"] = float64(hs.stats.QuantumSignCount)
	metrics["total_classic_verify_count"] = float64(hs.stats.ClassicVerifyCount)
	metrics["total_quantum_verify_count"] = float64(hs.stats.QuantumVerifyCount)

	// Sizes
	metrics["classic_signature_size_bytes"] = 64.0 // Ed25519 signature size
	if hs.quantumSigner != nil {
		metrics["quantum_signature_size_bytes"] = float64(hs.quantumSigner.SignatureSize())
		metrics["quantum_public_key_size_bytes"] = float64(hs.quantumSigner.PublicKeySize())
	}

	// Uptime
	metrics["uptime_seconds"] = time.Since(hs.stats.LastReset).Seconds()

	return metrics
}

// GetInfo returns information about the hybrid signer
func (hs *HybridSigner) GetInfo() map[string]interface{} {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	info := map[string]interface{}{
		"classic_public_key": hex.EncodeToString(hs.classicSigner.publicKey),
		"quantum_enabled":    hs.useQuantum,
		"quantum_available":  hs.quantumSigner != nil,
	}

	if hs.quantumSigner != nil {
		info["quantum_algorithm"] = hs.quantumSigner.AlgorithmName()
		info["quantum_security_level"] = hs.quantumSigner.SecurityLevel()
		info["quantum_resistant"] = hs.quantumSigner.IsQuantumResistant()
	}

	return info
} 