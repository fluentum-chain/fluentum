package quantum

import (
	"crypto/rand"
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

var (
	dilithiumMode       = dilithium.ModeByName("Dilithium3")
	ErrInvalidPublicKey = errors.New("invalid public key")
)

type DilithiumSigner struct {
	privKey dilithium.PrivateKey
}

func NewDilithiumSigner() (*DilithiumSigner, error) {
	if dilithiumMode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	_, privKey, err := dilithiumMode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &DilithiumSigner{privKey: privKey}, nil
}

func (ds *DilithiumSigner) Sign(message []byte) ([]byte, error) {
	if dilithiumMode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	signature := make([]byte, dilithiumMode.SignatureSize())
	ds.privKey.Sign(signature, message, nil)
	return signature, nil
}

func (s *DilithiumSigner) Verify(pubKey []byte, msg []byte, sig []byte) (bool, error) {
	pub := dilithium.Mode3.PublicKeyFromBytes(pubKey)
	if pub == nil {
		return false, ErrInvalidPublicKey
	}
	return pub.Verify(msg, sig), nil
}
