package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// FeatureConfig represents the configuration for features
type FeatureConfig struct {
	Features struct {
		Enabled            bool `toml:"enabled"`
		AutoReload         bool `toml:"auto_reload"`
		CheckCompatibility bool `toml:"check_compatibility"`

		QuantumSigning struct {
			Enabled        bool `toml:"enabled"`
			DilithiumMode  int  `toml:"dilithium_mode"`
			QuantumHeaders bool `toml:"quantum_headers"`
			EnableMetrics  bool `toml:"enable_metrics"`
			MaxLatencyMs   int  `toml:"max_latency_ms"`
		} `toml:"quantum_signing"`

		StateSync struct {
			Enabled        bool `toml:"enabled"`
			FastSync       bool `toml:"fast_sync"`
			ChunkSize      int  `toml:"chunk_size"`
			MaxConcurrent  int  `toml:"max_concurrent"`
			TimeoutSeconds int  `toml:"timeout_seconds"`
		} `toml:"state_sync"`

		ZKRollup struct {
			Enabled            bool `toml:"enabled"`
			EnableProofs       bool `toml:"enable_proofs"`
			EnableVerification bool `toml:"enable_verification"`
			BatchSize          int  `toml:"batch_size"`
			ProofTimeout       int  `toml:"proof_timeout"`
		} `toml:"zk_rollup"`

		Distribution struct {
			UseGitSubmodules bool   `toml:"use_git_submodules"`
			AutoUpdate       bool   `toml:"auto_update"`
			RepositoryURL    string `toml:"repository_url"`
			Branch           string `toml:"branch"`
		} `toml:"distribution"`

		Compatibility struct {
			MinNodeVersion string `toml:"min_node_version"`
			MaxNodeVersion string `toml:"max_node_version"`
			APIVersion     string `toml:"api_version"`
		} `toml:"compatibility"`
	} `toml:"features"`
}

// FeatureLoader handles loading and managing features
type FeatureLoader struct {
	config         *FeatureConfig
	featureManager *FeatureManager
	configPath     string
}

// NewFeatureLoader creates a new feature loader
func NewFeatureLoader(configPath string, nodeVersion string) *FeatureLoader {
	return &FeatureLoader{
		configPath:     configPath,
		featureManager: NewFeatureManager(nodeVersion),
	}
}

// LoadConfiguration loads the feature configuration from file
func (fl *FeatureLoader) LoadConfiguration() error {
	// Check if config file exists
	if _, err := os.Stat(fl.configPath); os.IsNotExist(err) {
		// Create default configuration
		return fl.createDefaultConfig()
	}

	// Load configuration from file
	configData, err := os.ReadFile(fl.configPath)
	if err != nil {
		return fmt.Errorf("failed to read feature config file: %w", err)
	}

	var config FeatureConfig
	if err := toml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse feature config file: %w", err)
	}

	fl.config = &config
	return nil
}

