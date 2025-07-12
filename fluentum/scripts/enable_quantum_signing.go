// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fluentum-chain/fluentum/core"
	quantum "github.com/fluentum-chain/fluentum/features/quantum_signing"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/features.toml", "Path to features configuration file")
	nodeVersion := flag.String("node-version", "1.0.0", "Fluentum node version")
	enable := flag.Bool("enable", true, "Enable or disable quantum signing")
	mode := flag.Int("mode", 3, "Dilithium mode (2, 3, or 5)")
	enableMetrics := flag.Bool("metrics", true, "Enable performance metrics")
	maxLatencyMs := flag.Int("max-latency", 1000, "Maximum allowed signing latency in milliseconds")

	flag.Parse()

	// Create feature manager
	fm := core.NewFeatureManager(*nodeVersion)

	// Register quantum signing feature
	quantumFeature := quantum.NewQuantumSigningFeature()
	err := fm.RegisterFeature(quantumFeature)
	if err != nil {
		log.Fatalf("Failed to register quantum signing feature: %v", err)
	}

	// Create or update configuration
	config := map[string]interface{}{
		"enabled": *enable,
		"mode":     fmt.Sprintf("Dilithium%d", *mode),
		"quantum_headers": true,
		"enable_metrics": *enableMetrics,
		"max_latency_ms": *maxLatencyMs,
	}

	// Set the configuration
	fm.SetFeatureConfig("quantum_signing", config)

	// Initialize the feature
	err = quantumFeature.Init(config)
	if err != nil {
		log.Fatalf("Failed to initialize quantum signing feature: %v", err)
	}

	// Start the feature
	err = quantumFeature.Start()
	if err != nil {
		log.Fatalf("Failed to start quantum signing feature: %v", err)
	}

	// Get feature status
	status := quantumFeature.GetStatus()
	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal status: %v", err)
	}

	// Print status
	fmt.Println("Quantum Signing Feature Status:")
	fmt.Println(string(statusJSON))

	// Save configuration to file
	err = saveConfig(*configPath, config)
	if err != nil {
		log.Printf("Warning: Failed to save configuration to file: %v", err)
	} else {
		absPath, _ := filepath.Abs(*configPath)
		fmt.Printf("\nConfiguration saved to: %s\n", absPath)
	}

	// Generate sample key pair if enabled
	if *enable {
		fmt.Println("\nGenerating sample key pair...")
		pubKey, privKey, err := quantumFeature.GenerateKey()
		if err != nil {
			log.Printf("Failed to generate sample key pair: %v", err)
		} else {
			fmt.Println("Sample Public Key:", hex.EncodeToString(pubKey))
			fmt.Println("Sample Private Key:", hex.EncodeToString(privKey))
			fmt.Println("\nNote: Store private keys securely and never commit them to version control!")
		}
	}
}

// saveConfig saves the configuration to a TOML file
func saveConfig(path string, config map[string]interface{}) error {
	// Ensure the directory exists
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Write TOML header
	fmt.Fprintf(file, "# Fluentum Feature Configuration\n# Generated at %s\n\n[features.quantum_signing]\n", 
		time.Now().Format(time.RFC3339))

	// Write configuration values
	for key, value := range config {
		switch v := value.(type) {
		case string:
			fmt.Fprintf(file, "%s = \"%s\"\n", key, v)
		case bool:
			fmt.Fprintf(file, "%s = %t\n", key, v)
		case int, int64, float64:
			fmt.Fprintf(file, "%s = %v\n", key, v)
		default:
			// Try to marshal as JSON for complex types
			jsonValue, err := json.Marshal(v)
			if err != nil {
				continue
			}
			fmt.Fprintf(file, "%s = %s\n", key, string(jsonValue))
		}
	}

	return nil
}
