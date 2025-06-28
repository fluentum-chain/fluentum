package encoding

import (
	"fmt"

	"github.com/fluentum-chain/fluentum/crypto"
	"github.com/fluentum-chain/fluentum/crypto/ed25519"
	"github.com/fluentum-chain/fluentum/crypto/secp256k1"
	"github.com/fluentum-chain/fluentum/libs/json"
	protocrypto "github.com/fluentum-chain/fluentum/proto/tendermint/crypto"
)

func init() {
	json.RegisterType((*protocrypto.PublicKey)(nil), "tendermint.crypto.PublicKey")
	json.RegisterType((*protocrypto.PublicKey_Ed25519)(nil), "tendermint.crypto.PublicKey_Ed25519")
	json.RegisterType((*protocrypto.PublicKey_Secp256K1)(nil), "tendermint.crypto.PublicKey_Secp256K1")
}

// PubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey
func PubKeyToProto(k crypto.PubKey) (protocrypto.PublicKey, error) {
	var kp protocrypto.PublicKey
	switch k := k.(type) {
	case ed25519.PubKey:
		kp = protocrypto.PublicKey{
			Sum: &protocrypto.PublicKey_Ed25519{
				Ed25519: k,
			},
		}
	case secp256k1.PubKey:
		kp = protocrypto.PublicKey{
			Sum: &protocrypto.PublicKey_Secp256K1{
				Secp256K1: k,
			},
		}
	default:
		return kp, fmt.Errorf("toproto: key type %v is not supported", k)
	}
	return kp, nil
}

// PubKeyFromProto takes a protobuf Pubkey and transforms it to a crypto.Pubkey
func PubKeyFromProto(k protocrypto.PublicKey) (crypto.PubKey, error) {
	switch k := k.Sum.(type) {
	case *protocrypto.PublicKey_Ed25519:
		if len(k.Ed25519) != ed25519.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeyEd25519. Got %d, expected %d",
				len(k.Ed25519), ed25519.PubKeySize)
		}
		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, k.Ed25519)
		return pk, nil
	case *protocrypto.PublicKey_Secp256K1:
		if len(k.Secp256K1) != secp256k1.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeySecp256k1. Got %d, expected %d",
				len(k.Secp256K1), secp256k1.PubKeySize)
		}
		pk := make(secp256k1.PubKey, secp256k1.PubKeySize)
		copy(pk, k.Secp256K1)
		return pk, nil
	default:
		return nil, fmt.Errorf("fromproto: key type %v is not supported", k)
	}
}
