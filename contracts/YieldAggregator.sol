// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IStrategy.sol";
import "./interfaces/IAIOptimizer.sol";
import "./libraries/YieldMath.sol";

interface IERC20 {
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

interface IProtocolAdapter {
    function deposit(uint256 amount) external;
    function withdraw(uint256 amount) external;
    function getBalance(address user) external view returns (uint256);
}

contract YieldAggregator is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;

    // Constants
    uint256 public constant MIN_DEPOSIT = 100 * 10**18; // 100 FLU
    uint256 public constant MAX_STRATEGIES = 10;
    uint256 public constant MAX_SLIPPAGE = 50; // 5%
    
    // State
    IStrategy[] public strategies;
    mapping(address => bool) public isStrategyEnabled;
    mapping(address => uint256) public userDeposits;
    mapping(address => mapping(address => uint256)) public strategyDeposits;
    mapping(address => uint256) public lastHarvest;
    mapping(address => uint256) public totalHarvested;
    
    // AI Optimizer
    IAIOptimizer public immutable aiOptimizer;
    
    // Mapping from strategy name to protocol adapter
    mapping(string => address) public adapters;
    // User balances per strategy
    mapping(address => mapping(string => uint256)) public userStrategyBalances;
    
    // Events
    event StrategyAdded(address indexed strategy);
    event StrategyRemoved(address indexed strategy);
    event Deposit(
        address indexed user,
        address indexed strategy,
        uint256 amount
    );
    event Withdraw(
        address indexed user,
        address indexed strategy,
        uint256 amount
    );
    event Harvest(
        address indexed user,
        address indexed strategy,
        uint256 amount
    );
    event StrategyOptimized(
        address indexed user,
        address indexed strategy,
        uint256 allocation
    );
    event Staked(address indexed user, uint256 amount);
    event Lent(address indexed user, uint256 amount);
    event ProvidedLiquidity(address indexed user, uint256 amount);
    event Withdrawn(address indexed user, uint256 amount, string strategy);
    event AdapterRegistered(string strategy, address adapter);
    event Rebalanced(address indexed user, string fromStrategy, string toStrategy, uint256 amount);
    
    constructor(address _aiOptimizer) {
        require(_aiOptimizer != address(0), "Invalid optimizer");
        aiOptimizer = IAIOptimizer(_aiOptimizer);
    }
    
    function addStrategy(address strategy) external onlyOwner {
        require(strategy != address(0), "Invalid strategy");
        require(!isStrategyEnabled[strategy], "Strategy already added");
        require(strategies.length < MAX_STRATEGIES, "Max strategies reached");
        
        strategies.push(IStrategy(strategy));
        isStrategyEnabled[strategy] = true;
        
        emit StrategyAdded(strategy);
    }
    
    function removeStrategy(uint256 index) external onlyOwner {
        require(index < strategies.length, "Invalid index");
        
        address strategy = address(strategies[index]);
        isStrategyEnabled[strategy] = false;
        
        // Remove strategy from array
        strategies[index] = strategies[strategies.length - 1];
        strategies.pop();
        
        emit StrategyRemoved(strategy);
    }
    
    function deposit(
        uint256 amount,
        bytes calldata riskProfile
    ) external nonReentrant {
        require(amount >= MIN_DEPOSIT, "Insufficient deposit");
        require(strategies.length > 0, "No strategies available");
        
        // Transfer tokens
        IERC20(aiOptimizer.getToken()).safeTransferFrom(msg.sender, address(this), amount);
        
        // Get AI-optimized strategy allocation
        (address strategy, uint256 allocation) = aiOptimizer.optimizeStrategy(
            msg.sender,
            amount,
            riskProfile
        );
        
        require(isStrategyEnabled[strategy], "Strategy not enabled");
        require(allocation <= amount, "Invalid allocation");
        
        // Approve and deposit to strategy
        IERC20(aiOptimizer.getToken()).safeApprove(strategy, allocation);
        IStrategy(strategy).deposit(allocation);
        
        // Update state
        userDeposits[msg.sender] += amount;
        strategyDeposits[msg.sender][strategy] += allocation;
        
        emit Deposit(msg.sender, strategy, allocation);
        emit StrategyOptimized(msg.sender, strategy, allocation);
    }
    
