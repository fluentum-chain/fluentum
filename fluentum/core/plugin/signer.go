package plugin

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
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
	SignCount       int64         `json:"sign_count"`
	VerifyCount     int64         `json:"verify_count"`
	BatchCount      int64         `json:"batch_count"`
	TotalSignTime   time.Duration `json:"total_sign_time"`
	TotalVerifyTime time.Duration `json:"total_verify_time"`
	TotalBatchTime  time.Duration `json:"total_batch_time"`
	LastReset       time.Time     `json:"last_reset"`
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
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Algorithm     string    `json:"algorithm"`
	SecurityLevel string    `json:"security_level"`
	LoadedAt      time.Time `json:"loaded_at"`
	Path          string    `json:"path"`
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

// SignerPlugin interface for quantum-resistant signing
type SignerPlugin interface {
	// Initialize the signer with configuration
	Initialize(config map[string]interface{}) error

	// Generate a new key pair
	GenerateKeyPair() (*KeyPair, error)

	// Sign data with the private key
	Sign(data []byte) ([]byte, error)

	// Sign data with specific algorithm
	SignWithAlgorithm(data []byte, algorithm string) ([]byte, error)

	// Verify signature with public key
	Verify(data, signature []byte, publicKey []byte) (bool, error)

	// Get public key
	GetPublicKey() ([]byte, error)

	// Get private key (encrypted)
	GetPrivateKey() ([]byte, error)

	// Import key pair
	ImportKeyPair(privateKey, publicKey []byte) error

	// Export key pair
	ExportKeyPair() (*KeyPair, error)

	// Get supported algorithms
	GetSupportedAlgorithms() []string

	// Get signer info
	GetSignerInfo() map[string]string

	// Get performance metrics
	GetMetrics() map[string]float64

	// Reset metrics
	ResetMetrics()

	// Update configuration
	UpdateConfig(config map[string]interface{}) error
}

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	PrivateKey []byte    `json:"private_key"`
	PublicKey  []byte    `json:"public_key"`
	Algorithm  string    `json:"algorithm"`
	Created    time.Time `json:"created"`
	ID         string    `json:"id"`
}

// SignerConfig contains configuration for the signer
type SignerConfig struct {
	Algorithm           string `json:"algorithm"`
	KeySize             int    `json:"key_size"`
	HashAlgorithm       string `json:"hash_algorithm"`
	EnableQuantumResist bool   `json:"enable_quantum_resist"`
	EnableHybrid        bool   `json:"enable_hybrid"`
	KeyStoragePath      string `json:"key_storage_path"`
	EncryptionEnabled   bool   `json:"encryption_enabled"`
	EncryptionPassword  string `json:"encryption_password"`
}

// DefaultSignerConfig returns default configuration
func DefaultSignerConfig() *SignerConfig {
	return &SignerConfig{
		Algorithm:           "dilithium3",
		KeySize:             2048,
		HashAlgorithm:       "sha256",
		EnableQuantumResist: true,
		EnableHybrid:        true,
		KeyStoragePath:      "./keys",
		EncryptionEnabled:   true,
	}
}

// SignerMetrics tracks signer performance metrics
type SignerMetrics struct {
	SignCount       int64         `json:"sign_count"`
	VerifyCount     int64         `json:"verify_count"`
	AvgSignTime     time.Duration `json:"avg_sign_time"`
	AvgVerifyTime   time.Duration `json:"avg_verify_time"`
	TotalSignTime   time.Duration `json:"total_sign_time"`
	TotalVerifyTime time.Duration `json:"total_verify_time"`
	ErrorCount      int64         `json:"error_count"`
	LastSignTime    time.Time     `json:"last_sign_time"`
	LastVerifyTime  time.Time     `json:"last_verify_time"`
}

// QuantumSigner implements quantum-resistant signing
type QuantumSigner struct {
	config     *SignerConfig
	keyPair    *KeyPair
	metrics    *SignerMetrics
	algorithms map[string]SigningAlgorithm
	mutex      sync.RWMutex
}

