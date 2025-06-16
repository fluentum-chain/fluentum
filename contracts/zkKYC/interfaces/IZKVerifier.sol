// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IZKVerifier {
    function verifyProof(
        bytes calldata proof,
        uint256[2] calldata publicInputs
    ) external view returns (bool);
    
    function verifyBatch(
        bytes[] calldata proofs,
        uint256[2][] calldata publicInputs
    ) external view returns (bool[] memory);
    
    function getVerificationKey() external view returns (bytes memory);
    
    function getProofSize() external view returns (uint256);
    
    function getPublicInputSize() external view returns (uint256);
    
    function getVerificationGasCost() external view returns (uint256);
} 