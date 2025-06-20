package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/cloudflare/circl/sign/dilithium"
)

var (
	ErrInvalidPrivateKey = errors.New("invalid private key")
	ErrInvalidPublicKey  = errors.New("invalid public key")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrInvalidSignature  = errors.New("invalid signature")
)

type DilithiumPrivateKey struct{}
type DilithiumPublicKey struct{}

func GenerateKey() (DilithiumPrivateKey, DilithiumPublicKey) {
	return DilithiumPrivateKey{}, DilithiumPublicKey{}
}
func PrivateKeyFromBytes(b []byte) DilithiumPrivateKey { return DilithiumPrivateKey{} }
func PublicKeyFromBytes(b []byte) DilithiumPublicKey   { return DilithiumPublicKey{} }

// Replace all Mode3 references with this pattern
func generateDilithiumKeys() ([]byte, []byte, error) {
	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, nil, errors.New("Dilithium3 mode not supported")
	}

	pubKey, privKey, err := mode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return pubKey.Bytes(), privKey.Bytes(), nil
}

// Update all signing operations similarly
func signWithDilithium(privateKey []byte, msg []byte) ([]byte, error) {
	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	privKey := mode.NewKeyFromSeed(privateKey)
	signature := make([]byte, mode.SignatureSize())
	privKey.Sign(signature, msg, nil)

	return signature, nil
}

// GenerateQuantumKeys generates a new Dilithium3 key pair
func GenerateQuantumKeys() ([]byte, []byte, error) {
	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, nil, errors.New("Dilithium3 mode not supported")
	}

	pubKey, privKey, err := mode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return pubKey.Bytes(), privKey.Bytes(), nil
}

// QuantumSign signs a message using Dilithium3
func QuantumSign(privKey []byte, msg []byte) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, ErrInvalidPrivateKey
	}
	if len(msg) == 0 {
		return nil, ErrInvalidMessage
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	priv := mode.NewKeyFromSeed(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}

	// Hash message
	hash := sha256.Sum256(msg)

	// Sign hash
	signature := make([]byte, mode.SignatureSize())
	priv.Sign(signature, hash[:], nil)

	return signature, nil
}

// QuantumVerify verifies a Dilithium3 signature
func QuantumVerify(pubKey []byte, msg []byte, sig []byte) (bool, error) {
	if len(pubKey) == 0 {
		return false, ErrInvalidPublicKey
	}
	if len(msg) == 0 {
		return false, ErrInvalidMessage
	}
	if len(sig) == 0 {
		return false, ErrInvalidSignature
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return false, errors.New("Dilithium3 mode not supported")
	}

	pub := mode.NewPublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}

	// Hash message
	hash := sha256.Sum256(msg)

	// Verify signature
	return pub.Verify(hash[:], sig), nil
}

// QuantumBatchVerify verifies multiple Dilithium3 signatures
func QuantumBatchVerify(
	pubKeys [][]byte,
	msgs [][]byte,
	sigs [][]byte,
) ([]bool, error) {
	if len(pubKeys) != len(msgs) || len(msgs) != len(sigs) {
		return nil, errors.New("length mismatch")
	}

	results := make([]bool, len(pubKeys))

	for i := range pubKeys {
		valid, err := QuantumVerify(pubKeys[i], msgs[i], sigs[i])
		if err != nil {
			return nil, err
		}
		results[i] = valid
	}

	return results, nil
}

// QuantumKeyFromSeed generates a Dilithium3 key pair from a seed
func QuantumKeyFromSeed(seed []byte) ([]byte, []byte, error) {
	if len(seed) < 32 {
		return nil, nil, errors.New("seed too short")
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, nil, errors.New("Dilithium3 mode not supported")
	}

	pub, priv := mode.GenerateKeyFromSeed(seed)
	return pub.Bytes(), priv.Bytes(), nil
}

// QuantumPublicKeyFromPrivate derives a public key from a private key
func QuantumPublicKeyFromPrivate(privKey []byte) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, ErrInvalidPrivateKey
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	priv := mode.NewKeyFromSeed(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}

	return priv.Public().Bytes(), nil
}

// QuantumSignWithContext signs a message with additional context
func QuantumSignWithContext(
	privKey []byte,
	msg []byte,
	context []byte,
) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, ErrInvalidPrivateKey
	}
	if len(msg) == 0 {
		return nil, ErrInvalidMessage
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	priv := mode.NewKeyFromSeed(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}

	// Hash message and context
	hash := sha256.Sum256(append(msg, context...))

	// Sign hash
	signature := make([]byte, mode.SignatureSize())
	priv.Sign(signature, hash[:], nil)

	return signature, nil
}

// QuantumVerifyWithContext verifies a signature with additional context
func QuantumVerifyWithContext(
	pubKey []byte,
	msg []byte,
	sig []byte,
	context []byte,
) (bool, error) {
	if len(pubKey) == 0 {
		return false, ErrInvalidPublicKey
	}
	if len(msg) == 0 {
		return false, ErrInvalidMessage
	}
	if len(sig) == 0 {
		return false, ErrInvalidSignature
	}

	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return false, errors.New("Dilithium3 mode not supported")
	}

	pub := mode.NewPublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}

	// Hash message and context
	hash := sha256.Sum256(append(msg, context...))

	// Verify signature
	return pub.Verify(hash[:], sig), nil
}
