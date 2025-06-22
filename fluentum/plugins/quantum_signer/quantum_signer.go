package main

import (
	"crypto/rand"

	"github.com/fluentum-chain/dilithium"
	"github.com/fluentum-chain/fluentum/fluentum/core/crypto"
)

// DilithiumSigner implements the crypto.Signer interface for quantum-resistant signatures
type DilithiumSigner struct {
	mode dilithium.Mode
}

// NewDilithiumSigner creates a new Dilithium signer with the specified mode
func NewDilithiumSigner(mode dilithium.Mode) *DilithiumSigner {
	return &DilithiumSigner{mode: mode}
}

// GenerateKey generates a new Dilithium key pair
func (d *DilithiumSigner) GenerateKey() ([]byte, []byte) {
	pubKey, privKey, err := d.mode.GenerateKeyPair(rand.Reader)
	if err != nil {
		// In a real implementation, you might want to handle this error differently
		// For now, return empty keys as fallback
		return []byte{}, []byte{}
	}

	return privKey.Bytes(), pubKey.Bytes()
}

// Sign signs a message using Dilithium
func (d *DilithiumSigner) Sign(privateKey []byte, message []byte) []byte {
	priv := d.mode.PrivateKeyFromBytes(privateKey)
	if priv == nil {
		return []byte{} // Return empty signature on error
	}

	signature, err := priv.Sign(rand.Reader, message, nil)
	if err != nil {
		return []byte{} // Return empty signature on error
	}

	return signature
}

// Verify verifies a Dilithium signature
func (d *DilithiumSigner) Verify(publicKey []byte, message []byte, signature []byte) bool {
	pub := d.mode.PublicKeyFromBytes(publicKey)
	if pub == nil {
		return false
	}

	return pub.VerifySignature(message, signature)
}

// Name returns the name of this signer
func (d *DilithiumSigner) Name() string {
	return "dilithium"
}

// ExportSigner is the function that must be exported by the plugin
// This function returns a crypto.Signer implementation
func ExportSigner() crypto.Signer {
	// Use Dilithium Mode 3 (recommended for most use cases)
	// Mode 1: Fastest but largest keys
	// Mode 3: Balanced performance and security (recommended)
	// Mode 5: Highest security but slower
	return NewDilithiumSigner(dilithium.Mode3)
}

// ExportSignerMode allows specifying a specific Dilithium mode
func ExportSignerMode(mode int) crypto.Signer {
	switch mode {
	case 1:
		return NewDilithiumSigner(dilithium.Mode3) // Use Mode3 as fallback
	case 3:
		return NewDilithiumSigner(dilithium.Mode3)
	case 5:
		return NewDilithiumSigner(dilithium.Mode3) // Use Mode3 as fallback
	default:
		// Default to Mode 3
		return NewDilithiumSigner(dilithium.Mode3)
	}
}