// SigningAlgorithm represents a signing algorithm
type SigningAlgorithm interface {
	Name() string
	GenerateKeyPair() (*KeyPair, error)
	Sign(data []byte, privateKey []byte) ([]byte, error)
	Verify(data, signature, publicKey []byte) (bool, error)
	GetKeySize() int
	IsQuantumResistant() bool
}

// NewQuantumSigner creates a new quantum signer
func NewQuantumSigner(config *SignerConfig) *QuantumSigner {
	signer := &QuantumSigner{
		config:     config,
		metrics:    &SignerMetrics{},
		algorithms: make(map[string]SigningAlgorithm),
	}

	// Register supported algorithms
	signer.registerAlgorithms()

	return signer
}

// RegisterAlgorithms registers supported signing algorithms
func (qs *QuantumSigner) registerAlgorithms() {
	// Register Dilithium (quantum-resistant)
	qs.algorithms["dilithium3"] = &DilithiumAlgorithm{
		keySize: 2048,
		level:   3,
	}

	// Register RSA (classical)
	qs.algorithms["rsa"] = &RSAAlgorithm{
		keySize: qs.config.KeySize,
	}

	// Register hybrid algorithms
	if qs.config.EnableHybrid {
		qs.algorithms["hybrid-rsa-dilithium"] = &HybridAlgorithm{
			classical: qs.algorithms["rsa"],
			quantum:   qs.algorithms["dilithium3"],
		}
	}
}

// Initialize initializes the quantum signer
func (qs *QuantumSigner) Initialize(config map[string]interface{}) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	// Update configuration
	if algorithm, ok := config["algorithm"].(string); ok {
		qs.config.Algorithm = algorithm
	}
	if keySize, ok := config["key_size"].(int); ok {
		qs.config.KeySize = keySize
	}
	if enableQuantum, ok := config["enable_quantum_resist"].(bool); ok {
		qs.config.EnableQuantumResist = enableQuantum
	}

	// Generate key pair if not exists
	if qs.keyPair == nil {
		keyPair, err := qs.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate key pair: %w", err)
		}
		qs.keyPair = keyPair
	}

	return nil
}

// GenerateKeyPair generates a new key pair
func (qs *QuantumSigner) GenerateKeyPair() (*KeyPair, error) {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	algorithm, exists := qs.algorithms[qs.config.Algorithm]
	if !exists {
		return nil, fmt.Errorf("unsupported algorithm: %s", qs.config.Algorithm)
	}

	keyPair, err := algorithm.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	keyPair.Algorithm = qs.config.Algorithm
	keyPair.Created = time.Now()
	keyPair.ID = qs.generateKeyID()

	qs.keyPair = keyPair
	return keyPair, nil
}

// Sign signs data with the private key
func (qs *QuantumSigner) Sign(data []byte) ([]byte, error) {
	start := time.Now()

	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	if qs.keyPair == nil {
		return nil, fmt.Errorf("no key pair available")
	}

	algorithm, exists := qs.algorithms[qs.keyPair.Algorithm]
	if !exists {
		return nil, fmt.Errorf("unsupported algorithm: %s", qs.keyPair.Algorithm)
	}

	signature, err := algorithm.Sign(data, qs.keyPair.PrivateKey)
	if err != nil {
		qs.metrics.ErrorCount++
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	// Update metrics
	elapsed := time.Since(start)
	qs.metrics.SignCount++
	qs.metrics.TotalSignTime += elapsed
	qs.metrics.AvgSignTime = qs.metrics.TotalSignTime / time.Duration(qs.metrics.SignCount)
	qs.metrics.LastSignTime = time.Now()

	return signature, nil
}

// SignWithAlgorithm signs data with a specific algorithm
func (qs *QuantumSigner) SignWithAlgorithm(data []byte, algorithmName string) ([]byte, error) {
	start := time.Now()

	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	algorithm, exists := qs.algorithms[algorithmName]
	if !exists {
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithmName)
	}

	// Generate temporary key pair for this algorithm
	tempKeyPair, err := algorithm.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate temporary key pair: %w", err)
	}

	signature, err := algorithm.Sign(data, tempKeyPair.PrivateKey)
	if err != nil {
		qs.metrics.ErrorCount++
		return nil, fmt.Errorf("signing failed: %w", err)
	}

	// Update metrics
	elapsed := time.Since(start)
	qs.metrics.SignCount++
	qs.metrics.TotalSignTime += elapsed
	qs.metrics.AvgSignTime = qs.metrics.TotalSignTime / time.Duration(qs.metrics.SignCount)
	qs.metrics.LastSignTime = time.Now()

	return signature, nil
}

