package main

import (
	"C"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/fluentum-chain/fluentum/fluentum/core/plugin"
	"github.com/cloudflare/circl/sign/dilithium"
)

// QuantumSigner implements the SignerPlugin interface
type QuantumSigner struct {
	mode       dilithium.Mode
	stats      *plugin.PerformanceStats
	statsMutex sync.RWMutex
	config     plugin.PluginConfig
}

// exported symbol that will be looked for when loading the plugin
var SignerPlugin QuantumSigner

func init() {
	// Initialize with Dilithium3 by default
	SignerPlugin = QuantumSigner{
		mode:  dilithium.Mode3,
		stats: &plugin.PerformanceStats{},
		config: plugin.DefaultPluginConfig(),
	}
}

// Initialize initializes the plugin with configuration
//export Initialize
func Initialize(configJSON *C.char) error {
	var config plugin.PluginConfig
	err := json.Unmarshal([]byte(C.GoString(configJSON)), &config)
	if err != nil {
		return &plugin.PluginError{
			Code:    plugin.ErrCodeInternal,
			Message: "failed to parse configuration",
			Details: err.Error(),
		}
	}

	// Set the mode based on security level
	switch config.SecurityLevel {
	case "Dilithium2":
		SignerPlugin.mode = dilithium.Mode2
	case "Dilithium5":
		SignerPlugin.mode = dilithium.Mode5
	case "Dilithium3":
		fallthrough
	default:
		SignerPlugin.mode = dilithium.Mode3
	}

	SignerPlugin.config = config
	SignerPlugin.stats = &plugin.PerformanceStats{
		LastReset: time.Now(),
	}

	return nil
}

func (q *QuantumSigner) GenerateKeyPair() ([]byte, []byte, error) {
	pk, sk := q.mode.GenerateKey()
	return pk.Bytes(), sk.Bytes(), nil
}

func (q *QuantumSigner) Sign(privateKey []byte, message []byte) ([]byte, error) {
	if len(privateKey) != q.mode.PrivateKeySize() {
		return nil, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidKey,
			Message: "invalid private key size",
			Details: fmt.Sprintf("expected %d, got %d", q.mode.PrivateKeySize(), len(privateKey)),
		}
	}

	if len(message) == 0 {
		return nil, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidMessage,
			Message: "message cannot be empty",
		}
	}

	start := time.Now()

	sk := q.mode.NewKeyFromSeed(privateKey)
	signature := make([]byte, q.mode.SignatureSize())
	sk.Sign(signature, message, nil)

	q.statsMutex.Lock()
	defer q.statsMutex.Unlock()
	q.stats.SignCount++
	q.stats.TotalSignTime += time.Since(start)

	return signature, nil
}

func (q *QuantumSigner) SignAsync(ctx context.Context, privateKey []byte, message []byte) ([]byte, error) {
	// Create a channel for the result
	resultChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		signature, err := q.Sign(privateKey, message)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- signature
	}()

	// Wait for result or timeout
	select {
	case signature := <-resultChan:
		return signature, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, &plugin.PluginError{
			Code:    plugin.ErrCodeTimeout,
			Message: "signing operation timed out",
			Details: ctx.Err().Error(),
		}
	}
}

func (q *QuantumSigner) Verify(publicKey []byte, message []byte, signature []byte) (bool, error) {
	if len(publicKey) != q.mode.PublicKeySize() {
		return false, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidKey,
			Message: "invalid public key size",
			Details: fmt.Sprintf("expected %d, got %d", q.mode.PublicKeySize(), len(publicKey)),
		}
	}

	if len(signature) != q.mode.SignatureSize() {
		return false, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidSig,
			Message: "invalid signature size",
			Details: fmt.Sprintf("expected %d, got %d", q.mode.SignatureSize(), len(signature)),
		}
	}

	if len(message) == 0 {
		return false, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidMessage,
			Message: "message cannot be empty",
		}
	}

	start := time.Now()

	pk := q.mode.NewPublicKeyFromBytes(publicKey)
	valid := pk.Verify(message, signature)

	q.statsMutex.Lock()
	defer q.statsMutex.Unlock()
	q.stats.VerifyCount++
	q.stats.TotalVerifyTime += time.Since(start)

	return valid, nil
}

