package zk_kyc

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/fluentum-chain/fluentum/crypto/merkle"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

// KYCData represents the KYC information for a merchant
type KYCData struct {
	MerchantID   string   `json:"merchant_id"`
	Balance      *big.Int `json:"balance"`
	KYCLevel     int      `json:"kyc_level"`
	VerifiedAt   int64    `json:"verified_at"`
	ExpiresAt    int64    `json:"expires_at"`
	DocumentHash string   `json:"document_hash"`
	CountryCode  string   `json:"country_code"`
	BusinessType string   `json:"business_type"`
	RiskScore    int      `json:"risk_score"`
}

// KYCVerifier handles KYC verification using zero-knowledge proofs
type KYCVerifier struct {
	registry map[string]*KYCData
	items    [][]byte
}

// NewKYCVerifier creates a new KYC verifier
func NewKYCVerifier() (*KYCVerifier, error) {
	return &KYCVerifier{
		registry: make(map[string]*KYCData),
		items:    make([][]byte, 0),
	}, nil
}

// AddMerchant adds a merchant to the KYC registry
func (v *KYCVerifier) AddMerchant(data *KYCData) error {
	// Hash the merchant data
	hash, err := v.hashMerchantData(data)
	if err != nil {
		return err
	}

	// Convert hash to bytes
	hashBytes := hash.Bytes()

	// Add to items for merkle tree
	v.items = append(v.items, hashBytes)

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

	// Find the index of this hash in the items
	hashBytes := hash.Bytes()
	index := -1
	for i, item := range v.items {
		if len(item) == len(hashBytes) {
			match := true
			for j := range item {
				if item[j] != hashBytes[j] {
					match = false
					break
				}
			}
			if match {
				index = i
				break
			}
		}
	}

	if index == -1 {
		return nil, errors.New("merchant hash not found in tree")
	}

	// Generate merkle proof
	rootHash, proofs := merkle.ProofsFromByteSlices(v.items)
	if index >= len(proofs) {
		return nil, errors.New("proof index out of range")
	}

	// Create proof data
	proofData := struct {
		Proof     *merkle.Proof
		Balance   *big.Int
		KYCLevel  int
		ExpiresAt int64
		RootHash  []byte
	}{
		Proof:     proofs[index],
		Balance:   data.Balance,
		KYCLevel:  data.KYCLevel,
		ExpiresAt: data.ExpiresAt,
		RootHash:  rootHash,
	}

	return json.Marshal(proofData)
}

// VerifyProof verifies a KYC proof
func (v *KYCVerifier) VerifyProof(proof []byte, minBalance *big.Int) (bool, error) {
	var proofData struct {
		Proof     *merkle.Proof
		Balance   *big.Int
		KYCLevel  int
		ExpiresAt int64
		RootHash  []byte
	}

	err := json.Unmarshal(proof, &proofData)
	if err != nil {
		return false, err
	}

	// Verify merkle proof
	err = proofData.Proof.Verify(proofData.RootHash, proofData.Proof.LeafHash)
	if err != nil {
		return false, err
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
