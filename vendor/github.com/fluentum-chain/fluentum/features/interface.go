package features

// No imports needed as we use basic Go types

// FeatureInterface defines the interface that all features must implement
type FeatureInterface interface {
	// Initialize initializes the feature with the given config
	Initialize(config map[string]interface{}) error

	// Name returns the name of the feature
	Name() string

	// Version returns the version of the feature
	Version() string

	// Description returns a description of the feature
	Description() string

	// IsEnabled returns whether the feature is enabled
	IsEnabled() bool

	// SetEnabled enables or disables the feature
	SetEnabled(enabled bool)

	// Close cleans up resources used by the feature
	Close() error
}

// BaseFeature provides a default implementation of common FeatureInterface methods
type BaseFeature struct {
	name        string
	version     string
	description string
	enabled     bool
}

// NewBaseFeature creates a new BaseFeature
func NewBaseFeature(name, version, description string) *BaseFeature {
	return &BaseFeature{
		name:        name,
		version:     version,
		description: description,
		enabled:     true, // Features are enabled by default
	}
}

// Initialize implements FeatureInterface
func (f *BaseFeature) Initialize(cfg interface{}) error {
	return nil
}

// Name returns the name of the feature
func (f *BaseFeature) Name() string {
	return f.name
}

// Version returns the version of the feature
func (f *BaseFeature) Version() string {
	return f.version
}

// Description returns a description of the feature
func (f *BaseFeature) Description() string {
	return f.description
}

// IsEnabled returns whether the feature is enabled
func (f *BaseFeature) IsEnabled() bool {
	return f.enabled
}

// SetEnabled enables or disables the feature
func (f *BaseFeature) SetEnabled(enabled bool) {
	f.enabled = enabled
}

// Close cleans up resources used by the feature
func (f *BaseFeature) Close() error {
	// Default implementation does nothing
	return nil
}

// QMoEValidator defines the interface for QMoE validator features
type QMoEValidator interface {
	FeatureInterface

	// ValidateBatch validates a batch of transactions
	ValidateBatch(batch interface{}) (bool, error)

	// GetMetrics returns metrics about the validator
	GetMetrics() map[string]interface{}
}

// QuantumSigner defines the interface for quantum signer features
type QuantumSigner interface {
	FeatureInterface

	// Sign signs the given data
	Sign(data []byte) ([]byte, error)

	// Verify verifies the signature for the given data
	Verify(data, signature []byte) (bool, error)

	// GetPublicKey returns the public key
	GetPublicKey() ([]byte, error)
}

// FeatureFactory is a function that creates a new feature
type FeatureFactory func() (FeatureInterface, error)

// FeatureDescriptor describes a feature
// This is used in the plugin's main package
var Feature FeatureInterface

// RegisterFeature registers a feature factory
func RegisterFeature(factory FeatureFactory) {
	// This function is implemented by the feature manager
	// and will be called by the plugin's init() function
}
