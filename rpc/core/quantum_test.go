package core

import (
	"testing"

	cfg "github.com/fluentum-chain/fluentum/config"
	rpctypes "github.com/fluentum-chain/fluentum/rpc/jsonrpc/types"
)

func TestQuantumAPI_Status(t *testing.T) {
	api := &QuantumAPI{}
	ctx := &rpctypes.Context{}

	// Test with no active signer
	result, err := api.Status(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Enabled {
		t.Error("Expected disabled when no signer is active")
	}
	if result.SignerName != "none" {
		t.Errorf("Expected signer name 'none', got '%s'", result.SignerName)
	}
}

func TestQuantumAPI_Reload(t *testing.T) {
	api := &QuantumAPI{}
	ctx := &rpctypes.Context{}

	// Set up a mock environment
	origEnv := env
	defer func() { env = origEnv }()

	// Create a minimal environment for testing
	env = &Environment{
		Config: cfg.RPCConfig{}, // Non-empty config
	}

	// Test reload with default config (should fail since lib path doesn't exist)
	result, err := api.Reload(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should fail because the quantum library doesn't exist
	if result.Success {
		t.Error("Expected reload to fail with non-existent library")
	}
	if result.Error == "" {
		t.Error("Expected error message for failed reload")
	}
}

func TestQuantumReload_RouteFunction(t *testing.T) {
	ctx := &rpctypes.Context{}

	// Set up a mock environment
	origEnv := env
	defer func() { env = origEnv }()

	env = &Environment{
		Config: cfg.RPCConfig{},
	}

	// Test the route function
	result, err := QuantumReload(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Success {
		t.Error("Expected reload to fail with non-existent library")
	}
}

func TestQuantumStatus_RouteFunction(t *testing.T) {
	ctx := &rpctypes.Context{}

	// Test the route function
	result, err := QuantumStatus(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Enabled {
		t.Error("Expected disabled when no signer is active")
	}
}

func TestQuantumAPI_Integration(t *testing.T) {
	// Test that the API can be created and methods can be called
	api := &QuantumAPI{}
	if api == nil {
		t.Fatal("Failed to create QuantumAPI")
	}

	ctx := &rpctypes.Context{}

	// Test status method
	status, err := api.Status(ctx)
	if err != nil {
		t.Fatalf("Status method failed: %v", err)
	}

	// Verify status structure
	if status == nil {
		t.Fatal("Status result is nil")
	}

	// Test that we can check the enabled field
	_ = status.Enabled
	_ = status.SignerName
	_ = status.Mode
}
