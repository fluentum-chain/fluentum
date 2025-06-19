package quantum

import (
	"crypto/rand"
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

var (
	ErrInvalidPrivateKey = errors.New("invalid private key")
	ErrInvalidPublicKey  = errors.New("invalid public key")
)

var Mode3 = 0

type DilithiumSigner struct{}

func (s *DilithiumSigner) GenerateKeyPair() ([]byte, []byte, error) {
	pub, priv, err := dilithium.Mode3.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return pub.Bytes(), priv.Bytes(), nil
}

func (s *DilithiumSigner) Sign(privKey []byte, msg []byte) ([]byte, error) {
	priv := dilithium.Mode3.PrivateKeyFromBytes(privKey)
	if priv == nil {
		return nil, ErrInvalidPrivateKey
	}
	sig, err := priv.Sign(rand.Reader, msg, nil)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *DilithiumSigner) Verify(pubKey []byte, msg []byte, sig []byte) (bool, error) {
	pub := dilithium.Mode3.PublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}
	return pub.Verify(msg, sig), nil
}
