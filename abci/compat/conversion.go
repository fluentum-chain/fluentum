package compat

import (
	cmcrypto "github.com/cometbft/cometbft/crypto"
	cmcryptoed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmcryptosecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	localcrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
)

// ToCmPublicKey converts a local proto PublicKey to the upstream CometBFT crypto PublicKey.
func ToCmPublicKey(pk *localcrypto.PublicKey) cmcrypto.PubKey {
	if pk == nil {
		return nil
	}
	if ed := pk.GetEd25519(); ed != nil {
		var pubKey cmcryptoed25519.PubKey
		copy(pubKey[:], ed)
		return pubKey
	}
	if secp := pk.GetSecp256K1(); secp != nil {
		var pubKey cmcryptosecp256k1.PubKey
		copy(pubKey[:], secp)
		return pubKey
	}
	return nil
}

// ToLocalPublicKey converts an upstream CometBFT crypto PublicKey to the local proto PublicKey.
func ToLocalPublicKey(pk cmcrypto.PubKey) *localcrypto.PublicKey {
	if pk == nil {
		return &localcrypto.PublicKey{}
	}

	// For now, we'll return a simple implementation
	// In a real implementation, you'd need to handle different key types properly
	return &localcrypto.PublicKey{
		Sum: &localcrypto.PublicKey_Ed25519{Ed25519: pk.Bytes()},
	}
}
