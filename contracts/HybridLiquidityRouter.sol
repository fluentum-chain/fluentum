// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IFLUMXToken.sol";
import "./interfaces/IFluentumDEX.sol";
import "./interfaces/ICEXVault.sol";

contract HybridLiquidityRouter is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;

    // Constants
    uint256 public constant THRESHOLD = 10_000 * 10**18; // $10k
    uint256 public constant MAX_SLIPPAGE = 50; // 5%
    uint256 public constant MIN_LIQUIDITY = 1000 * 10**18; // 1000 FLUMX
    
    // State
    ICEXVault public immutable cexVault;
    IFluentumDEX public immutable dex;
    IFLUMXToken public immutable fluxToken;
    mapping(address => bool) public isWhitelisted;
    mapping(address => mapping(address => bool)) public isPairEnabled;
    mapping(address => uint256) public lastTrade;
    mapping(address => uint256) public totalTrades;
    mapping(address => uint256) public totalVolume;
    
    // Events
    event TokenPairEnabled(address indexed tokenA, address indexed tokenB);
    event TradeExecuted(
        address indexed user,
        address indexed tokenIn,
        address indexed tokenOut,
        uint256 amountIn,
        uint256 amountOut,
        bool isCEX
    );
    event LiquidityAdded(
        address indexed provider,
        address indexed tokenA,
        address indexed tokenB,
        uint256 amountA,
        uint256 amountB,
        bool isCEX
    );
    event LiquidityRemoved(
        address indexed provider,
        address indexed tokenA,
        address indexed tokenB,
        uint256 amountA,
        uint256 amountB,
        bool isCEX
    );
    event ThresholdUpdated(uint256 newThreshold);
    event SlippageUpdated(uint256 newSlippage);
    
    constructor(
        address _cexVault,
        address _dex,
        address _fluxToken
    ) {
        require(_cexVault != address(0), "Invalid CEX vault");
        require(_dex != address(0), "Invalid DEX");
        require(_fluxToken != address(0), "Invalid token");
        
        cexVault = ICEXVault(_cexVault);
        dex = IFluentumDEX(_dex);
        fluxToken = IFLUMXToken(_fluxToken);
    }
    
    function enableTokenPair(address tokenA, address tokenB) external onlyOwner {
        require(tokenA != address(0) && tokenB != address(0), "Invalid token");
        require(tokenA != tokenB, "Same token");
        require(!isPairEnabled[tokenA][tokenB], "Pair already enabled");
        
        isWhitelisted[tokenA] = true;
        isWhitelisted[tokenB] = true;
        isPairEnabled[tokenA][tokenB] = true;
        isPairEnabled[tokenB][tokenA] = true;
        
        emit TokenPairEnabled(tokenA, tokenB);
    }
    
    function executeTrade(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 minAmountOut
    ) external nonReentrant returns (uint256 received) {
        require(isPairEnabled[tokenIn][tokenOut], "Invalid pair");
        require(amountIn > 0, "Invalid amount");
        
        // Check if enough time has passed since last trade
        require(block.timestamp >= lastTrade[msg.sender] + 1 minutes, "Trade too soon");
        
        // Transfer tokens from user
        IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn);
        
        // Route trade based on amount
        if (amountIn > THRESHOLD) {
            received = _executeCEXTrade(msg.sender, tokenIn, tokenOut, amountIn);
        } else {
            received = _executeDEXTrade(tokenIn, tokenOut, amountIn);
        }
        
        require(received >= minAmountOut, "Slippage exceeded");
        
        // Transfer tokens to user
        IERC20(tokenOut).safeTransfer(msg.sender, received);
        
        // Update state
        lastTrade[msg.sender] = block.timestamp;
        totalTrades[msg.sender]++;
        totalVolume[msg.sender] += amountIn;
        
        emit TradeExecuted(msg.sender, tokenIn, tokenOut, amountIn, received, amountIn > THRESHOLD);
    }
    
    function _executeCEXTrade(
        address user,
        address tokenIn,
        address tokenOut,
        uint256 amount
    ) internal returns (uint256) {
        // Approve CEX vault
        IERC20(tokenIn).safeApprove(address(cexVault), amount);
        
        // Execute trade on CEX
        uint256 received = cexVault.executeTrade(user, tokenIn, tokenOut, amount);
        
        return received;
    }
    
    function _executeDEXTrade(
        address tokenIn,
        address tokenOut,
        uint256 amount
    ) internal returns (uint256) {
        // Approve DEX
        IERC20(tokenIn).safeApprove(address(dex), amount);
        
        // Execute trade on DEX
        uint256 received = dex.swap(tokenIn, tokenOut, amount, 0);
        
        return received;
    }
    
    function addLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB,
        bool isCEX
    ) external nonReentrant {
        require(isPairEnabled[tokenA][tokenB], "Invalid pair");
        require(amountA >= MIN_LIQUIDITY && amountB >= MIN_LIQUIDITY, "Insufficient liquidity");
        
        // Transfer tokens
        IERC20(tokenA).safeTransferFrom(msg.sender, address(this), amountA);
        IERC20(tokenB).safeTransferFrom(msg.sender, address(this), amountB);
        
        if (isCEX) {
            // Add liquidity to CEX
            IERC20(tokenA).safeApprove(address(cexVault), amountA);
            IERC20(tokenB).safeApprove(address(cexVault), amountB);
            cexVault.addLiquidity(msg.sender, tokenA, tokenB, amountA, amountB);
        } else {
            // Add liquidity to DEX
            IERC20(tokenA).safeApprove(address(dex), amountA);
            IERC20(tokenB).safeApprove(address(dex), amountB);
            dex.addLiquidity(tokenA, tokenB, amountA, amountB);
        }
        
        emit LiquidityAdded(msg.sender, tokenA, tokenB, amountA, amountB, isCEX);
    }
    
    function removeLiquidity(
        address tokenA,
        address tokenB,
        uint256 amountA,
        uint256 amountB,
        bool isCEX
    ) external nonReentrant {
        require(isPairEnabled[tokenA][tokenB], "Invalid pair");
        require(amountA > 0 && amountB > 0, "Invalid amount");
        
        if (isCEX) {
            // Remove liquidity from CEX
            cexVault.removeLiquidity(msg.sender, tokenA, tokenB, amountA, amountB);
        } else {
            // Remove liquidity from DEX
            dex.removeLiquidity(tokenA, tokenB, amountA, amountB);
        }
        
        emit LiquidityRemoved(msg.sender, tokenA, tokenB, amountA, amountB, isCEX);
    }
    
    function updateThreshold(uint256 newThreshold) external onlyOwner {
        require(newThreshold > 0, "Invalid threshold");
        THRESHOLD = newThreshold;
        emit ThresholdUpdated(newThreshold);
    }
    
    function updateSlippage(uint256 newSlippage) external onlyOwner {
        require(newSlippage <= 100, "Invalid slippage");
        MAX_SLIPPAGE = newSlippage;
        emit SlippageUpdated(newSlippage);
    }
    
    function getTradeInfo(address user) external view returns (
        uint256 lastTradeTime,
        uint256 totalTradesCount,
        uint256 totalVolumeAmount
    ) {
        return (
            lastTrade[user],
            totalTrades[user],
            totalVolume[user]
        );
    }
    
    function getExpectedOutput(
        address tokenIn,
        address tokenOut,
        uint256 amountIn
    ) external view returns (uint256) {
        require(isPairEnabled[tokenIn][tokenOut], "Invalid pair");
        
        if (amountIn > THRESHOLD) {
            return cexVault.getExpectedOutput(tokenIn, tokenOut, amountIn);
        } else {
            return dex.calculateOutputAmount(tokenIn, tokenOut, amountIn);
        }
    }
} 