// Verify verifies a signature
func (qs *QuantumSigner) Verify(data, signature, publicKey []byte) (bool, error) {
	start := time.Now()

	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	// Determine algorithm from public key or use default
	algorithmName := qs.config.Algorithm
	if qs.keyPair != nil {
		algorithmName = qs.keyPair.Algorithm
	}

	algorithm, exists := qs.algorithms[algorithmName]
	if !exists {
		return false, fmt.Errorf("unsupported algorithm: %s", algorithmName)
	}

	valid, err := algorithm.Verify(data, signature, publicKey)
	if err != nil {
		qs.metrics.ErrorCount++
		return false, fmt.Errorf("verification failed: %w", err)
	}

	// Update metrics
	elapsed := time.Since(start)
	qs.metrics.VerifyCount++
	qs.metrics.TotalVerifyTime += elapsed
	qs.metrics.AvgVerifyTime = qs.metrics.TotalVerifyTime / time.Duration(qs.metrics.VerifyCount)
	qs.metrics.LastVerifyTime = time.Now()

	return valid, nil
}

// GetPublicKey returns the public key
func (qs *QuantumSigner) GetPublicKey() ([]byte, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	if qs.keyPair == nil {
		return nil, fmt.Errorf("no key pair available")
	}

	return qs.keyPair.PublicKey, nil
}

// GetPrivateKey returns the encrypted private key
func (qs *QuantumSigner) GetPrivateKey() ([]byte, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	if qs.keyPair == nil {
		return nil, fmt.Errorf("no key pair available")
	}

	if qs.config.EncryptionEnabled {
		// Return encrypted private key
		return qs.encryptPrivateKey(qs.keyPair.PrivateKey)
	}

	return qs.keyPair.PrivateKey, nil
}

// ImportKeyPair imports a key pair
func (qs *QuantumSigner) ImportKeyPair(privateKey, publicKey []byte) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	// Decrypt private key if encrypted
	if qs.config.EncryptionEnabled {
		decrypted, err := qs.decryptPrivateKey(privateKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt private key: %w", err)
		}
		privateKey = decrypted
	}

	qs.keyPair = &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Algorithm:  qs.config.Algorithm,
		Created:    time.Now(),
		ID:         qs.generateKeyID(),
	}

	return nil
}

// ExportKeyPair exports the current key pair
func (qs *QuantumSigner) ExportKeyPair() (*KeyPair, error) {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	if qs.keyPair == nil {
		return nil, fmt.Errorf("no key pair available")
	}

	exported := &KeyPair{
		PrivateKey: make([]byte, len(qs.keyPair.PrivateKey)),
		PublicKey:  make([]byte, len(qs.keyPair.PublicKey)),
		Algorithm:  qs.keyPair.Algorithm,
		Created:    qs.keyPair.Created,
		ID:         qs.keyPair.ID,
	}

	copy(exported.PrivateKey, qs.keyPair.PrivateKey)
	copy(exported.PublicKey, qs.keyPair.PublicKey)

	return exported, nil
}

// GetSupportedAlgorithms returns supported algorithms
func (qs *QuantumSigner) GetSupportedAlgorithms() []string {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	algorithms := make([]string, 0, len(qs.algorithms))
	for name := range qs.algorithms {
		algorithms = append(algorithms, name)
	}

	return algorithms
}

