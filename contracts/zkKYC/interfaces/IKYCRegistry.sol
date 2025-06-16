// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IKYCRegistry {
    struct KYCData {
        uint256 level;
        uint256 lastUpdate;
        bool isActive;
        bytes32 dataHash;
    }
    
    function registerKYC(
        address user,
        uint256 level,
        bytes32 dataHash
    ) external;
    
    function updateKYC(
        address user,
        uint256 newLevel,
        bytes32 newDataHash
    ) external;
    
    function revokeKYC(address user) external;
    
    function getKYCData(
        address user
    ) external view returns (KYCData memory);
    
    function isKYCValid(
        address user,
        uint256 requiredLevel
    ) external view returns (bool);
    
    function getKYCLevel(address user) external view returns (uint256);
    
    function getLastUpdate(address user) external view returns (uint256);
    
    function isActive(address user) external view returns (bool);
    
    function getDataHash(address user) external view returns (bytes32);
} 