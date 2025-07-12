package main

import (
	"fmt"
	"log"
	"time"

	"quantum_signing"
)

func main() {
	// Create a new quantum signer instance
	signer, err := quantum_signing.NewDilithiumSigner()
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	// Print the security mode being used
	fmt.Printf("Using %s\n\n", signer.GetMode())

	// Generate a new key pair
	fmt.Println("Generating quantum-resistant key pair...")
	publicKey, privateKey, err := signer.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Print key sizes
	pkSize, skSize := signer.KeySize()
	fmt.Printf("Public key size: %d bytes\n", len(publicKey))
	fmt.Printf("Private key size: %d bytes\n\n", len(privateKey))

	// Create a message to sign
	message := []byte("This is a test message for quantum signing")

	// Sign the message
	fmt.Println("Signing message...")
	startSign := time.Now()
	signature, err := signer.Sign(privateKey, message)
	signTime := time.Since(startSign)

	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	// Print signature size
	fmt.Printf("Signature size: %d bytes\n\n", len(signature))

	// Verify the signature
	fmt.Println("Verifying signature...")
	startVerify := time.Now()
	valid, err := signer.Verify(publicKey, message, signature)
	verifyTime := time.Since(startVerify)

	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	// Print results
	if valid {
		fmt.Println("✅ Signature is valid!")
	} else {
		fmt.Println("❌ Signature is invalid!")
	}

	// Print timing information
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Sign time: %v\n", signTime)
	fmt.Printf("  Verify time: %v\n", verifyTime)

	// Test with invalid signature
	testInvalidSignature(signer, publicKey, privateKey, message, signature)
}

func testInvalidSignature(signer *quantum_signing.DilithiumSigner, publicKey, privateKey, message, originalSig []byte) {
	fmt.Println("\nTesting with invalid signature...")

	// Create an invalid signature by modifying the original one
	invalidSig := make([]byte, len(originalSig))
	copy(invalidSig, originalSig)
	if len(invalidSig) > 0 {
		invalidSig[0] ^= 0xFF // Flip some bits to make the signature invalid
	}

	// Verify the invalid signature
	valid, err := signer.Verify(publicKey, message, invalidSig)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		fmt.Println("✅ Correctly detected invalid signature!")
	} else {
		fmt.Println("❌ Failed to detect invalid signature!")
	}

	// Test with modified message
	modifiedMessage := []byte("This is a modified message")
	valid, err = signer.Verify(publicKey, modifiedMessage, originalSig)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		fmt.Println("✅ Correctly detected signature mismatch with modified message!")
	} else {
		fmt.Println("❌ Failed to detect signature mismatch with modified message!")
	}

	// Test with wrong public key
	_, wrongPrivateKey, err := signer.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate wrong key pair: %v", err)
	}

	wrongPublicKey, err := signer.PublicKey(wrongPrivateKey)
	if err != nil {
		log.Fatalf("Failed to get wrong public key: %v", err)
	}

	valid, err = signer.Verify(wrongPublicKey, message, originalSig)
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	if !valid {
		fmt.Println("✅ Correctly detected signature with wrong public key!")
	} else {
		fmt.Println("❌ Failed to detect signature with wrong public key!")
	}
}
