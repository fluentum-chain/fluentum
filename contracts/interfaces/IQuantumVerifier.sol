// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IQuantumVerifier {
    function verifySignature(
        bytes calldata publicKey,
        bytes32 messageHash,
        bytes calldata signature
    ) external view returns (bool);
    
    function verifyBatch(
        bytes[] calldata publicKeys,
        bytes32[] calldata messageHashes,
        bytes[] calldata signatures
    ) external view returns (bool[] memory);
    
    function getVerificationKey() external view returns (bytes memory);
    
    function getProofSize() external view returns (uint256);
    
    function getPublicInputSize() external view returns (uint256);
    
    function getVerificationGasCost() external view returns (uint256);
} 