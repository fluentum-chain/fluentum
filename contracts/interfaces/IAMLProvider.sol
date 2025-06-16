// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IAMLProvider {
    function screenTransaction(
        address user,
        uint256 amount,
        address counterparty,
        uint256 usdValue
    ) external returns (bytes memory);
    
    function screenBatch(
        address[] calldata users,
        uint256[] calldata amounts,
        address[] calldata counterparties,
        uint256[] calldata usdValues
    ) external returns (bytes[] memory);
    
    function getProviderInfo() external view returns (
        string memory name,
        string memory version,
        uint256 maxBatchSize,
        uint256 minConfidence
    );
    
    function getScreeningCost(
        uint256 usdValue
    ) external view returns (uint256);
    
    function getBatchScreeningCost(
        uint256[] calldata usdValues
    ) external view returns (uint256);
    
    function isProviderActive() external view returns (bool);
    
    function getLastUpdate() external view returns (uint256);
    
    function getSupportedJurisdictions() external view returns (string[] memory);
    
    function getComplianceRules() external view returns (bytes memory);
} 