//go:build !plugin
// +build !plugin

package quantum_signing

import (
	"github.com/fluentum-chain/fluentum/core"
)

// Register registers the quantum signing feature with the feature manager
func Register(fm *core.FeatureManager) error {
	// Create a new instance of the quantum signing feature
	feature := NewQuantumSigningFeature()
	
	// Register the feature with the feature manager
	err := fm.RegisterFeature(feature)
	if err != nil {
		return err
	}

	// Default configuration for the quantum signing feature
	config := map[string]interface{}{
		"enabled": true,  // Enable by default
		"mode":     "Dilithium3", // Default security level
	}

	// Set the default configuration
	fm.SetFeatureConfig("quantum_signing", config)
	
	return nil
}
