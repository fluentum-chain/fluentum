package plugin

import (
	"context"
	"time"
)

// SignerPlugin is the interface that all signing plugins must implement
type SignerPlugin interface {
	// GenerateKeyPair generates a new public/private key pair
	GenerateKeyPair() (publicKey []byte, privateKey []byte, err error)
	
	// Sign creates a digital signature
	Sign(privateKey []byte, message []byte) (signature []byte, err error)
	
	// SignAsync creates a digital signature asynchronously
	SignAsync(ctx context.Context, privateKey []byte, message []byte) (signature []byte, err error)
	
	// Verify checks a digital signature
	Verify(publicKey []byte, message []byte, signature []byte) (bool, error)
	
	// VerifyAsync checks a digital signature asynchronously
	VerifyAsync(ctx context.Context, publicKey []byte, message []byte, signature []byte) (bool, error)
	
	// BatchVerify verifies multiple signatures efficiently
	BatchVerify(publicKeys [][]byte, messages [][]byte, signatures [][]byte) ([]bool, error)
	
	// SignatureSize returns the expected size of signatures
	SignatureSize() int
	
	// PublicKeySize returns the expected size of public keys
	PublicKeySize() int
	
	// PrivateKeySize returns the expected size of private keys
	PrivateKeySize() int
	
	// AlgorithmName returns the name of the algorithm
	AlgorithmName() string
	
	// SecurityLevel returns the security level (e.g., "128-bit", "256-bit")
	SecurityLevel() string
	
	// PerformanceMetrics returns performance statistics
	PerformanceMetrics() map[string]float64
	
	// ResetMetrics resets performance metrics
	ResetMetrics()
	
	// IsQuantumResistant returns true if the algorithm is quantum-resistant
	IsQuantumResistant() bool
}

// PluginConfig contains configuration for the plugin
type PluginConfig struct {
	SecurityLevel    string            `json:"security_level"`    // e.g., "Dilithium3"
	BatchSize        int               `json:"batch_size"`        // Batch size for operations
	ConcurrencyLevel int               `json:"concurrency_level"` // Number of concurrent operations
	Timeout          time.Duration     `json:"timeout"`           // Operation timeout
	CustomParams     map[string]string `json:"custom_params"`     // Custom parameters
}

// PerformanceStats tracks performance metrics
type PerformanceStats struct {
	SignCount      int64         `json:"sign_count"`
	VerifyCount    int64         `json:"verify_count"`
	BatchCount     int64         `json:"batch_count"`
	TotalSignTime  time.Duration `json:"total_sign_time"`
	TotalVerifyTime time.Duration `json:"total_verify_time"`
	TotalBatchTime time.Duration `json:"total_batch_time"`
	LastReset      time.Time     `json:"last_reset"`
}

// DefaultPluginConfig returns default configuration
func DefaultPluginConfig() PluginConfig {
	return PluginConfig{
		SecurityLevel:    "Dilithium3",
		BatchSize:        100,
		ConcurrencyLevel: 4,
		Timeout:          5 * time.Second,
		CustomParams:     make(map[string]string),
	}
}

// PluginInfo contains information about a loaded plugin
type PluginInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Algorithm   string    `json:"algorithm"`
	SecurityLevel string  `json:"security_level"`
	LoadedAt    time.Time `json:"loaded_at"`
	Path        string    `json:"path"`
}

// PluginError represents plugin-specific errors
type PluginError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *PluginError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrCodeInvalidKey     = "INVALID_KEY"
	ErrCodeInvalidMessage = "INVALID_MESSAGE"
	ErrCodeInvalidSig     = "INVALID_SIGNATURE"
	ErrCodeTimeout        = "TIMEOUT"
	ErrCodeNotSupported   = "NOT_SUPPORTED"
	ErrCodeInternal       = "INTERNAL_ERROR"
) 