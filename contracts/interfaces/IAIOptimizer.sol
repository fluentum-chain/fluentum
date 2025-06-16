// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IAIOptimizer {
    function getToken() external view returns (address);
    
    function optimizeStrategy(
        address user,
        uint256 amount,
        bytes calldata riskProfile
    ) external returns (address strategy, uint256 allocation);
    
    function rebalanceStrategy(
        address user,
        uint256 totalAmount,
        address[] calldata currentStrategies,
        uint256[] calldata currentAllocations,
        bytes calldata riskProfile
    ) external returns (address[] memory strategies, uint256[] memory allocations);
    
    function getStrategyMetrics(
        address strategy
    ) external view returns (
        uint256 apy,
        uint256 risk,
        uint256 tvl,
        uint256 utilization
    );
    
    function updateModel(bytes calldata modelData) external;
} 