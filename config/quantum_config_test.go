package config

import (
	"testing"
)

func TestQuantumConfigValidation(t *testing.T) {
	// Test default quantum config
	quantumConfig := DefaultQuantumConfig()
	if quantumConfig.Enabled != false {
		t.Errorf("Expected default quantum enabled to be false, got %v", quantumConfig.Enabled)
	}
	if quantumConfig.LibPath != "/usr/local/lib/fluentum/quantum.so" {
		t.Errorf("Expected default quantum lib_path to be /usr/local/lib/fluentum/quantum.so, got %s", quantumConfig.LibPath)
	}
	if quantumConfig.Mode != "dilithium3" {
		t.Errorf("Expected default quantum mode to be dilithium3, got %s", quantumConfig.Mode)
	}

	// Test validation
	err := quantumConfig.ValidateBasic()
	if err != nil {
		t.Errorf("Expected no validation error for default config, got %v", err)
	}

	// Test enabled config with missing lib_path
	quantumConfig.Enabled = true
	quantumConfig.LibPath = ""
	err = quantumConfig.ValidateBasic()
	if err == nil {
		t.Error("Expected validation error for enabled quantum with missing lib_path")
	}

	// Test enabled config with missing mode
	quantumConfig.LibPath = "/usr/local/lib/fluentum/quantum.so"
	quantumConfig.Mode = ""
	err = quantumConfig.ValidateBasic()
	if err == nil {
		t.Error("Expected validation error for enabled quantum with missing mode")
	}

	// Test valid enabled config
	quantumConfig.Mode = "dilithium3"
	err = quantumConfig.ValidateBasic()
	if err != nil {
		t.Errorf("Expected no validation error for valid enabled config, got %v", err)
	}
}

func TestConsensusConfigSignatureScheme(t *testing.T) {
	// Test default consensus config
	consensusConfig := DefaultConsensusConfig()
	if consensusConfig.SignatureScheme != "ecdsa" {
		t.Errorf("Expected default signature_scheme to be ecdsa, got %s", consensusConfig.SignatureScheme)
	}

	// Test test consensus config
	testConsensusConfig := TestConsensusConfig()
	if testConsensusConfig.SignatureScheme != "ecdsa" {
		t.Errorf("Expected test signature_scheme to be ecdsa, got %s", testConsensusConfig.SignatureScheme)
	}
}

func TestConfigWithQuantum(t *testing.T) {
	// Test that the main config includes quantum config
	config := DefaultConfig()
	if config.Quantum == nil {
		t.Error("Expected config to include quantum configuration")
	}

	// Test validation of main config with quantum
	err := config.ValidateBasic()
	if err != nil {
		t.Errorf("Expected no validation error for default config, got %v", err)
	}

	// Test validation with enabled quantum
	config.Quantum.Enabled = true
	config.Quantum.LibPath = "/usr/local/lib/fluentum/quantum.so"
	config.Quantum.Mode = "dilithium3"
	err = config.ValidateBasic()
	if err != nil {
		t.Errorf("Expected no validation error for valid quantum config, got %v", err)
	}
}
