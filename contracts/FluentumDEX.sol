// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./GasStation.sol";
import "./interfaces/IFLUXToken.sol";

contract FluentumDEX is GasStation {
    using SafeERC20 for IERC20;

    // Constants
    uint256 public constant MIN_LIQUIDITY = 1000 * 10**18; // 1000 FLUX
    uint256 public constant MAX_SLIPPAGE = 50; // 5%
    
    // State
    mapping(address => mapping(address => uint256)) public reserves;
    mapping(address => bool) public isWhitelisted;
    mapping(address => uint256) public lastSwap;
    mapping(address => uint256) public totalSwaps;
    
    // Events
    event TokenPairAdded(address indexed tokenA, address indexed tokenB);
    event SwapExecuted(
        address indexed user,
        address indexed tokenIn,
        address indexed tokenOut,
        uint256 amountIn,
        uint256 amountOut
    );
    event LiquidityAdded(
        address indexed provider,
        address indexed tokenA,
        address indexed tokenB,
        uint256 amountA,
        uint256 amountB
    );
    event LiquidityRemoved(
        address indexed provider,
        address indexed tokenA,
        address indexed tokenB,
        uint256 amountA,
        uint256 amountB
    );
    
    constructor(
        address _fluxToken,
        address _staking
    ) GasStation(_fluxToken, _staking) {}
    
    function addTokenPair(address tokenA, address tokenB) external onlyOwner {
        require(tokenA != address(0) && tokenB != address(0), "Invalid token");
        require(tokenA != tokenB, "Same token");
        require(!isWhitelisted[tokenA] && !isWhitelisted[tokenB], "Already whitelisted");
        
        isWhitelisted[tokenA] = true;
        isWhitelisted[tokenB] = true;
        
        emit TokenPairAdded(tokenA, tokenB);
    }
    
    function swap(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut
    ) external nonReentrant zeroGas {
        require(isWhitelisted[tokenIn] && isWhitelisted[tokenOut], "Invalid pair");
        require(amountIn > 0, "Invalid amount");
        
        // Check if enough time has passed since last swap
        require(block.timestamp >= lastSwap[msg.sender] + 1 minutes, "Swap too soon");
        
        // Calculate output amount
        uint256 amountOut = calculateOutputAmount(tokenIn, tokenOut, amountIn);
        require(amountOut >= minAmountOut, "Slippage too high");
        
        // Transfer tokens
        IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn);
        IERC20(tokenOut).safeTransfer(msg.sender, amountOut);
        
        // Update reserves
        reserves[tokenIn][tokenOut] += amountIn;
        reserves[tokenOut][tokenIn] -= amountOut;
        
        // Update state
        lastSwap[msg.sender] = block.timestamp;
        totalSwaps[msg.sender]++;
        
        emit SwapExecuted(msg.sender, tokenIn, tokenOut, amountIn, amountOut);
    }
    
    function addLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external nonReentrant {
        require(isWhitelisted[tokenA] && isWhitelisted[tokenB], "Invalid pair");
        require(amountA >= MIN_LIQUIDITY && amountB >= MIN_LIQUIDITY, "Insufficient liquidity");
        
        // Transfer tokens
        IERC20(tokenA).safeTransferFrom(msg.sender, address(this), amountA);
        IERC20(tokenB).safeTransferFrom(msg.sender, address(this), amountB);
        
        // Update reserves
        reserves[tokenA][tokenB] += amountA;
        reserves[tokenB][tokenA] += amountB;
        
        emit LiquidityAdded(msg.sender, tokenA, tokenB, amountA, amountB);
    }
    
    function removeLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB
    ) external nonReentrant {
        require(isWhitelisted[tokenA] && isWhitelisted[tokenB], "Invalid pair");
        require(amountA > 0 && amountB > 0, "Invalid amount");
        
        // Check reserves
        require(reserves[tokenA][tokenB] >= amountA, "Insufficient reserve A");
        require(reserves[tokenB][tokenA] >= amountB, "Insufficient reserve B");
        
        // Update reserves
        reserves[tokenA][tokenB] -= amountA;
        reserves[tokenB][tokenA] -= amountB;
        
        // Transfer tokens
        IERC20(tokenA).safeTransfer(msg.sender, amountA);
        IERC20(tokenB).safeTransfer(msg.sender, amountB);
        
        emit LiquidityRemoved(msg.sender, tokenA, tokenB, amountA, amountB);
    }
    
    function calculateOutputAmount(
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) public view returns (uint256) {
        require(isWhitelisted[tokenIn] && isWhitelisted[tokenOut], "Invalid pair");
        
        uint256 reserveIn = reserves[tokenIn][tokenOut];
        uint256 reserveOut = reserves[tokenOut][tokenIn];
        
        require(reserveIn > 0 && reserveOut > 0, "Insufficient liquidity");
        
        // Constant product formula: x * y = k
        uint256 amountInWithFee = amountIn * 997; // 0.3% fee
        uint256 numerator = amountInWithFee * reserveOut;
        uint256 denominator = reserveIn * 1000 + amountInWithFee;
        
        return numerator / denominator;
    }
    
    function getReserves(
        address tokenA,
        address tokenB
    ) external view returns (uint256 reserveA, uint256 reserveB) {
        require(isWhitelisted[tokenA] && isWhitelisted[tokenB], "Invalid pair");
        return (reserves[tokenA][tokenB], reserves[tokenB][tokenA]);
    }
    
    function getSwapInfo(address user) external view returns (
        uint256 lastSwapTime,
        uint256 totalSwapsCount
    ) {
        return (lastSwap[user], totalSwaps[user]);
    }
} 