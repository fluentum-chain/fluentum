package zk_kyc

import (
	"math/big"
	"testing"
	"time"
)

func TestKYCVerifier(t *testing.T) {
	verifier, err := NewKYCVerifier()
	if err != nil {
		t.Fatalf("Failed to create KYC verifier: %v", err)
	}

	// Create test merchant data
	merchantData := &KYCData{
		MerchantID:    "test_merchant_1",
		Balance:       big.NewInt(1000000000), // 10 FLU
		KYCLevel:      3,
		VerifiedAt:    time.Now().Unix(),
		ExpiresAt:     time.Now().AddDate(1, 0, 0).Unix(),
		DocumentHash:  "0x1234567890abcdef",
		CountryCode:   "US",
		BusinessType:  "retail",
		RiskScore:     75,
	}

	// Add merchant to registry
	err = verifier.AddMerchant(merchantData)
	if err != nil {
		t.Fatalf("Failed to add merchant: %v", err)
	}

	// Generate proof
	proof, err := verifier.GenerateProof(merchantData.MerchantID)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	// Test verification with valid minimum balance
	minBalance := big.NewInt(500000000) // 5 FLU
	valid, err := verifier.VerifyProof(proof, minBalance)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}
	if !valid {
		t.Error("Proof verification failed with valid minimum balance")
	}

	// Test verification with invalid minimum balance
	minBalance = big.NewInt(2000000000) // 20 FLU
	valid, err = verifier.VerifyProof(proof, minBalance)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}
	if valid {
		t.Error("Proof verification succeeded with invalid minimum balance")
	}

	// Test with expired KYC
	expiredData := &KYCData{
		MerchantID:    "test_merchant_2",
		Balance:       big.NewInt(1000000000),
		KYCLevel:      3,
		VerifiedAt:    time.Now().Unix(),
		ExpiresAt:     time.Now().AddDate(-1, 0, 0).Unix(), // Expired
		DocumentHash:  "0xfedcba0987654321",
		CountryCode:   "US",
		BusinessType:  "retail",
		RiskScore:     75,
	}

	err = verifier.AddMerchant(expiredData)
	if err != nil {
		t.Fatalf("Failed to add expired merchant: %v", err)
	}

	proof, err = verifier.GenerateProof(expiredData.MerchantID)
	if err != nil {
		t.Fatalf("Failed to generate proof for expired merchant: %v", err)
	}

	valid, err = verifier.VerifyProof(proof, minBalance)
	if err != nil {
		t.Fatalf("Failed to verify expired proof: %v", err)
	}
	if valid {
		t.Error("Proof verification succeeded with expired KYC")
	}

	// Test with insufficient KYC level
	lowKYCData := &KYCData{
		MerchantID:    "test_merchant_3",
		Balance:       big.NewInt(1000000000),
		KYCLevel:      1, // Insufficient level
		VerifiedAt:    time.Now().Unix(),
		ExpiresAt:     time.Now().AddDate(1, 0, 0).Unix(),
		DocumentHash:  "0xabcdef1234567890",
		CountryCode:   "US",
		BusinessType:  "retail",
		RiskScore:     75,
	}

	err = verifier.AddMerchant(lowKYCData)
	if err != nil {
		t.Fatalf("Failed to add low KYC merchant: %v", err)
	}

	proof, err = verifier.GenerateProof(lowKYCData.MerchantID)
	if err != nil {
		t.Fatalf("Failed to generate proof for low KYC merchant: %v", err)
	}

	valid, err = verifier.VerifyProof(proof, minBalance)
	if err != nil {
		t.Fatalf("Failed to verify low KYC proof: %v", err)
	}
	if valid {
		t.Error("Proof verification succeeded with insufficient KYC level")
	}
} 