// Example demonstrates the usage of the quantum signing package
package main

import (
	"fmt"
	"log"

	"github.com/fluentum/quantum_signing"
)

func main() {
	// Create a new quantum signer instance
	signer, err := quantum_signing.NewDilithiumSigner()
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	// Generate a new key pair
	fmt.Println("Generating quantum-resistant key pair...")
	pubKey, privKey, err := signer.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a message to sign
	message := []byte("This is a test message for quantum signing")

	// Sign the message
	fmt.Println("Signing message...")
	signature, err := signer.Sign(privKey, message)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	// Verify the signature
	fmt.Println("Verifying signature...")
	valid, err := signer.Verify(pubKey, message, signature)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	// Print results
	if valid {
		fmt.Println("✅ Signature is valid!")
	} else {
		fmt.Println("❌ Signature is invalid!")
	}

	// Test with invalid signature
	invalidSig := make([]byte, len(signature))
	copy(invalidSig, signature)
	if len(invalidSig) > 0 {
		invalidSig[0] ^= 0xFF // Flip some bits to make the signature invalid
	}

	valid, err = signer.Verify(pubKey, message, invalidSig)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		fmt.Println("✅ Correctly detected invalid signature!")
	} else {
		fmt.Println("❌ Failed to detect invalid signature!")
	}
}
