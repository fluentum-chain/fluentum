package quantum

import (
	"crypto/rand"
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

var (
	dilithiumMode       = dilithium.Mode3
	ErrInvalidPublicKey = errors.New("invalid public key")
)

type DilithiumSigner struct {
	privKey []byte
}

func NewDilithiumSigner() (*DilithiumSigner, error) {
	_, privKey, err := dilithiumMode.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &DilithiumSigner{privKey: privKey.Bytes()}, nil
}

func (ds *DilithiumSigner) Sign(message []byte) ([]byte, error) {
	priv := dilithiumMode.PrivateKeyFromBytes(ds.privKey)
	if priv == nil {
		return nil, errors.New("invalid private key")
	}
	return priv.Sign(rand.Reader, message, nil)
}

func (ds *DilithiumSigner) Verify(pubKey []byte, msg []byte, sig []byte) (bool, error) {
	pub := dilithiumMode.PublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}
	return pub.Verify(msg, sig), nil
}
