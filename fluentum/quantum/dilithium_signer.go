package quantum

import (
	"crypto/rand"
	"errors"

	"github.com/cloudflare/circl/sign/dilithium"
)

type DilithiumSigner struct {
	mode    dilithium.Mode
	privKey dilithium.PrivateKey
}

func NewDilithiumSigner() (*DilithiumSigner, error) {
	mode := dilithium.ModeByName("Dilithium3")
	if mode == nil {
		return nil, errors.New("Dilithium3 mode not supported")
	}

	_, privKey, err := mode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &DilithiumSigner{
		mode:    mode,
		privKey: privKey,
	}, nil
}

func (ds *DilithiumSigner) Sign(message []byte) ([]byte, error) {
	signature := make([]byte, ds.mode.SignatureSize())
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
