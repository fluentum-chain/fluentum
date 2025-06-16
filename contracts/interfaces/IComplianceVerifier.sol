// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IComplianceVerifier {
    function verifyScreening(
        address user,
        uint256 amount,
        address counterparty,
        bool sanctionsMatch,
        bool pepMatch,
        uint256 riskScore,
        uint256 confidence
    ) external view returns (bool);
    
    function verifyBatch(
        address[] calldata users,
        uint256[] calldata amounts,
        address[] calldata counterparties,
        bool[] calldata sanctionsMatches,
        bool[] calldata pepMatches,
        uint256[] calldata riskScores,
        uint256[] calldata confidences
    ) external view returns (bool[] memory);
    
    function getVerificationKey() external view returns (bytes memory);
    
    function getProofSize() external view returns (uint256);
    
    function getPublicInputSize() external view returns (uint256);
    
    function getVerificationGasCost() external view returns (uint256);
} 