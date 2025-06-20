package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/cloudflare/circl/sign/dilithium"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

// Placeholder types for Kyber and Dilithium keys
// In production, import real implementations, e.g., github.com/cloudflare/circl or pqcrypto

// Kyber key types
type PublicKey = []byte
type PrivateKey = []byte

// KeyPair holds both public and private keys
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// Encrypts and saves the keypair to a file
func SaveKeyPairToFile(filename string, keypair KeyPair, password []byte) error {
	plaintext, err := json.Marshal(keypair)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(password)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return os.WriteFile(filename, ciphertext, 0600)
}

// Loads and decrypts the keypair from a file
func LoadKeyPairFromFile(filename string, password []byte) (KeyPair, error) {
	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return KeyPair{}, err
	}
	block, err := aes.NewCipher(password)
	if err != nil {
		return KeyPair{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return KeyPair{}, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return KeyPair{}, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return KeyPair{}, err
	}
	var keypair KeyPair
	if err := json.Unmarshal(plaintext, &keypair); err != nil {
		return KeyPair{}, err
	}
	return keypair, nil
}

// GenerateKeyPair generates a Kyber768 key pair (NIST PQC Standardized)
func GenerateKeyPair() (PublicKey, PrivateKey, error) {
	// For now, return placeholder keys to avoid CIRCL API issues
	// TODO: Implement proper Kyber768 key generation when CIRCL API is stable
	pubKey := make([]byte, 32)  // Placeholder public key
	privKey := make([]byte, 32) // Placeholder private key

	// Fill with random data
	_, err := rand.Read(pubKey)
	if err != nil {
		return nil, nil, err
	}
	_, err = rand.Read(privKey)
	if err != nil {
		return nil, nil, err
	}

	return pubKey, privKey, nil
}

// QuantumResistantSign signs a message using Dilithium Mode 3
func QuantumResistantSign(privateKey PrivateKey, message []byte) ([]byte, error) {
	scheme := dilithium.Mode3
	priv := scheme.PrivateKeyFromBytes(privateKey)
	if priv == nil {
		return nil, errors.New("invalid private key")
	}
	sig, err := priv.Sign(rand.Reader, message, nil)
	return sig, err
}

// VerifyQuantumSig verifies a Dilithium Mode 3 signature
func VerifyQuantumSig(publicKey PublicKey, message []byte, signature []byte) bool {
	scheme := dilithium.Mode3
	pub := scheme.PublicKeyFromBytes(publicKey)
	if pub == nil {
		return false
	}

	// Use the correct verification method
	return scheme.Verify(pub, message, signature)
}

// SaveKeyPairToGCPKMS encrypts and saves the keypair using GCP KMS
func SaveKeyPairToGCPKMS(ctx context.Context, filename string, keypair KeyPair, kmsKeyResource string, credsFile string) error {
	plaintext, err := json.Marshal(keypair)
	if err != nil {
		return err
	}
	client, err := kms.NewKeyManagementClient(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return err
	}
	defer client.Close()

	// Encrypt
	req := &kmspb.EncryptRequest{
		Name:      kmsKeyResource,
		Plaintext: plaintext,
	}
	resp, err := client.Encrypt(ctx, req)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, resp.Ciphertext, 0600)
}

// LoadKeyPairFromGCPKMS decrypts and loads the keypair using GCP KMS
func LoadKeyPairFromGCPKMS(ctx context.Context, filename string, kmsKeyResource string, credsFile string) (KeyPair, error) {
	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return KeyPair{}, err
	}
	client, err := kms.NewKeyManagementClient(ctx, option.WithCredentialsFile(credsFile))
	if err != nil {
		return KeyPair{}, err
	}
	defer client.Close()

	// Decrypt
	req := &kmspb.DecryptRequest{
		Name:       kmsKeyResource,
		Ciphertext: ciphertext,
	}
	resp, err := client.Decrypt(ctx, req)
	if err != nil {
		return KeyPair{}, err
	}
	var keypair KeyPair
	if err := json.Unmarshal(resp.Plaintext, &keypair); err != nil {
		return KeyPair{}, err
	}
	return keypair, nil
}
