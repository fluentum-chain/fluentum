package quantum_signing_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fluentum/quantum_signing"
	"github.com/stretchr/testify/require"
)

func TestQuantumSigningExample(t *testing.T) {
	// Create a new quantum signer instance
	signer, err := quantum_signing.NewDilithiumSigner()
	require.NoError(t, err, "Failed to create signer")

	// Generate a new key pair
	pubKey, privKey, err := signer.GenerateKey()
	require.NoError(t, err, "Failed to generate key pair")

	// Test message to sign
	message := []byte("Test message for quantum signing")

	// Sign the message
	startSign := time.Now()
	signature, err := signer.Sign(privKey, message)
	signTime := time.Since(startSign)

	require.NoError(t, err, "Failed to sign message")
	require.NotEmpty(t, signature, "Signature should not be empty")

	// Verify the signature
	startVerify := time.Now()
	valid, err := signer.Verify(pubKey, message, signature)
	verifyTime := time.Since(startVerify)

	require.NoError(t, err, "Verification failed")
	require.True(t, valid, "Signature should be valid")

	// Print timing information
	fmt.Printf("Sign time: %v\n", signTime)
	fmt.Printf("Verify time: %v\n", verifyTime)
}

func TestQuantumSigningWithDifferentMessages(t *testing.T) {
	signer, err := quantum_signing.NewDilithiumSigner()
	require.NoError(t, err)

	_, privKey, err := signer.GenerateKey()
	require.NoError(t, err)

	// Test with different message lengths
	tests := []struct {
		name    string
		message string
	}{
		{"empty message", ""},
		{"short message", "Hello"},
		{"long message", "This is a much longer message that should test the hashing and signing of larger data payloads."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := []byte(tt.message)
			signature, err := signer.Sign(privKey, msg)
			require.NoError(t, err)

			valid, err := signer.Verify(privKey, msg, signature)
			require.NoError(t, err)
			require.True(t, valid)
		})
	}
}
