package quantum_signing

import (
	"fmt"
	"testing"
)

func TestSmoke(t *testing.T) {
	// Create a new quantum signer instance
	signer, err := NewDilithiumSigner()
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Generate a new key pair
	pubKey, privKey, err := signer.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a message to sign
	message := []byte("Test message for smoke testing")

	// Sign the message
	signature, err := signer.Sign(privKey, message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Verify the signature
	valid, err := signer.Verify(pubKey, message, signature)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		t.Error("Signature verification failed")
	}

	// Test with invalid signature
	invalidSig := make([]byte, len(signature))
	copy(invalidSig, signature)
	if len(invalidSig) > 0 {
		invalidSig[0] ^= 0xFF // Flip some bits to make the signature invalid
	}

	valid, err = signer.Verify(pubKey, message, invalidSig)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if valid {
		t.Error("Invalid signature was incorrectly verified")
	}

	fmt.Println("✅ Smoke test passed successfully")
}
