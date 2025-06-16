package quantum

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/dilithium"
)

// DilithiumMode represents the security level of Dilithium
type DilithiumMode int

const (
	// Mode2 provides 128-bit security
	Mode2 DilithiumMode = iota
	// Mode3 provides 192-bit security
	Mode3
	// Mode5 provides 256-bit security
	Mode5
)

var (
	// DefaultMode is the recommended security level
	DefaultMode = Mode3

	// modeMap maps DilithiumMode to dilithium.Mode
	modeMap = map[DilithiumMode]dilithium.Mode{
		Mode2: dilithium.Mode2,
		Mode3: dilithium.Mode3,
		Mode5: dilithium.Mode5,
	}

	ErrInvalidMode = errors.New("invalid dilithium mode")
	ErrInvalidKey  = errors.New("invalid key")
)

// DilithiumSigner implements quantum-resistant digital signatures using Dilithium
type DilithiumSigner struct {
	mode       DilithiumMode
	PublicKey  []byte
	PrivateKey []byte
}

// NewDilithiumSigner creates a new DilithiumSigner instance
func NewDilithiumSigner(mode DilithiumMode) (*DilithiumSigner, error) {
	if _, ok := modeMap[mode]; !ok {
		return nil, ErrInvalidMode
	}

	return &DilithiumSigner{
		mode: mode,
	}, nil
}

// GenerateKeyPair generates a new Dilithium key pair
func (d *DilithiumSigner) GenerateKeyPair() error {
	pk, sk, err := modeMap[d.mode].GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	d.PublicKey = pk.Bytes()
	d.PrivateKey = sk.Bytes()
	return nil
}

// Sign creates a quantum-resistant signature for the given message
func (d *DilithiumSigner) Sign(message []byte) ([]byte, error) {
	if d.PrivateKey == nil {
		return nil, ErrInvalidKey
	}

	sk := modeMap[d.mode].PrivateKeyFromBytes(d.PrivateKey)
	if sk == nil {
		return nil, ErrInvalidKey
	}

	sig, err := sk.Sign(rand.Reader, message, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return sig, nil
}

// Verify verifies a quantum-resistant signature
func (d *DilithiumSigner) Verify(message, signature []byte) (bool, error) {
	if d.PublicKey == nil {
		return false, ErrInvalidKey
	}

	pk := modeMap[d.mode].PublicKeyFromBytes(d.PublicKey)
	if pk == nil {
		return false, ErrInvalidKey
	}

	return pk.Verify(message, signature), nil
}

// Package-level functions for direct use

// GenerateKeyPair generates a new Dilithium key pair using the default mode
func GenerateKeyPair() ([]byte, []byte, error) {
	pk, sk, err := modeMap[DefaultMode].GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return pk.Bytes(), sk.Bytes(), nil
}

// Sign creates a quantum-resistant signature for the given message
func Sign(privateKey []byte, msg []byte) ([]byte, error) {
	sk := modeMap[DefaultMode].PrivateKeyFromBytes(privateKey)
	if sk == nil {
		return nil, ErrInvalidKey
	}

	sig, err := sk.Sign(rand.Reader, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return sig, nil
}

// Verify verifies a quantum-resistant signature
func Verify(publicKey []byte, msg []byte, sig []byte) (bool, error) {
	pk := modeMap[DefaultMode].PublicKeyFromBytes(publicKey)
	if pk == nil {
		return false, ErrInvalidKey
	}

	return pk.Verify(msg, sig), nil
}

// VerifySignature is a convenience function for verifying signatures in the consensus package
func VerifySignature(publicKey, message, signature []byte) bool {
	valid, _ := Verify(publicKey, message, signature)
	return valid
}
