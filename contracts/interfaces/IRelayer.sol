// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IRelayer {
    function relayTransaction(
        address targetChain,
        address contractAddress,
        bytes calldata payload,
        uint256 gasAmount,
        bytes32 txId
    ) external;
    
    function relayBatch(
        address targetChain,
        address[] calldata contractAddresses,
        bytes[] calldata payloads,
        uint256[] calldata gasAmounts,
        bytes32[] calldata txIds
    ) external;
    
    function getRelayerFee(
        address targetChain,
        uint256 gasAmount
    ) external view returns (uint256);
    
    function getRelayerCount() external view returns (uint256);
    
    function getActiveRelayers() external view returns (address[] memory);
    
    function isRelayerActive(address relayer) external view returns (bool);
    
    function getRelayerStake(address relayer) external view returns (uint256);
    
    function getRelayerPerformance(
        address relayer
    ) external view returns (
        uint256 successCount,
        uint256 failureCount,
        uint256 totalGasUsed
    );
} 