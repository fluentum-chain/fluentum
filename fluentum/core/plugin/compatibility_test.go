package plugin

import (
	"os"
	"testing"

	"github.com/fluentum-chain/fluentum/fluentum/version"
)

func TestQuantumManifestParsing(t *testing.T) {
	// Test valid manifest
	validManifest := `{
		"name": "quantum_signing",
		"min_core_version": "v0.6.0",
		"api_version": "1.0",
		"checksum": "sha256:abc123...",
		"supported_modes": ["dilithium2", "dilithium3", "dilithium5"]
	}`

	tmpFile, err := os.CreateTemp("", "manifest_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validManifest); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	manifest, err := GetQuantumManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse valid manifest: %v", err)
	}

	if manifest.Name != "quantum_signing" {
		t.Errorf("Expected name 'quantum_signing', got '%s'", manifest.Name)
	}
	if manifest.MinCoreVersion != "v0.6.0" {
		t.Errorf("Expected min_core_version 'v0.6.0', got '%s'", manifest.MinCoreVersion)
	}
	if manifest.APIVersion != "1.0" {
		t.Errorf("Expected api_version '1.0', got '%s'", manifest.APIVersion)
	}
	if len(manifest.SupportedModes) != 3 {
		t.Errorf("Expected 3 supported modes, got %d", len(manifest.SupportedModes))
	}
}

func TestVerifyQuantumMode(t *testing.T) {
	validManifest := `{
		"name": "quantum_signing",
		"min_core_version": "v0.6.0",
		"api_version": "1.0",
		"checksum": "sha256:abc123...",
		"supported_modes": ["dilithium2", "dilithium3", "dilithium5"]
	}`

	tmpFile, err := os.CreateTemp("", "manifest_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validManifest); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Test supported modes
	supportedModes := []string{"dilithium2", "dilithium3", "dilithium5"}
	for _, mode := range supportedModes {
		if err := VerifyQuantumMode(tmpFile.Name(), mode); err != nil {
			t.Errorf("Mode '%s' should be supported: %v", mode, err)
		}
	}

	// Test case insensitive
	if err := VerifyQuantumMode(tmpFile.Name(), "DILITHIUM3"); err != nil {
		t.Errorf("Mode 'DILITHIUM3' should be supported (case insensitive): %v", err)
	}

	// Test unsupported mode
	if err := VerifyQuantumMode(tmpFile.Name(), "unsupported_mode"); err == nil {
		t.Error("Expected error for unsupported mode")
	}
}

func TestValidateQuantumPlugin(t *testing.T) {
	// Save and set version.Version for compatibility
	origVersion := version.Version
	version.Version = "v0.6.0"
	defer func() { version.Version = origVersion }()

	validManifest := `{
		"name": "quantum_signing",
		"min_core_version": "v0.6.0",
		"api_version": "1.0",
		"checksum": "sha256:abc123...",
		"supported_modes": ["dilithium2", "dilithium3", "dilithium5"]
	}`

	tmpFile, err := os.CreateTemp("", "manifest_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validManifest); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Create a dummy library file
	libFile, err := os.CreateTemp("", "lib_*.so")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFile.Name())
	libFile.Close()

	// Test valid plugin
	if err := ValidateQuantumPlugin(tmpFile.Name(), libFile.Name()); err != nil {
		t.Errorf("Expected no error for valid plugin: %v", err)
	}

	// Test missing library file
	if err := ValidateQuantumPlugin(tmpFile.Name(), "/nonexistent/lib.so"); err == nil {
		t.Error("Expected error for missing library file")
	}
}

func TestInvalidManifest(t *testing.T) {
	// Test invalid JSON
	invalidManifest := `{
		"name": "quantum_signing",
		"min_core_version": "v0.6.0",
		"api_version": "1.0",
		"checksum": "sha256:abc123...",
		"supported_modes": ["dilithium2", "dilithium3", "dilithium5"
	}`

	tmpFile, err := os.CreateTemp("", "manifest_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidManifest); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if _, err := GetQuantumManifest(tmpFile.Name()); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
