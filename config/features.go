package config

import "time"

// FeaturesConfig holds configuration for the feature manager
type FeaturesConfig struct {
	// Enabled is a list of enabled features
	Enabled []string `mapstructure:"enabled"`

	// VersionRequirements specifies version requirements for features
	VersionRequirements []string `mapstructure:"version_requirements"`

	// AutoUpdate enables automatic updates for features
	AutoUpdate bool `mapstructure:"auto_update"`

	// UpdateCheckInterval specifies how often to check for updates
	UpdateCheckInterval time.Duration `mapstructure:"update_check_interval"`

	// Registry configuration
	Registry RegistryConfig `mapstructure:"registry"`

	// Loader configuration
	Loader LoaderConfig `mapstructure:"loader"`

	// Metrics configuration
	Metrics MetricsConfig `mapstructure:"metrics"`

	// QMoE Validator specific configuration
	QMoEValidator QMoEConfig `mapstructure:"qmoe_validator"`

	// Quantum Signer specific configuration
	QuantumSigner QuantumSignerConfig `mapstructure:"quantum_signer"`
}

// RegistryConfig holds configuration for the feature registry
type RegistryConfig struct {
	// LocalPath is the path where features are stored locally
	LocalPath string `mapstructure:"local_path"`

	// RemoteRegistry is the URL of the remote feature registry
	RemoteRegistry string `mapstructure:"remote_registry"`

	// InsecureSkipVerify skips TLS verification for the remote registry
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

// LoaderConfig holds configuration for the feature loader
type LoaderConfig struct {
	// MaxConcurrentLoads is the maximum number of features to load concurrently
	MaxConcurrentLoads int `mapstructure:"max_concurrent_loads"`

	// LoadTimeout is the maximum time to wait for a feature to load
	LoadTimeout time.Duration `mapstructure:"load_timeout"`

	// IsolationEnabled enables process isolation for features
	IsolationEnabled bool `mapstructure:"isolation_enabled"`
}

// MetricsConfig holds configuration for feature metrics
type MetricsConfig struct {
	// Enabled enables feature metrics collection
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the address to serve metrics on
	Endpoint string `mapstructure:"endpoint"`

	// CollectionInterval is how often to collect metrics
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
}

// QMoEConfig holds configuration for the QMoE validator
type QMoEConfig struct {
	// Enabled enables the QMoE validator
	Enabled bool `mapstructure:"enabled"`

	// Quantization enables quantization
	Quantization bool `mapstructure:"quantization"`

	// SparseActivation enables sparse activation
	SparseActivation bool `mapstructure:"sparse_activation"`

	// NumExperts is the number of experts in the QMoE model
	NumExperts int `mapstructure:"num_experts"`

	// ConfidenceThreshold is the minimum confidence threshold for batch validation
	ConfidenceThreshold float64 `mapstructure:"confidence_threshold"`

	// GasSavingsThreshold is the minimum gas savings threshold for batch optimization
	GasSavingsThreshold float64 `mapstructure:"gas_savings_threshold"`

	// ModelPath is the path to the QMoE model file
	ModelPath string `mapstructure:"model_path"`
}

// QuantumSignerConfig holds configuration for the quantum signer
type QuantumSignerConfig struct {
	// Enabled enables the quantum signer
	Enabled bool `mapstructure:"enabled"`

	// KeyType is the type of quantum-safe key to use (e.g., "dilithium3")
	KeyType string `mapstructure:"key_type"`

	// KeyPath is the path to the quantum key file
	KeyPath string `mapstructure:"key_path"`

	// SigningAlgorithm is the quantum-safe signing algorithm to use
	SigningAlgorithm string `mapstructure:"signing_algorithm"`

	// KeySize is the size of the quantum key in bits
	KeySize int `mapstructure:"key_size"`
}

// DefaultFeaturesConfig returns a default configuration for the feature manager
func DefaultFeaturesConfig() *FeaturesConfig {
	return &FeaturesConfig{
		Enabled:             []string{},
		VersionRequirements: []string{},
		AutoUpdate:          true,
		UpdateCheckInterval: 24 * time.Hour,
		Registry: RegistryConfig{
			LocalPath:          "$HOME/.fluentum/features",
			RemoteRegistry:     "https://features.fluentum.xyz",
			InsecureSkipVerify: false,
		},
		Loader: LoaderConfig{
			MaxConcurrentLoads: 5,
			LoadTimeout:        30 * time.Second,
			IsolationEnabled:   true,
		},
		Metrics: MetricsConfig{
			Enabled:            true,
			Endpoint:           ":9090",
			CollectionInterval: time.Minute,
		},
		QMoEValidator: QMoEConfig{
			Enabled:             false,
			Quantization:        true,
			SparseActivation:    true,
			NumExperts:          8,
			ConfidenceThreshold: 0.7,
			GasSavingsThreshold: 0.3,
			ModelPath:           "$HOME/.fluentum/models/qmoe.bin",
		},
		QuantumSigner: QuantumSignerConfig{
			Enabled:          false,
			KeyType:          "dilithium3",
			KeyPath:          "$HOME/.fluentum/keys/quantum",
			SigningAlgorithm: "dilithium",
			KeySize:          2048,
		},
	}
}
