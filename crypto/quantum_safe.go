package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"dilithium/dilithium3"
)

var (
	ErrInvalidKeySize     = errors.New("invalid key size")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrInvalidMessage     = errors.New("invalid message")
	ErrInvalidPublicKey   = errors.New("invalid public key")
	ErrInvalidPrivateKey  = errors.New("invalid private key")
)

// GenerateQuantumKeys generates a new Dilithium3 key pair
func GenerateQuantumKeys() ([]byte, []byte, error) {
	pub, priv, err := dilithium3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	
	return pub.Bytes(), priv.Bytes(), nil
}

// QuantumSign signs a message using Dilithium3
func QuantumSign(privKey []byte, msg []byte) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, ErrInvalidPrivateKey
	}
	if len(msg) == 0 {
		return nil, ErrInvalidMessage
	}
	
	priv := dilithium3.PrivateKeyFromBytes(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	
	// Hash message
	hash := sha256.Sum256(msg)
	
	// Sign hash
	sig, err := priv.Sign(rand.Reader, hash[:], nil)
	if err != nil {
		return nil, err
	}
	
	return sig, nil
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
	
	pub := dilithium3.PublicKeyFromBytes(pubKey)
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
	
	pub, priv, err := dilithium3.GenerateKeyFromSeed(seed)
	if err != nil {
		return nil, nil, err
	}
	
	return pub.Bytes(), priv.Bytes(), nil
}

// QuantumPublicKeyFromPrivate derives a public key from a private key
func QuantumPublicKeyFromPrivate(privKey []byte) ([]byte, error) {
	if len(privKey) == 0 {
		return nil, ErrInvalidPrivateKey
	}
	
	priv := dilithium3.PrivateKeyFromBytes(privKey)
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
	
	priv := dilithium3.PrivateKeyFromBytes(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	
	// Hash message and context
	hash := sha256.Sum256(append(msg, context...))
	
	// Sign hash
	sig, err := priv.Sign(rand.Reader, hash[:], nil)
	if err != nil {
		return nil, err
	}
	
	return sig, nil
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
	
	pub := dilithium3.PublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}
	
	// Hash message and context
	hash := sha256.Sum256(append(msg, context...))
	
	// Verify signature
	return pub.Verify(hash[:], sig), nil
} 