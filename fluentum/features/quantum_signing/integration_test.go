//go:build integration
// +build integration

package quantum_signing_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fluentum-chain/fluentum/core"
	quantum "github.com/fluentum-chain/fluentum/features/quantum_signing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuantumSigningIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a new feature manager
	nodeVersion := "1.0.0"
	fm := core.NewFeatureManager(nodeVersion)

	// Register the quantum signing feature
	quantumFeature := quantum.NewQuantumSigningFeature()
	err := fm.RegisterFeature(quantumFeature)
	require.NoError(t, err, "Failed to register quantum signing feature")

	// Initialize with test configuration
	config := map[string]interface{}{
		"enabled": true,
		"mode":      "Dilithium3",
	}
	err = quantumFeature.Init(config)
	require.NoError(t, err, "Failed to initialize quantum signing feature")

	// Start the feature
	err = quantumFeature.Start()
	require.NoError(t, err, "Failed to start quantum signing feature")

	// Test key generation
	pubKey, privKey, err := quantumFeature.GenerateKey()
	require.NoError(t, err, "Failed to generate key pair")
	assert.NotEmpty(t, pubKey, "Public key should not be empty")
	assert.NotEmpty(t, privKey, "Private key should not be empty")

	// Test signing and verification
	message := []byte("test message")
	signature, err := quantumFeature.Sign(privKey, message)
	require.NoError(t, err, "Failed to sign message")
	assert.NotEmpty(t, signature, "Signature should not be empty")

	// Verify the signature
	valid, err := quantumFeature.Verify(pubKey, message, signature)
	require.NoError(t, err, "Failed to verify signature")
	assert.True(t, valid, "Signature should be valid")

	// Test with a different message (should not verify)
	wrongMessage := []byte("wrong message")
	valid, err = quantumFeature.Verify(pubKey, wrongMessage, signature)
	require.NoError(t, err, "Verification should not error with wrong message")
	assert.False(t, valid, "Signature should not be valid for wrong message")

	// Test public key derivation
	derivedPubKey, err := quantumFeature.PublicKey(privKey)
	require.NoError(t, err, "Failed to derive public key")
	assert.Equal(t, pubKey, derivedPubKey, "Derived public key should match generated public key")

	// Test block signing and verification
	t.Run("Block Signing and Verification", func(t *testing.T) {
		blockData := []byte("block data to sign")
		signature, err := quantumFeature.Sign(privKey, blockData)
		require.NoError(t, err, "Failed to sign block data")

		valid, err := quantumFeature.Verify(pubKey, blockData, signature)
		require.NoError(t, err, "Failed to verify block signature")
		assert.True(t, valid, "Block signature should be valid")
	})

	// Test performance with larger data
	t.Run("Performance Test", func(t *testing.T) {
		largeData := make([]byte, 1024*1024) // 1MB of data
		_, err := rand.Read(largeData)
		require.NoError(t, err, "Failed to generate random data")

		start := time.Now()
		signature, err := quantumFeature.Sign(privKey, largeData)
		signTime := time.Since(start)
		require.NoError(t, err, "Failed to sign large data")

		start = time.Now()
		valid, err := quantumFeature.Verify(pubKey, largeData, signature)
		verifyTime := time.Since(start)

		require.NoError(t, err, "Failed to verify large data signature")
		assert.True(t, valid, "Large data signature should be valid")

		t.Logf("Signed 1MB in %v, verified in %v", signTime, verifyTime)
	})

	// Test feature status
	t.Run("Feature Status", func(t *testing.T) {
		status := quantumFeature.GetStatus()
		assert.NotNil(t, status, "Status should not be nil")
		t.Logf("Feature status: %+v", status)

		// Check that the feature reports as enabled
		assert.True(t, quantumFeature.IsEnabled(), "Feature should be enabled")
	})

	// Stop the feature
	err = quantumFeature.Stop()
	require.NoError(t, err, "Failed to stop quantum signing feature")
}

// TestQuantumSignerPlugin tests the quantum signer plugin implementation
func TestQuantumSignerPlugin(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a new quantum signer
	signer, err := quantum.NewDilithiumSigner()
	require.NoError(t, err, "Failed to create quantum signer")

	// Test key generation
	pubKey, privKey, err := signer.GenerateKey()
	require.NoError(t, err, "Failed to generate key pair")
	assert.NotEmpty(t, pubKey, "Public key should not be empty")
	assert.NotEmpty(t, privKey, "Private key should not be empty")

	// Test signing and verification
	message := []byte("test message for plugin")
	signature, err := signer.Sign(privKey, message)
	require.NoError(t, err, "Failed to sign message with plugin")
	assert.NotEmpty(t, signature, "Signature should not be empty")

	// Verify the signature
	valid, err := signer.Verify(pubKey, message, signature)
	require.NoError(t, err, "Failed to verify signature with plugin")
	assert.True(t, valid, "Signature should be valid")

	// Test public key derivation
	derivedPubKey, err := signer.PublicKey(privKey)
	require.NoError(t, err, "Failed to derive public key with plugin")
	assert.Equal(t, pubKey, derivedPubKey, "Derived public key should match generated public key")
}

// TestMain handles setup and teardown for integration tests
func TestMain(m *testing.M) {
	// Add any setup code here
	fmt.Println("Setting up integration tests...")

	// Run tests
	code := m.Run()

	// Add any teardown code here
	fmt.Println("Tearing down integration tests...")

	os.Exit(code)
}
