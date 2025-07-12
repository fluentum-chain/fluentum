package main

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/cloudflare/circl/sign/dilithium"
)

// TestDilithiumSigning tests basic signing and verification with Dilithium
func TestDilithiumSigning(t *testing.T) {
	// Use Dilithium3 mode for testing (balanced security and performance)
	signer := dilithium.Mode3

	// Generate a new key pair
	publicKey, privateKey, err := signer.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a message to sign
	message := []byte("Test message for quantum signing")

	// Sign the message
	signature := signer.Sign(privateKey, message)

	// Verify the signature
	if !signer.Verify(publicKey, message, signature) {
		t.Error("Failed to verify valid signature")
	}

	// Test with modified message (should fail)
	modifiedMessage := []byte("Modified test message")
	if signer.Verify(publicKey, modifiedMessage, signature) {
		t.Error("Verified signature with modified message")
	}
}

// TestDilithiumKeySizes verifies the expected key and signature sizes
func TestDilithiumKeySizes(t *testing.T) {
	signer := dilithium.Mode3
	
	// Generate a key pair to get actual sizes
	publicKey, privateKey, _ := signer.GenerateKey(rand.Reader)
	
	// Convert to bytes to check sizes
	pubBytes := publicKey.Bytes()
	privBytes := privateKey.Bytes()
	
	// Check public key size
	if len(pubBytes) == 0 {
		t.Error("Public key size is zero")
	}
	t.Logf("Public key size: %d bytes", len(pubBytes))
	
	// Check private key size
	if len(privBytes) == 0 {
		t.Error("Private key size is zero")
	}
	t.Logf("Private key size: %d bytes", len(privBytes))
	
	// Check signature size for a small message
	message := []byte("Test")
	signature := signer.Sign(privateKey, message)
	
	if len(signature) == 0 {
		t.Error("Signature size is zero")
	}
	t.Logf("Signature size: %d bytes", len(signature))
}

// BenchmarkDilithiumSigning benchmarks the signing operation
func BenchmarkDilithiumSigning(b *testing.B) {
	signer := dilithium.Mode3
	_, privateKey, _ := signer.GenerateKey(rand.Reader)
	message := make([]byte, 256) // 256 bytes of random data
	rand.Read(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = signer.Sign(privateKey, message)
	}
}

// BenchmarkDilithiumVerification benchmarks the verification operation
func BenchmarkDilithiumVerification(b *testing.B) {
	signer := dilithium.Mode3
	publicKey, privateKey, _ := signer.GenerateKey(rand.Reader)
	message := make([]byte, 256)
	rand.Read(message)
	signature := signer.Sign(privateKey, message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !signer.Verify(publicKey, message, signature) {
			b.Fatal("Verification failed")
		}
	}
}

// TestDilithiumModes tests all available Dilithium modes
func TestDilithiumModes(t *testing.T) {
	modes := []struct {
		name string
		mode dilithium.Mode
	}{
		{"Dilithium2", dilithium.Mode2}, // 128-bit security
		{"Dilithium3", dilithium.Mode3}, // 192-bit security (recommended)
		{"Dilithium5", dilithium.Mode5}, // 256-bit security
	}

	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			mode := m.mode
			// Generate key pair
			publicKey, privateKey, err := mode.GenerateKey(rand.Reader)
			if err != nil {
				t.Fatalf("Failed to generate key pair: %v", err)
			}

			// Test with different message sizes
			messageSizes := []int{0, 1, 64, 1024, 4096}
			for _, size := range messageSizes {
				message := make([]byte, size)
				rand.Read(message)

				// Sign the message
				signature := mode.Sign(privateKey, message)

				// Verify the signature
				if !mode.Verify(publicKey, message, signature) {
					t.Errorf("Failed to verify signature for message size %d", size)
				}

				// Test with modified signature (should fail)
				if len(signature) > 0 {
					modifiedSig := make([]byte, len(signature))
					copy(modifiedSig, signature)
					modifiedSig[0] ^= 0xFF // Flip some bits

					if mode.Verify(publicKey, message, modifiedSig) {
						t.Error("Verified with modified signature")
					}
				}

				// Test with modified message (should fail)
				if len(message) > 0 {
					modifiedMsg := make([]byte, len(message))
					copy(modifiedMsg, message)
					modifiedMsg[0] ^= 0xFF // Flip some bits

					if mode.Verify(publicKey, modifiedMsg, signature) {
						t.Error("Verified with modified message")
					}
				}
			}

			// Test with the serialized keys
			message := []byte("Test message")
			signature := mode.Sign(privateKey, message)

			if !mode.Verify(publicKey, message, signature) {
				t.Error("Failed to verify with keys")
			}

			// Log key and signature sizes
			pubBytes := publicKey.Bytes()
			privBytes := privateKey.Bytes()

			t.Logf("Mode: %s, PublicKey: %d bytes, PrivateKey: %d bytes, Signature: %d bytes",
				m.name, len(pubBytes), len(privBytes), len(signature))
		})
	}
}

// TestDilithiumPerformance measures and logs performance metrics
func TestDilithiumPerformance(t *testing.T) {
	signer := dilithium.Mode3
	publicKey, privateKey, _ := signer.GenerateKey(rand.Reader)
	message := make([]byte, 1024) // 1KB message
	rand.Read(message)

	// Measure signing time
	start := time.Now()
	signature := signer.Sign(privateKey, message)
	signTime := time.Since(start)

	// Measure verification time
	start = time.Now()
	if !signer.Verify(publicKey, message, signature) {
		t.Fatal("Verification failed")
	}
	verifyTime := time.Since(start)

	// Get key sizes
	pubBytes := publicKey.Bytes()
	privBytes := privateKey.Bytes()

	t.Logf("Performance metrics for Dilithium3:")
	t.Logf("  - Public key size: %d bytes", len(pubBytes))
	t.Logf("  - Private key size: %d bytes", len(privBytes))
	t.Logf("  - Signature size: %d bytes", len(signature))
	t.Logf("  - Signing time: %v", signTime)
	t.Logf("  - Verification time: %v", verifyTime)
}
