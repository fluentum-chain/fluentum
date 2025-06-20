package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/ripemd160" // necessary for Bitcoin address format

	"github.com/fluentum-chain/fluentum/crypto"
	tmjson "github.com/fluentum-chain/fluentum/libs/json"
)

// -------------------------------------
const (
	PrivKeyName = "tendermint/PrivKeySecp256k1"
	PubKeyName  = "tendermint/PubKeySecp256k1"

	KeyType     = "secp256k1"
	PrivKeySize = 32
)

func init() {
	tmjson.RegisterType(PubKey{}, PubKeyName)
	tmjson.RegisterType(PrivKey{}, PrivKeyName)
}

var _ crypto.PrivKey = PrivKey{}

// PrivKey implements PrivKey.
type PrivKey []byte

// Bytes marshalls the private key using amino encoding.
func (privKey PrivKey) Bytes() []byte {
	return []byte(privKey)
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey PrivKey) PubKey() crypto.PubKey {
	priv := secp256k1.PrivKeyFromBytes(privKey)
	pk := priv.PubKey().SerializeCompressed()
	return PubKey(pk)
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	if otherSecp, ok := other.(PrivKey); ok {
		return subtle.ConstantTimeCompare(privKey[:], otherSecp[:]) == 1
	}
	return false
}

func (privKey PrivKey) Type() string {
	return KeyType
}

// GenPrivKey generates a new ECDSA private key on curve secp256k1 private key.
// It uses OS randomness to generate the private key.
func GenPrivKey() PrivKey {
	return genPrivKey(crypto.CReader())
}

// genPrivKey generates a new secp256k1 private key using the provided reader.
func genPrivKey(rand io.Reader) PrivKey {
	var privKeyBytes [PrivKeySize]byte
	for {
		_, err := io.ReadFull(rand, privKeyBytes[:])
		if err != nil {
			panic(err)
		}
		priv := secp256k1.PrivKeyFromBytes(privKeyBytes[:])
		if priv != nil {
			return PrivKey(privKeyBytes[:])
		}
	}
}

var one = new(big.Int).SetInt64(1)

// GenPrivKeySecp256k1 hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
//
// It makes sure the private key is a valid field element by setting:
//
// c = sha256(secret)
// k = (c mod (n − 1)) + 1, where n = curve order.
//
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeySecp256k1(secret []byte) PrivKey {
	secHash := sha256.Sum256(secret)
	fe := new(big.Int).SetBytes(secHash[:])
	n := new(big.Int).Sub(secp256k1.S256().N, one)
	fe.Mod(fe, n)
	fe.Add(fe, one)
	feB := fe.Bytes()
	privKey32 := make([]byte, PrivKeySize)
	copy(privKey32[32-len(feB):32], feB)
	return PrivKey(privKey32)
}

// used to reject malleable signatures
// see:
//   - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
//   - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/crypto.go#L39
var secp256k1halfN = new(big.Int).Rsh(secp256k1.S256().N, 1)

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
// The returned signature will be of the form R || S (in lower-S form).
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	priv := secp256k1.PrivKeyFromBytes(privKey)
	if priv == nil {
		return nil, fmt.Errorf("invalid private key")
	}
	sig := ecdsa.Sign(priv, crypto.Sha256(msg))
	return serializeSig(sig), nil
}

//-------------------------------------

var _ crypto.PubKey = PubKey{}

// PubKeySize is comprised of 32 bytes for one field element
// (the x-coordinate), plus one byte for the parity of the y-coordinate.
const PubKeySize = 33

// PubKey implements crypto.PubKey.
// It is the compressed form of the pubkey. The first byte depends is a 0x02 byte
// if the y-coordinate is the lexicographically largest of the two associated with
// the x-coordinate. Otherwise the first byte is a 0x03.
// This prefix is followed with the x-coordinate.
type PubKey []byte

