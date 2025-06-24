package zk_kyc

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
)

// KYCData represents the KYC information for a merchant
type KYCData struct {
	MerchantID    string   `json:"merchant_id"`
	Balance       *big.Int `json:"balance"`
	KYCLevel      int      `json:"kyc_level"`
	VerifiedAt    int64    `json:"verified_at"`
	ExpiresAt     int64    `json:"expires_at"`
	DocumentHash  string   `json:"document_hash"`
	CountryCode   string   `json:"country_code"`
	BusinessType  string   `json:"business_type"`
	RiskScore     int      `json:"risk_score"`
}

// KYCVerifier handles KYC verification using zero-knowledge proofs
type KYCVerifier struct {
	tree     *merkletree.MerkleTree
	registry map[string]*KYCData
}

// NewKYCVerifier creates a new KYC verifier
func NewKYCVerifier() (*KYCVerifier, error) {
	tree, err := merkletree.NewMerkleTree(context.Background(), merkletree.MemoryStorage{}, 32)
	if err != nil {
		return nil, err
	}

	return &KYCVerifier{
		tree:     tree,
		registry: make(map[string]*KYCData),
	}, nil
}

// AddMerchant adds a merchant to the KYC registry
func (v *KYCVerifier) AddMerchant(data *KYCData) error {
	// Hash the merchant data
	hash, err := v.hashMerchantData(data)
	if err != nil {
		return err
	}

	// Add to merkle tree
	err = v.tree.Add(context.Background(), hash)
	if err != nil {
		return err
	}

	// Store in registry
	v.registry[data.MerchantID] = data
	return nil
}

// GenerateProof generates a zero-knowledge proof for KYC verification
func (v *KYCVerifier) GenerateProof(merchantID string) ([]byte, error) {
	data, exists := v.registry[merchantID]
	if !exists {
		return nil, errors.New("merchant not found")
	}

	// Hash the merchant data
	hash, err := v.hashMerchantData(data)
	if err != nil {
		return nil, err
	}

	// Generate merkle proof
	proof, err := v.tree.GenerateProof(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	// Create proof data
	proofData := struct {
		Proof     *merkletree.Proof
		Balance   *big.Int
		KYCLevel  int
		ExpiresAt int64
	}{
		Proof:     proof,
		Balance:   data.Balance,
		KYCLevel:  data.KYCLevel,
		ExpiresAt: data.ExpiresAt,
	}

	return json.Marshal(proofData)
}

// VerifyProof verifies a KYC proof
func (v *KYCVerifier) VerifyProof(proof []byte, minBalance *big.Int) (bool, error) {
	var proofData struct {
		Proof     *merkletree.Proof
		Balance   *big.Int
		KYCLevel  int
		ExpiresAt int64
	}

	err := json.Unmarshal(proof, &proofData)
	if err != nil {
		return false, err
	}

	// Verify merkle proof
	valid, err := v.tree.VerifyProof(context.Background(), proofData.Proof)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, nil
	}

	// Check balance
	if proofData.Balance.Cmp(minBalance) < 0 {
		return false, nil
	}

	// Check KYC level
	if proofData.KYCLevel < 2 {
		return false, nil
	}

	// Check expiration
	if proofData.ExpiresAt < time.Now().Unix() {
		return false, nil
	}

	return true, nil
}

// hashMerchantData creates a Poseidon hash of the merchant data
func (v *KYCVerifier) hashMerchantData(data *KYCData) (*big.Int, error) {
	// Convert data to bytes
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create SHA256 hash
	shaHash := sha256.Sum256(bytes)
	
	// Convert to big.Int
	hash := new(big.Int).SetBytes(shaHash[:])
	
	// Create Poseidon hash
	return poseidon.Hash([]*big.Int{hash})
} 
