pragma circom 2.0.0;

include "circomlib/poseidon.circom";
include "circomlib/comparators.circom";

template KYCMerchantVerification() {
    // Public inputs
    signal input minBalance;
    signal input minKycLevel;
    
    // Private inputs
    signal private input kycProof;
    signal private input balance;
    signal private input kycLevel;
    signal private input dataHash;
    
    // Output
    signal output verified;
    
    // Components
    component poseidon = Poseidon(3);
    component balanceCheck = GreaterThan(64);
    component levelCheck = GreaterThan(64);
    component hashCheck = IsEqual();
    
    // Verify KYC proof
    poseidon.inputs[0] <== kycProof;
    poseidon.inputs[1] <== balance;
    poseidon.inputs[2] <== kycLevel;
    
    // Verify data hash
    hashCheck.in[0] <== poseidon.out;
    hashCheck.in[1] <== dataHash;
    
    // Verify minimum balance
    balanceCheck.in[0] <== balance;
    balanceCheck.in[1] <== minBalance;
    
    // Verify minimum KYC level
    levelCheck.in[0] <== kycLevel;
    levelCheck.in[1] <== minKycLevel;
    
    // All checks must pass
    verified <== hashCheck.out * balanceCheck.out * levelCheck.out;
}

component main = KYCMerchantVerification(); 