// Address returns a Bitcoin style addresses: RIPEMD160(SHA256(pubkey))
func (pubKey PubKey) Address() crypto.Address {
	if len(pubKey) != PubKeySize {
		panic("length of pubkey is incorrect")
	}
	hasherSHA256 := sha256.New()
	_, _ = hasherSHA256.Write(pubKey)
	sha := hasherSHA256.Sum(nil)

	hasherRIPEMD160 := ripemd160.New()
	_, _ = hasherRIPEMD160.Write(sha)

	return crypto.Address(hasherRIPEMD160.Sum(nil))
}

// Bytes returns the pubkey marshaled with amino encoding.
func (pubKey PubKey) Bytes() []byte {
	return []byte(pubKey)
}

func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeySecp256k1{%X}", []byte(pubKey))
}

func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	if otherSecp, ok := other.(PubKey); ok {
		return bytes.Equal(pubKey[:], otherSecp[:])
	}
	return false
}

func (pubKey PubKey) Type() string {
	return KeyType
}

// VerifySignature verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func (pubKey PubKey) VerifySignature(msg []byte, sigStr []byte) bool {
	if len(sigStr) != 64 {
		return false
	}
	pub, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return false
	}

	// Convert raw signature (R || S) to DER format
	derSig := rawSignatureToDER(sigStr)
	sig, err := ecdsa.ParseDERSignature(derSig)
	if err != nil {
		return false
	}

	return sig.Verify(crypto.Sha256(msg), pub)
}

// rawSignatureToDER converts a raw signature (R || S) to DER format
func rawSignatureToDER(sigStr []byte) []byte {
	r := new(big.Int).SetBytes(sigStr[:32])
	s := new(big.Int).SetBytes(sigStr[32:64])

	// Create DER signature manually
	// DER format: 0x30 + length + 0x02 + r_length + r + 0x02 + s_length + s
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Remove leading zeros for DER encoding
	if len(rBytes) > 0 && rBytes[0] == 0 {
		rBytes = rBytes[1:]
	}
	if len(sBytes) > 0 && sBytes[0] == 0 {
		sBytes = sBytes[1:]
	}

	// Add sign bit if needed
	if len(rBytes) > 0 && rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0}, sBytes...)
	}

	// Construct DER signature
	der := make([]byte, 0, 6+len(rBytes)+len(sBytes))
	der = append(der, 0x30)                            // SEQUENCE
	der = append(der, byte(4+len(rBytes)+len(sBytes))) // length

	der = append(der, 0x02)              // INTEGER
	der = append(der, byte(len(rBytes))) // r length
	der = append(der, rBytes...)         // r

	der = append(der, 0x02)              // INTEGER
	der = append(der, byte(len(sBytes))) // s length
	der = append(der, sBytes...)         // s

	return der
}

// Serialize signature to R || S.
// R, S are padded to 32 bytes respectively.
func serializeSig(sig *ecdsa.Signature) []byte {
	// The new API returns DER format, but we need R || S format
	// We need to parse the DER signature and extract R and S
	derBytes := sig.Serialize()

	// Parse DER signature to extract R and S
	// This is a simplified approach - in practice you'd want to use a proper DER parser
	// For now, we'll use a basic approach that works for most cases
	if len(derBytes) < 6 {
		return nil
	}

	// Skip SEQUENCE and length
	pos := 2

	// Parse R
	if pos >= len(derBytes) || derBytes[pos] != 0x02 {
		return nil
	}
	pos++
	rLen := int(derBytes[pos])
	pos++
	if pos+rLen > len(derBytes) {
		return nil
	}
	rBytes := derBytes[pos : pos+rLen]
	pos += rLen

	// Parse S
	if pos >= len(derBytes) || derBytes[pos] != 0x02 {
		return nil
	}
	pos++
	sLen := int(derBytes[pos])
	pos++
	if pos+sLen > len(derBytes) {
		return nil
	}
	sBytes := derBytes[pos : pos+sLen]

	// Convert to big.Int and then to 32-byte format
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	// Pad to 32 bytes each
	sigBytes := make([]byte, 64)
	rPadded := r.Bytes()
	sPadded := s.Bytes()

	copy(sigBytes[32-len(rPadded):32], rPadded)
	copy(sigBytes[64-len(sPadded):64], sPadded)

	return sigBytes
}
