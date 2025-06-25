package compat

import (
	cmabci "github.com/cometbft/cometbft/abci/types"
	localcrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
)

// ToCmPublicKey converts a local proto PublicKey to the upstream CometBFT ABCI Go PublicKey.
func ToCmPublicKey(pk *localcrypto.PublicKey) cmabci.PublicKey {
	if pk == nil {
		return cmabci.PublicKey{}
	}
	if ed := pk.GetEd25519(); ed != nil {
		return cmabci.Ed25519PublicKey(ed)
	}
	if secp := pk.GetSecp256K1(); secp != nil {
		return cmabci.Secp256k1PublicKey(secp)
	}
	return cmabci.PublicKey{}
}

// ToLocalPublicKey converts an upstream CometBFT ABCI Go PublicKey to the local proto PublicKey.
func ToLocalPublicKey(pk cmabci.PublicKey) *localcrypto.PublicKey {
	switch pk := pk.(type) {
	case cmabci.Ed25519PublicKey:
		return &localcrypto.PublicKey{
			Sum: &localcrypto.PublicKey_Ed25519{Ed25519: []byte(pk)},
		}
	case cmabci.Secp256k1PublicKey:
		return &localcrypto.PublicKey{
			Sum: &localcrypto.PublicKey_Secp256K1{Secp256K1: []byte(pk)},
		}
	default:
		return &localcrypto.PublicKey{}
	}
}
