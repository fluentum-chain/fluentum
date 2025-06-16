// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ICEXVault {
    function executeTrade(
        address user,
        address tokenIn,
        address tokenOut,
        uint256 amount
    ) external returns (uint256);
    
    function addLiquidity(
        address provider,
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external;
    
    function removeLiquidity(
        address provider,
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external;
    
    function getExpectedOutput(
        address tokenIn,
        address tokenOut,
        uint256 amount
    ) external view returns (uint256);
    
    function getReserves(
        address tokenA,
        address tokenB
    ) external view returns (uint256 reserveA, uint256 reserveB);
    
    function getLiquidityInfo(
        address provider,
        address tokenA,
        address tokenB
    ) external view returns (
        uint256 amountA,
        uint256 amountB,
        uint256 lastUpdate
    );
} 