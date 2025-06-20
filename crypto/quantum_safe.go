package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

var (
	ErrInvalidPrivateKey = errors.New("invalid private key")
	ErrInvalidPublicKey  = errors.New("invalid public key")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrInvalidSignature  = errors.New("invalid signature")
)

var dilithiumMode = dilithium.Mode3

type DilithiumPrivateKey struct{}
type DilithiumPublicKey struct{}

func GenerateKey() (DilithiumPrivateKey, DilithiumPublicKey) {
	return DilithiumPrivateKey{}, DilithiumPublicKey{}
}
func PrivateKeyFromBytes(b []byte) DilithiumPrivateKey { return DilithiumPrivateKey{} }
func PublicKeyFromBytes(b []byte) DilithiumPublicKey   { return DilithiumPublicKey{} }

// GenerateQuantumKeys generates a new Dilithium3 key pair
func GenerateQuantumKeys() ([]byte, []byte, error) {
	pub, priv, err := dilithiumMode.GenerateKeyPair(rand.Reader)
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
	priv := dilithiumMode.PrivateKeyFromBytes(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	hash := sha256.Sum256(msg)
	return priv.Sign(rand.Reader, hash[:], nil)
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
	pub := dilithiumMode.PublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}
	hash := sha256.Sum256(msg)
	return pub.VerifySignature(hash[:], sig), nil
}
