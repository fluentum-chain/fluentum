// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IStrategy {
    function deposit(uint256 amount) external;
    function withdraw(uint256 amount) external returns (uint256);
    function harvest() external returns (uint256);
    function totalDeposits() external view returns (uint256);
    function getAPY() external view returns (uint256);
    function getRisk() external view returns (uint256);
    function getToken() external view returns (address);
    function getRewardToken() external view returns (address);
    function getLastHarvest() external view returns (uint256);
    function getPendingRewards() external view returns (uint256);
} 