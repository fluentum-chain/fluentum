package crypto

import (
	"crypto/rand"
	"fmt"

	"github.com/cloudflare/circl/sign/dilithium"
)

// Use this for all mode references
var dilithiumMode = dilithium.Mode3

func GenerateQuantumKeys() ([]byte, []byte, error) {
	pubKey, privKey, err := dilithiumMode.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate keys: %w", err)
	}
	return pubKey.Bytes(), privKey.Bytes(), nil
}
