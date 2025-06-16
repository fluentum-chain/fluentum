pragma circom 2.0.0;

include "circomlib/comparators.circom";
include "circomlib/poseidon.circom";

// Template for verifying KYC membership
template VerifyKYCMembership() {
    signal input proof;
    signal output valid;
    signal output balance;

    // Poseidon hash for proof verification
    component hasher = Poseidon(1);
    hasher.inputs[0] <== proof;

    // Verify proof against KYC registry
    component registry = KYCRegistry();
    registry.proof <== hasher.out;
    
    valid <== registry.valid;
    balance <== registry.balance;
}

// Template for KYC merchant verification
template KYCMerchant() {
    signal input kycProof;
    signal input minBalance;
    signal output verified;
    
    // Verify KYC membership and get balance
    component verifier = VerifyKYCMembership();
    verifier.proof <== kycProof;
    
    // Check if balance meets minimum requirement
    component balanceCheck = GreaterThan(64);
    balanceCheck.in[0] <== verifier.balance;
    balanceCheck.in[1] <== minBalance;
    
    // Final verification combines KYC validity and balance check
    verified <== verifier.valid && balanceCheck.out;
}

// KYC Registry template for storing and verifying KYC data
template KYCRegistry() {
    signal input proof;
    signal output valid;
    signal output balance;

    // Merkle tree for KYC data
    component tree = MerkleTree(32);
    
    // Verify proof against merkle root
    component verifier = MerkleVerifier();
    verifier.root <== tree.root;
    verifier.proof <== proof;
    
    valid <== verifier.valid;
    balance <== verifier.balance;
}

// Main component
component main = KYCMerchant(); 