func (q *QuantumSigner) VerifyAsync(ctx context.Context, publicKey []byte, message []byte, signature []byte) (bool, error) {
	// Create a channel for the result
	resultChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	go func() {
		valid, err := q.Verify(publicKey, message, signature)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- valid
	}()

	// Wait for result or timeout
	select {
	case valid := <-resultChan:
		return valid, nil
	case err := <-errChan:
		return false, err
	case <-ctx.Done():
		return false, &plugin.PluginError{
			Code:    plugin.ErrCodeTimeout,
			Message: "verification operation timed out",
			Details: ctx.Err().Error(),
		}
	}
}

func (q *QuantumSigner) BatchVerify(publicKeys [][]byte, messages [][]byte, signatures [][]byte) ([]bool, error) {
	if len(publicKeys) != len(messages) || len(messages) != len(signatures) {
		return nil, &plugin.PluginError{
			Code:    plugin.ErrCodeInvalidMessage,
			Message: "batch size mismatch",
			Details: fmt.Sprintf("publicKeys: %d, messages: %d, signatures: %d", len(publicKeys), len(messages), len(signatures)),
		}
	}

	if len(publicKeys) == 0 {
		return []bool{}, nil
	}

	start := time.Now()
	results := make([]bool, len(publicKeys))

	// Use goroutines for concurrent verification
	concurrency := q.config.ConcurrencyLevel
	if concurrency <= 0 {
		concurrency = 4
	}

	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := range publicKeys {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			valid, err := q.Verify(publicKeys[index], messages[index], signatures[index])
			if err != nil {
				// Log error but continue with other verifications
				fmt.Printf("Batch verification error at index %d: %v\n", index, err)
				results[index] = false
			} else {
				results[index] = valid
			}
		}(i)
	}

	wg.Wait()

	q.statsMutex.Lock()
	defer q.statsMutex.Unlock()
	q.stats.BatchCount++
	q.stats.TotalBatchTime += time.Since(start)

	return results, nil
}

func (q *QuantumSigner) SignatureSize() int {
	return q.mode.SignatureSize()
}

func (q *QuantumSigner) PublicKeySize() int {
	return q.mode.PublicKeySize()
}

func (q *QuantumSigner) PrivateKeySize() int {
	return q.mode.PrivateKeySize()
}

func (q *QuantumSigner) AlgorithmName() string {
	return q.mode.String()
}

func (q *QuantumSigner) SecurityLevel() string {
	switch q.mode {
	case dilithium.Mode2:
		return "128-bit"
	case dilithium.Mode3:
		return "192-bit"
	case dilithium.Mode5:
		return "256-bit"
	default:
		return "unknown"
	}
}

func (q *QuantumSigner) PerformanceMetrics() map[string]float64 {
	q.statsMutex.RLock()
	defer q.statsMutex.RUnlock()

	metrics := make(map[string]float64)

	// Calculate averages
	if q.stats.SignCount > 0 {
		metrics["avg_sign_time_ms"] = float64(q.stats.TotalSignTime.Microseconds()) / float64(q.stats.SignCount) / 1000
	}
	if q.stats.VerifyCount > 0 {
		metrics["avg_verify_time_ms"] = float64(q.stats.TotalVerifyTime.Microseconds()) / float64(q.stats.VerifyCount) / 1000
	}
	if q.stats.BatchCount > 0 {
		metrics["avg_batch_time_ms"] = float64(q.stats.TotalBatchTime.Microseconds()) / float64(q.stats.BatchCount) / 1000
	}

	// Add counts
	metrics["total_sign_count"] = float64(q.stats.SignCount)
	metrics["total_verify_count"] = float64(q.stats.VerifyCount)
	metrics["total_batch_count"] = float64(q.stats.BatchCount)

	// Add sizes
	metrics["signature_size_bytes"] = float64(q.SignatureSize())
	metrics["public_key_size_bytes"] = float64(q.PublicKeySize())
	metrics["private_key_size_bytes"] = float64(q.PrivateKeySize())

	// Add uptime
	metrics["uptime_seconds"] = time.Since(q.stats.LastReset).Seconds()

	return metrics
}

func (q *QuantumSigner) ResetMetrics() {
	q.statsMutex.Lock()
	defer q.statsMutex.Unlock()

	q.stats = &plugin.PerformanceStats{
		LastReset: time.Now(),
	}
}

func (q *QuantumSigner) IsQuantumResistant() bool {
	return true
}

func main() {} // Required but unused 