// Package quantum_signing provides quantum-resistant digital signatures using CRYSTALS-Dilithium.
package quantum_signing

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cloudflare/circl/sign/dilithium"
)

// DilithiumSigner implements quantum-resistant digital signatures using CRYSTALS-Dilithium.
type DilithiumSigner struct {
	mode dilithium.Mode
}

// NewDilithiumSigner creates a new instance of DilithiumSigner with the recommended security level.
func NewDilithiumSigner() (*DilithiumSigner, error) {
	// Using Mode3 by default (recommended security level)
	mode := dilithium.Mode3
	return &DilithiumSigner{
		mode: mode,
	}, nil
}

// GenerateKey generates a new key pair.
func (d *DilithiumSigner) GenerateKey() ([]byte, []byte, error) {
	// Generate a new key pair
	publicKey, privateKey, err := d.mode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Serialize the keys
	skBytes, err := privateKey.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	pkBytes, err := publicKey.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	return pkBytes, skBytes, nil
}

// Sign signs a message using the private key.
func (d *DilithiumSigner) Sign(privateKey []byte, message []byte) ([]byte, error) {
	if len(message) == 0 {
		return nil, errors.New("message cannot be empty")
	}

	// Deserialize the private key
	sk := d.mode.NewKeyFromSeed(privateKey)

	// Sign the message
	signature := d.mode.Sign(sk, message)
	return signature, nil
}

// Verify verifies a signature using the public key.
func (d *DilithiumSigner) Verify(publicKey []byte, message []byte, signature []byte) (bool, error) {
	if len(message) == 0 {
		return false, errors.New("message cannot be empty")
	}

	if len(signature) == 0 {
		return false, errors.New("signature cannot be empty")
	}

	// Deserialize the public key
	pk, err := d.mode.UnmarshalBinaryPublicKey(publicKey)
	if err != nil {
		return false, fmt.Errorf("invalid public key: %w", err)
	}

	// Verify the signature
	isValid := d.mode.Verify(pk, message, signature)
	return isValid, nil
}

// PublicKey returns the public key corresponding to the given private key.
func (d *DilithiumSigner) PublicKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) == 0 {
		return nil, errors.New("private key cannot be empty")
	}

	// Deserialize the private key to get the public key
	sk := d.mode.NewKeyFromSeed(privateKey)
	pkBytes, err := sk.Public().MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	return pkBytes, nil
}

// KeySize returns the size of the keys in bytes.
func (d *DilithiumSigner) KeySize() (int, int) {
	// Return the size of public and private keys in bytes
	return d.mode.PublicKeySize(), d.mode.PrivateKeySize()
}

// SignatureSize returns the size of the signature in bytes.
func (d *DilithiumSigner) SignatureSize() int {
	return d.mode.SignatureSize()
}

// GetMode returns the security mode being used.
func (d *DilithiumSigner) GetMode() string {
	switch d.mode {
	case dilithium.Mode2:
		return "Dilithium2 (128-bit security)"
	case dilithium.Mode3:
		return "Dilithium3 (192-bit security)"
	case dilithium.Mode5:
		return "Dilithium5 (256-bit security)"
	default:
		return "Unknown"
	}
}

// PerformanceMetrics tracks performance metrics for the signer.
type PerformanceMetrics struct {
	mu                sync.Mutex
	signTimes         []time.Duration
	verifyTimes       []time.Duration
	maxSamples        int
	signCount         int
	verifyCount       int
	totalSignTime     time.Duration
	totalVerifyTime   time.Duration
}

// NewPerformanceMetrics creates a new PerformanceMetrics instance.
func NewPerformanceMetrics(maxSamples int) *PerformanceMetrics {
	return &PerformanceMetrics{
		signTimes:   make([]time.Duration, 0, maxSamples),
		verifyTimes: make([]time.Duration, 0, maxSamples),
		maxSamples:  maxSamples,
	}
}

// RecordSign records the time taken for a sign operation.
func (p *PerformanceMetrics) RecordSign(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.signCount++
	p.totalSignTime += duration

	if len(p.signTimes) >= p.maxSamples {
		p.signTimes = p.signTimes[1:]
	}
	p.signTimes = append(p.signTimes, duration)
}

// RecordVerify records the time taken for a verify operation.
func (p *PerformanceMetrics) RecordVerify(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.verifyCount++
	p.totalVerifyTime += duration

	if len(p.verifyTimes) >= p.maxSamples {
		p.verifyTimes = p.verifyTimes[1:]
	}
	p.verifyTimes = append(p.verifyTimes, duration)
}

// GetMetrics returns a map of performance metrics.
func (p *PerformanceMetrics) GetMetrics() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	metrics := make(map[string]interface{})

	// Calculate average sign time
	var avgSignTime time.Duration
	if p.signCount > 0 {
		avgSignTime = p.totalSignTime / time.Duration(p.signCount)
	}

	// Calculate average verify time
	var avgVerifyTime time.Duration
	if p.verifyCount > 0 {
		avgVerifyTime = p.totalVerifyTime / time.Duration(p.verifyCount)
	}

	// Add metrics to the map
	metrics["sign_count"] = p.signCount
	metrics["verify_count"] = p.verifyCount
	metrics["avg_sign_time"] = avgSignTime.String()
	metrics["avg_verify_time"] = avgVerifyTime.String()
	metrics["total_sign_time"] = p.totalSignTime.String()
	metrics["total_verify_time"] = p.totalVerifyTime.String()

	// Add recent samples
	recentSignTimes := make([]string, len(p.signTimes))
	for i, t := range p.signTimes {
		recentSignTimes[i] = t.String()
	}
	metrics["recent_sign_times"] = recentSignTimes

	recentVerifyTimes := make([]string, len(p.verifyTimes))
	for i, t := range p.verifyTimes {
		recentVerifyTimes[i] = t.String()
	}
	metrics["recent_verify_times"] = recentVerifyTimes

	return metrics
}

// GetStats returns a formatted string with performance statistics.
func (p *PerformanceMetrics) GetStats() string {
	metrics := p.GetMetrics()
	return fmt.Sprintf(`Performance Statistics:
  Sign Operations: %d
  Verify Operations: %d
  Average Sign Time: %s
  Average Verify Time: %s
  Total Sign Time: %s
  Total Verify Time: %s`,
		metrics["sign_count"],
		metrics["verify_count"],
		metrics["avg_sign_time"],
		metrics["avg_verify_time"],
		metrics["total_sign_time"],
		metrics["total_verify_time"],
	)
}

// GenerateRandomBytes generates a slice of random bytes of the specified length.
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// HexEncode encodes a byte slice to a hexadecimal string.
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode decodes a hexadecimal string to a byte slice.
func HexDecode(hexString string) ([]byte, error) {
	return hex.DecodeString(hexString)
}
