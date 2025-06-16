// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library QuantumMath {
    uint256 public constant PRECISION = 1e18;
    uint256 public constant MAX_SIGNATURE_SIZE = 2700;
    uint256 public constant MAX_PUBLIC_KEY_SIZE = 1472;
    uint256 public constant MAX_MESSAGE_SIZE = 32;
    
    function calculateSignatureHash(
        bytes calldata signature
    ) internal pure returns (bytes32) {
        return keccak256(signature);
    }
    
    function calculateMessageHash(
        bytes calldata message
    ) internal pure returns (bytes32) {
        return keccak256(message);
    }
    
    function calculatePublicKeyHash(
        bytes calldata publicKey
    ) internal pure returns (bytes32) {
        return keccak256(publicKey);
    }
    
    function calculateSignatureTimeout(
        uint256 timestamp,
        uint256 timeout
    ) internal pure returns (uint256) {
        return timestamp + timeout;
    }
    
    function calculateBatchVerificationGas(
        uint256 signatureCount,
        uint256 baseGas
    ) internal pure returns (uint256) {
        return baseGas * signatureCount;
    }
    
    function calculateVerificationCost(
        uint256 gasPrice,
        uint256 gasUsed
    ) internal pure returns (uint256) {
        return gasPrice * gasUsed;
    }
    
    function calculateSignatureSize(
        bytes calldata signature
    ) internal pure returns (uint256) {
        return signature.length;
    }
    
    function calculatePublicKeySize(
        bytes calldata publicKey
    ) internal pure returns (uint256) {
        return publicKey.length;
    }
    
    function calculateMessageSize(
        bytes calldata message
    ) internal pure returns (uint256) {
        return message.length;
    }
    
    function calculateSignatureOverhead(
        uint256 messageSize,
        uint256 signatureSize
    ) internal pure returns (uint256) {
        return signatureSize - messageSize;
    }
    
    function calculateVerificationOverhead(
        uint256 publicKeySize,
        uint256 signatureSize
    ) internal pure returns (uint256) {
        return publicKeySize + signatureSize;
    }
} 