// GetSignerInfo returns signer information
func (qs *QuantumSigner) GetSignerInfo() map[string]string {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	info := map[string]string{
		"algorithm":            qs.config.Algorithm,
		"key_size":             fmt.Sprintf("%d", qs.config.KeySize),
		"hash_algorithm":       qs.config.HashAlgorithm,
		"quantum_resistant":    fmt.Sprintf("%t", qs.config.EnableQuantumResist),
		"hybrid_enabled":       fmt.Sprintf("%t", qs.config.EnableHybrid),
		"encryption_enabled":   fmt.Sprintf("%t", qs.config.EncryptionEnabled),
		"supported_algorithms": fmt.Sprintf("%v", qs.GetSupportedAlgorithms()),
	}

	if qs.keyPair != nil {
		info["key_id"] = qs.keyPair.ID
		info["key_created"] = qs.keyPair.Created.Format(time.RFC3339)
	}

	return info
}

// GetMetrics returns performance metrics
func (qs *QuantumSigner) GetMetrics() map[string]float64 {
	qs.mutex.RLock()
	defer qs.mutex.RUnlock()

	return map[string]float64{
		"sign_count":           float64(qs.metrics.SignCount),
		"verify_count":         float64(qs.metrics.VerifyCount),
		"avg_sign_time_ms":     qs.metrics.AvgSignTime.Seconds() * 1000,
		"avg_verify_time_ms":   qs.metrics.AvgVerifyTime.Seconds() * 1000,
		"error_count":          float64(qs.metrics.ErrorCount),
		"total_sign_time_ms":   qs.metrics.TotalSignTime.Seconds() * 1000,
		"total_verify_time_ms": qs.metrics.TotalVerifyTime.Seconds() * 1000,
	}
}

// ResetMetrics resets performance metrics
func (qs *QuantumSigner) ResetMetrics() {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	qs.metrics = &SignerMetrics{}
}

// UpdateConfig updates signer configuration
func (qs *QuantumSigner) UpdateConfig(config map[string]interface{}) error {
	qs.mutex.Lock()
	defer qs.mutex.Unlock()

	// Update configuration fields
	if algorithm, ok := config["algorithm"].(string); ok {
		qs.config.Algorithm = algorithm
	}
	if keySize, ok := config["key_size"].(int); ok {
		qs.config.KeySize = keySize
	}
	if enableQuantum, ok := config["enable_quantum_resist"].(bool); ok {
		qs.config.EnableQuantumResist = enableQuantum
	}
	if enableHybrid, ok := config["enable_hybrid"].(bool); ok {
		qs.config.EnableHybrid = enableHybrid
	}

	// Re-register algorithms if needed
	qs.registerAlgorithms()

	return nil
}

// GenerateKeyID generates a unique key ID
func (qs *QuantumSigner) generateKeyID() string {
	// Simple hash-based ID generation
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("%x", hash[:8])
}

// EncryptPrivateKey encrypts the private key
func (qs *QuantumSigner) encryptPrivateKey(privateKey []byte) ([]byte, error) {
	// Simple XOR encryption (in production, use proper encryption)
	if qs.config.EncryptionPassword == "" {
		return privateKey, nil
	}

	encrypted := make([]byte, len(privateKey))
	password := []byte(qs.config.EncryptionPassword)

	for i, b := range privateKey {
		encrypted[i] = b ^ password[i%len(password)]
	}

	return encrypted, nil
}

// DecryptPrivateKey decrypts the private key
func (qs *QuantumSigner) decryptPrivateKey(encryptedKey []byte) ([]byte, error) {
	// Simple XOR decryption (in production, use proper decryption)
	if qs.config.EncryptionPassword == "" {
		return encryptedKey, nil
	}

	decrypted := make([]byte, len(encryptedKey))
	password := []byte(qs.config.EncryptionPassword)

	for i, b := range encryptedKey {
		decrypted[i] = b ^ password[i%len(password)]
	}

	return decrypted, nil
}
