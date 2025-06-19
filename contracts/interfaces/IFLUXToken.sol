// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IFLUXToken {
    function mint(address to, uint256 amount) external;
    function burn(uint256 amount) external;
    function transfer(address to, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function approve(address spender, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
    function totalSupply() external view returns (uint256);
    function allowance(address owner, address spender) external view returns (uint256);
} 