package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fluentum-chain/fluentum/version"
)

// QuantumManifest represents the manifest.json structure for quantum plugins
type QuantumManifest struct {
	Name           string   `json:"name"`
	MinCoreVersion string   `json:"min_core_version"`
	APIVersion     string   `json:"api_version"`
	Checksum       string   `json:"checksum"`
	SupportedModes []string `json:"supported_modes"`
}

// VerifyQuantumCompatibility verifies that a quantum plugin is compatible with the current core version
func VerifyQuantumCompatibility(manifestPath string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var mf QuantumManifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return err
	}

	if !version.Compatible(mf.MinCoreVersion) {
		return fmt.Errorf("requires core version >= %s", mf.MinCoreVersion)
	}

	return nil
}

// VerifyQuantumMode checks if the specified mode is supported by the plugin
func VerifyQuantumMode(manifestPath, mode string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var mf QuantumManifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return err
	}

	for _, supportedMode := range mf.SupportedModes {
		if strings.EqualFold(supportedMode, mode) {
			return nil
		}
	}

	return fmt.Errorf("mode '%s' not supported by plugin. Supported modes: %v", mode, mf.SupportedModes)
}

// GetQuantumManifest loads and returns the quantum manifest from a file
func GetQuantumManifest(manifestPath string) (*QuantumManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var mf QuantumManifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, err
	}

	return &mf, nil
}

// ValidateQuantumPlugin performs comprehensive validation of a quantum plugin
func ValidateQuantumPlugin(manifestPath, libPath string) error {
	// Verify manifest exists and is valid
	manifest, err := GetQuantumManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("invalid manifest: %v", err)
	}

	// Check version compatibility
	if err := VerifyQuantumCompatibility(manifestPath); err != nil {
		return err
	}

	// Verify library file exists
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin library not found: %s", libPath)
	}

	// Validate manifest fields
	if manifest.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	if manifest.APIVersion == "" {
		return fmt.Errorf("manifest missing required field: api_version")
	}
	if len(manifest.SupportedModes) == 0 {
		return fmt.Errorf("manifest missing required field: supported_modes")
	}

	return nil
}