    function withdraw(
        address strategy,
        uint256 amount
    ) external nonReentrant {
        require(isStrategyEnabled[strategy], "Strategy not enabled");
        require(strategyDeposits[msg.sender][strategy] >= amount, "Insufficient deposit");
        
        // Withdraw from strategy
        uint256 withdrawn = IStrategy(strategy).withdraw(amount);
        require(withdrawn >= amount, "Slippage exceeded");
        
        // Update state
        userDeposits[msg.sender] -= amount;
        strategyDeposits[msg.sender][strategy] -= amount;
        
        // Transfer tokens to user
        IERC20(aiOptimizer.getToken()).safeTransfer(msg.sender, withdrawn);
        
        emit Withdraw(msg.sender, strategy, amount);
    }
    
    function harvest(address strategy) external nonReentrant {
        require(isStrategyEnabled[strategy], "Strategy not enabled");
        require(strategyDeposits[msg.sender][strategy] > 0, "No deposit");
        
        // Check harvest cooldown
        require(block.timestamp >= lastHarvest[msg.sender] + 1 days, "Harvest too soon");
        
        // Harvest rewards
        uint256 harvested = IStrategy(strategy).harvest();
        require(harvested > 0, "No rewards");
        
        // Update state
        lastHarvest[msg.sender] = block.timestamp;
        totalHarvested[msg.sender] += harvested;
        
        // Transfer rewards to user
        IERC20(aiOptimizer.getToken()).safeTransfer(msg.sender, harvested);
        
        emit Harvest(msg.sender, strategy, harvested);
    }
    
    function rebalance(
        bytes calldata riskProfile
    ) external nonReentrant {
        require(userDeposits[msg.sender] > 0, "No deposits");
        
        // Get current allocations
        address[] memory currentStrategies = new address[](strategies.length);
        uint256[] memory currentAllocations = new uint256[](strategies.length);
        
        for (uint256 i = 0; i < strategies.length; i++) {
            currentStrategies[i] = address(strategies[i]);
            currentAllocations[i] = strategyDeposits[msg.sender][currentStrategies[i]];
        }
        
        // Get AI-optimized new allocations
        (address[] memory newStrategies, uint256[] memory newAllocations) = aiOptimizer.rebalanceStrategy(
            msg.sender,
            userDeposits[msg.sender],
            currentStrategies,
            currentAllocations,
            riskProfile
        );
        
        // Execute rebalancing
        for (uint256 i = 0; i < newStrategies.length; i++) {
            if (newAllocations[i] > currentAllocations[i]) {
                // Deposit more
                uint256 depositAmount = newAllocations[i] - currentAllocations[i];
                IERC20(aiOptimizer.getToken()).safeApprove(newStrategies[i], depositAmount);
                IStrategy(newStrategies[i]).deposit(depositAmount);
                strategyDeposits[msg.sender][newStrategies[i]] += depositAmount;
            } else if (newAllocations[i] < currentAllocations[i]) {
                // Withdraw excess
                uint256 withdrawAmount = currentAllocations[i] - newAllocations[i];
                uint256 withdrawn = IStrategy(newStrategies[i]).withdraw(withdrawAmount);
                strategyDeposits[msg.sender][newStrategies[i]] -= withdrawAmount;
            }
        }
    }
    
    function getStrategyInfo(
        address strategy
    ) external view returns (
        bool enabled,
        uint256 totalDeposits,
        uint256 apy,
        uint256 risk
    ) {
        return (
            isStrategyEnabled[strategy],
            IStrategy(strategy).totalDeposits(),
            IStrategy(strategy).getAPY(),
            IStrategy(strategy).getRisk()
        );
    }
    
    function getUserInfo(
        address user
    ) external view returns (
        uint256 totalDeposit,
        uint256 totalHarvested,
        uint256 lastHarvestTime
    ) {
        return (
            userDeposits[user],
            totalHarvested[user],
            lastHarvest[user]
        );
    }
    
    function getStrategyDeposits(
        address user,
        address strategy
    ) external view returns (uint256) {
        return strategyDeposits[user][strategy];
    }

