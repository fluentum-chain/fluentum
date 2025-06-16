// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IFluentumVerifier {
    function verifyProof(
        bytes32 txId,
        address user,
        bytes calldata payload,
        bytes calldata proof
    ) external view returns (bool);
    
    function verifyBatch(
        bytes32[] calldata txIds,
        address[] calldata users,
        bytes[] calldata payloads,
        bytes[] calldata proofs
    ) external view returns (bool[] memory);
    
    function getVerificationKey() external view returns (bytes memory);
    
    function getProofSize() external view returns (uint256);
    
    function getPublicInputSize() external view returns (uint256);
    
    function getVerificationGasCost() external view returns (uint256);
} 