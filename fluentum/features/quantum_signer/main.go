package main

import (
	"github.com/fluentum-chain/fluentum/features"
	"github.com/fluentum-chain/fluentum/features/quantum_signer"
)

// This file serves as the entry point for the Quantum Signer plugin.
// It exports the Feature symbol that will be used by the feature manager.

// Feature is the exported symbol that will be used by the feature manager to load this plugin.
// The feature manager will call the Initialize method with the configuration.
var Feature = quantum_signer.New(nil, &features.QuantumSignerConfig{})

// This init function registers the feature with the feature manager.
// It's called when the plugin is loaded.
func init() {
	// The feature manager will replace this function with its own implementation
	// that properly registers the feature.
	if features.RegisterFeature != nil {
		features.RegisterFeature(func() (features.FeatureInterface, error) {
			return Feature, nil
		})
	}
}

// main is required for the plugin to be built as a shared library.
// It's not used when the plugin is loaded by the feature manager.
func main() {}