// createDefaultConfig creates a default configuration file
func (fl *FeatureLoader) createDefaultConfig() error {
	// Ensure directory exists
	configDir := filepath.Dir(fl.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default configuration
	defaultConfig := `# Fluentum Features Configuration
# This file configures the modular features for the Fluentum node

[features]
# Global feature settings
enabled = true
auto_reload = false
check_compatibility = true

# Quantum Signing Feature Configuration
[features.quantum_signing]
enabled = true
dilithium_mode = 3
quantum_headers = true
enable_metrics = true
max_latency_ms = 50

# State Sync Feature Configuration
[features.state_sync]
enabled = false
fast_sync = true
chunk_size = 1000
max_concurrent = 10
timeout_seconds = 30

# ZK Rollup Feature Configuration
[features.zk_rollup]
enabled = false
enable_proofs = true
enable_verification = true
batch_size = 100
proof_timeout = 60

# Feature Distribution Settings
[features.distribution]
use_git_submodules = true
auto_update = false
repository_url = "https://github.com/fluentum-chain/fluentum-features"
branch = "main"

# Feature Version Compatibility
[features.compatibility]
min_node_version = "v0.1.0"
max_node_version = "v1.0.0"
api_version = "v1.0.0"
`

	if err := os.WriteFile(fl.configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create default config file: %w", err)
	}

	// Load the default configuration
	return fl.LoadConfiguration()
}

// InitializeFeatures initializes all features based on configuration
func (fl *FeatureLoader) InitializeFeatures() error {
	if fl.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Load all available features
	if err := fl.featureManager.LoadFeatures(); err != nil {
		return fmt.Errorf("failed to load features: %w", err)
	}

	// Configure features based on loaded configuration
	fl.configureFeatures()

	// Initialize features
	if err := fl.featureManager.InitializeFeatures(); err != nil {
		return fmt.Errorf("failed to initialize features: %w", err)
	}

	return nil
}

// configureFeatures configures features based on the loaded configuration
func (fl *FeatureLoader) configureFeatures() {
	// Configure quantum signing feature
	quantumConfig := map[string]interface{}{
		"enabled":         fl.config.Features.QuantumSigning.Enabled,
		"dilithium_mode":  fl.config.Features.QuantumSigning.DilithiumMode,
		"quantum_headers": fl.config.Features.QuantumSigning.QuantumHeaders,
		"enable_metrics":  fl.config.Features.QuantumSigning.EnableMetrics,
		"max_latency_ms":  fl.config.Features.QuantumSigning.MaxLatencyMs,
	}
	fl.featureManager.SetFeatureConfig("quantum_signing", quantumConfig)

	// Configure state sync feature
	stateSyncConfig := map[string]interface{}{
		"enabled":         fl.config.Features.StateSync.Enabled,
		"fast_sync":       fl.config.Features.StateSync.FastSync,
		"chunk_size":      fl.config.Features.StateSync.ChunkSize,
		"max_concurrent":  fl.config.Features.StateSync.MaxConcurrent,
		"timeout_seconds": fl.config.Features.StateSync.TimeoutSeconds,
	}
	fl.featureManager.SetFeatureConfig("state_sync", stateSyncConfig)

	// Configure ZK rollup feature
	zkRollupConfig := map[string]interface{}{
		"enabled":             fl.config.Features.ZKRollup.Enabled,
		"enable_proofs":       fl.config.Features.ZKRollup.EnableProofs,
		"enable_verification": fl.config.Features.ZKRollup.EnableVerification,
		"batch_size":          fl.config.Features.ZKRollup.BatchSize,
		"proof_timeout":       fl.config.Features.ZKRollup.ProofTimeout,
	}
	fl.featureManager.SetFeatureConfig("zk_rollup", zkRollupConfig)
}

// StartFeatures starts all enabled features
func (fl *FeatureLoader) StartFeatures() error {
	return fl.featureManager.StartFeatures()
}

// StopFeatures stops all features
func (fl *FeatureLoader) StopFeatures() error {
	return fl.featureManager.StopFeatures()
}

// ReloadFeatures reloads all features
func (fl *FeatureLoader) ReloadFeatures() error {
	// Reload configuration
	if err := fl.LoadConfiguration(); err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Reconfigure features
	fl.configureFeatures()

	// Reload all features
	return fl.featureManager.ReloadAllFeatures()
}

// GetFeatureManager returns the feature manager
func (fl *FeatureLoader) GetFeatureManager() *FeatureManager {
	return fl.featureManager
}

// GetFeatureStatus returns the status of all features
func (fl *FeatureLoader) GetFeatureStatus() map[string]interface{} {
	return fl.featureManager.GetFeatureStatus()
}

// ValidateConfiguration validates the feature configuration
func (fl *FeatureLoader) ValidateConfiguration() error {
	if fl.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate quantum signing configuration
	if fl.config.Features.QuantumSigning.Enabled {
		mode := fl.config.Features.QuantumSigning.DilithiumMode
		if mode != 1 && mode != 3 && mode != 5 {
			return fmt.Errorf("invalid dilithium mode: %d (must be 1, 3, or 5)", mode)
		}

		if fl.config.Features.QuantumSigning.MaxLatencyMs <= 0 {
			return fmt.Errorf("invalid max latency: %d ms (must be positive)", fl.config.Features.QuantumSigning.MaxLatencyMs)
		}
	}

	// Validate state sync configuration
	if fl.config.Features.StateSync.Enabled {
		if fl.config.Features.StateSync.ChunkSize <= 0 {
			return fmt.Errorf("invalid chunk size: %d (must be positive)", fl.config.Features.StateSync.ChunkSize)
		}

		if fl.config.Features.StateSync.MaxConcurrent <= 0 {
			return fmt.Errorf("invalid max concurrent: %d (must be positive)", fl.config.Features.StateSync.MaxConcurrent)
		}

		if fl.config.Features.StateSync.TimeoutSeconds <= 0 {
			return fmt.Errorf("invalid timeout: %d seconds (must be positive)", fl.config.Features.StateSync.TimeoutSeconds)
		}
	}

	// Validate ZK rollup configuration
	if fl.config.Features.ZKRollup.Enabled {
		if fl.config.Features.ZKRollup.BatchSize <= 0 {
			return fmt.Errorf("invalid batch size: %d (must be positive)", fl.config.Features.ZKRollup.BatchSize)
		}

		if fl.config.Features.ZKRollup.ProofTimeout <= 0 {
			return fmt.Errorf("invalid proof timeout: %d seconds (must be positive)", fl.config.Features.ZKRollup.ProofTimeout)
		}
	}

	return nil
}
