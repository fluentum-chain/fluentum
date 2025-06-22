package core

import (
	"fmt"
	"sync"

	"../features/quantum_signing"
	"../features/state_sync"
	"../features/zk_rollup"
)

// Feature interface that all features must implement
type Feature interface {
	Name() string
	Version() string
	Init(config map[string]interface{}) error
	Start() error
	Stop() error
	Reload() error
	CheckCompatibility(nodeVersion string) error
	IsEnabled() bool
}

// FeatureManager manages all features in the Fluentum node
type FeatureManager struct {
	features    map[string]Feature
	config      map[string]map[string]interface{}
	nodeVersion string
	mu          sync.RWMutex
}

// NewFeatureManager creates a new feature manager
func NewFeatureManager(nodeVersion string) *FeatureManager {
	return &FeatureManager{
		features:    make(map[string]Feature),
		config:      make(map[string]map[string]interface{}),
		nodeVersion: nodeVersion,
	}
}

// RegisterFeature registers a feature with the manager
func (fm *FeatureManager) RegisterFeature(feature Feature) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	name := feature.Name()
	if _, exists := fm.features[name]; exists {
		return fmt.Errorf("feature %s already registered", name)
	}

	fm.features[name] = feature
	return nil
}

// LoadFeatures loads all available features
func (fm *FeatureManager) LoadFeatures() error {
	// Register quantum signing feature
	quantumFeature := quantum_signing.NewQuantumSigningFeature()
	if err := fm.RegisterFeature(quantumFeature); err != nil {
		return fmt.Errorf("failed to register quantum signing feature: %w", err)
	}

	// Register state sync feature
	stateSyncFeature := state_sync.NewStateSyncFeature()
	if err := fm.RegisterFeature(stateSyncFeature); err != nil {
		return fmt.Errorf("failed to register state sync feature: %w", err)
	}

	// Register ZK rollup feature
	zkRollupFeature := zk_rollup.NewZKRollupFeature()
	if err := fm.RegisterFeature(zkRollupFeature); err != nil {
		return fmt.Errorf("failed to register ZK rollup feature: %w", err)
	}

	return nil
}

// InitializeFeatures initializes all features with their configurations
func (fm *FeatureManager) InitializeFeatures() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for name, feature := range fm.features {
		config := fm.config[name]
		if config == nil {
			config = make(map[string]interface{})
		}

		// Check compatibility
		if err := feature.CheckCompatibility(fm.nodeVersion); err != nil {
			return fmt.Errorf("feature %s compatibility check failed: %w", name, err)
		}

		// Initialize the feature
		if err := feature.Init(config); err != nil {
			return fmt.Errorf("failed to initialize feature %s: %w", name, err)
		}
	}

	return nil
}

// StartFeatures starts all enabled features
func (fm *FeatureManager) StartFeatures() error {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for name, feature := range fm.features {
		if feature.IsEnabled() {
			if err := feature.Start(); err != nil {
				return fmt.Errorf("failed to start feature %s: %w", name, err)
			}
		}
	}

	return nil
}

// StopFeatures stops all features
func (fm *FeatureManager) StopFeatures() error {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for name, feature := range fm.features {
		if err := feature.Stop(); err != nil {
			return fmt.Errorf("failed to stop feature %s: %w", name, err)
		}
	}

	return nil
}

// ReloadFeature reloads a specific feature
func (fm *FeatureManager) ReloadFeature(name string) error {
	fm.mu.RLock()
	feature, exists := fm.features[name]
	fm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("feature %s not found", name)
	}

	return feature.Reload()
}

// ReloadAllFeatures reloads all features
func (fm *FeatureManager) ReloadAllFeatures() error {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for name, feature := range fm.features {
		if err := feature.Reload(); err != nil {
			return fmt.Errorf("failed to reload feature %s: %w", name, err)
		}
	}

	return nil
}

// GetFeature returns a feature by name
func (fm *FeatureManager) GetFeature(name string) (Feature, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	feature, exists := fm.features[name]
	return feature, exists
}

// SetFeatureConfig sets the configuration for a feature
func (fm *FeatureManager) SetFeatureConfig(name string, config map[string]interface{}) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.config[name] = config
}

// GetFeatureStatus returns the status of all features
func (fm *FeatureManager) GetFeatureStatus() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	status := make(map[string]interface{})
	for name, feature := range fm.features {
		status[name] = map[string]interface{}{
			"enabled": feature.IsEnabled(),
			"version": feature.Version(),
		}
	}

	return status
}

// ListFeatures returns a list of all registered features
func (fm *FeatureManager) ListFeatures() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	features := make([]string, 0, len(fm.features))
	for name := range fm.features {
		features = append(features, name)
	}

	return features
}

// EnableFeature enables a specific feature
func (fm *FeatureManager) EnableFeature(name string) error {
	fm.mu.RLock()
	feature, exists := fm.features[name]
	fm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("feature %s not found", name)
	}

	// Update config to enable the feature
	config := fm.config[name]
	if config == nil {
		config = make(map[string]interface{})
	}
	config["enabled"] = true
	fm.SetFeatureConfig(name, config)

	// Reinitialize the feature
	return feature.Init(config)
}

// DisableFeature disables a specific feature
func (fm *FeatureManager) DisableFeature(name string) error {
	fm.mu.RLock()
	feature, exists := fm.features[name]
	fm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("feature %s not found", name)
	}

	// Update config to disable the feature
	config := fm.config[name]
	if config == nil {
		config = make(map[string]interface{})
	}
	config["enabled"] = false
	fm.SetFeatureConfig(name, config)

	// Reinitialize the feature
	return feature.Init(config)
}