    // Example: Stake funds (could be to a staking contract)
    function stake(uint256 amount) external {
        IERC20(aiOptimizer.getToken()).safeTransferFrom(msg.sender, address(this), amount);
        // TODO: Integrate with staking protocol
        emit Staked(msg.sender, amount);
    }

    // Example: Lend funds (could be to a lending protocol)
    function lend(uint256 amount) external {
        IERC20(aiOptimizer.getToken()).safeTransferFrom(msg.sender, address(this), amount);
        // TODO: Integrate with lending protocol
        emit Lent(msg.sender, amount);
    }

    // Example: Provide liquidity (could be to an AMM)
    function provideLiquidity(uint256 amount) external {
        IERC20(aiOptimizer.getToken()).safeTransferFrom(msg.sender, address(this), amount);
        // TODO: Integrate with LP protocol
        emit ProvidedLiquidity(msg.sender, amount);
    }

    // Example: Withdraw funds from a strategy
    function withdraw(uint256 amount, string calldata strategy) external onlyOwner {
        // TODO: Withdraw from the specified strategy
        IERC20(aiOptimizer.getToken()).safeTransfer(msg.sender, amount);
        emit Withdrawn(msg.sender, amount, strategy);
    }

    // Register a protocol adapter for a strategy
    function registerAdapter(string calldata strategy, address adapter) external onlyOwner {
        adapters[strategy] = adapter;
        emit AdapterRegistered(strategy, adapter);
    }

    // Deposit to a strategy via its adapter
    function depositToStrategy(string calldata strategy, uint256 amount) external {
        address adapter = adapters[strategy];
        require(adapter != address(0), "Adapter not registered");
        IERC20(aiOptimizer.getToken()).safeTransferFrom(msg.sender, address(this), amount);
        IERC20(aiOptimizer.getToken()).safeApprove(adapter, amount);
        IProtocolAdapter(adapter).deposit(amount);
        userStrategyBalances[msg.sender][strategy] += amount;
        emitDepositEvent(strategy, msg.sender, amount);
    }

    // Withdraw from a strategy via its adapter
    function withdrawFromStrategy(string calldata strategy, uint256 amount) external {
        address adapter = adapters[strategy];
        require(adapter != address(0), "Adapter not registered");
        require(userStrategyBalances[msg.sender][strategy] >= amount, "Insufficient balance");
        IProtocolAdapter(adapter).withdraw(amount);
        IERC20(aiOptimizer.getToken()).safeTransfer(msg.sender, amount);
        userStrategyBalances[msg.sender][strategy] -= amount;
        emit Withdrawn(msg.sender, amount, strategy);
    }

    // Automated rebalancing: move funds between strategies
    function rebalance(address user, string calldata fromStrategy, string calldata toStrategy, uint256 amount) external onlyOwner {
        require(userStrategyBalances[user][fromStrategy] >= amount, "Insufficient balance to rebalance");
        address fromAdapter = adapters[fromStrategy];
        address toAdapter = adapters[toStrategy];
        require(fromAdapter != address(0) && toAdapter != address(0), "Adapters not registered");
        // Withdraw from old strategy
        IProtocolAdapter(fromAdapter).withdraw(amount);
        IERC20(aiOptimizer.getToken()).safeTransfer(toAdapter, amount);
        IProtocolAdapter(toAdapter).deposit(amount);
        userStrategyBalances[user][fromStrategy] -= amount;
        userStrategyBalances[user][toStrategy] += amount;
        emit Rebalanced(user, fromStrategy, toStrategy, amount);
    }

    // Helper to emit the correct event for deposit
    function emitDepositEvent(string calldata strategy, address user, uint256 amount) internal {
        bytes32 s = keccak256(bytes(strategy));
        if (s == keccak256(bytes("staking"))) {
            emit Staked(user, amount);
        } else if (s == keccak256(bytes("lending"))) {
            emit Lent(user, amount);
        } else if (s == keccak256(bytes("lp"))) {
            emit ProvidedLiquidity(user, amount);
        }
    }

    // Optional: Add more strategies or allow owner to call arbitrary protocol adapters
} 