// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IFluentumDEX {
    function swap(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut
    ) external returns (uint256);
    
    function addLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external;
    
    function removeLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external;
    
    function calculateOutputAmount(
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) external view returns (uint256);
    
    function getReserves(
        address tokenA,
        address tokenB
    ) external view returns (uint256 reserveA, uint256 reserveB);
    
    function isWhitelisted(address token) external view returns (bool);
} 