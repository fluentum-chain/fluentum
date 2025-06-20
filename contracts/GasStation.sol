// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IFLUXToken.sol";
import "./interfaces/IStaking.sol";

contract GasStation is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;

    // Constants
    uint256 public constant MIN_STAKE = 500 * 10**18; // 500 FLUX
    uint256 public constant GAS_REFUND_MULTIPLIER = 100000;
    uint256 public constant MAX_GAS_REFUND = 1 ether;
    
    // State
    IFLUXToken public immutable fluxToken;
    IStaking public immutable staking;
    mapping(address => bool) public isStaker;
    mapping(address => uint256) public stakerStakes;
    mapping(address => uint256) public lastGasRefund;
    mapping(address => uint256) public totalGasRefunded;
    
    // Events
    event StakerRegistered(address indexed staker, uint256 amount);
    event StakerUnregistered(address indexed staker);
    event GasRefunded(address indexed user, uint256 amount);
    event GasRefundLimitUpdated(uint256 newLimit);
    event GasRefundMultiplierUpdated(uint256 newMultiplier);
    
    constructor(address _fluxToken, address _staking) {
        require(_fluxToken != address(0), "Invalid token address");
        require(_staking != address(0), "Invalid staking address");
        fluxToken = IFLUXToken(_fluxToken);
        staking = IStaking(_staking);
    }
    
    modifier zeroGas() {
        if (isStaker[msg.sender]) {
            uint256 gasStart = gasleft();
            _;
            uint256 gasUsed = gasStart - gasleft();
            _refundGas(gasUsed);
        } else {
            _;
        }
    }
    
    function registerStaker() external nonReentrant {
        require(!isStaker[msg.sender], "Already registered");
        
        // Check delegation in staking contract
        uint256 delegation = staking.getDelegationInfo(0, msg.sender); // Assuming validator ID 0
        require(delegation >= MIN_STAKE, "Insufficient stake");
        
        isStaker[msg.sender] = true;
        stakerStakes[msg.sender] = delegation;
        
        emit StakerRegistered(msg.sender, delegation);
    }
    
    function unregisterStaker() external nonReentrant {
        require(isStaker[msg.sender], "Not registered");
        
        isStaker[msg.sender] = false;
        stakerStakes[msg.sender] = 0;
        
        emit StakerUnregistered(msg.sender);
    }
    
    function _refundGas(uint256 gasUsed) internal {
        uint256 refund = (gasUsed + 21000) * tx.gasprice * GAS_REFUND_MULTIPLIER;
        refund = refund > MAX_GAS_REFUND ? MAX_GAS_REFUND : refund;
        
        // Check if enough time has passed since last refund
        require(block.timestamp >= lastGasRefund[msg.sender] + 1 hours, "Refund too soon");
        
        // Update state
        lastGasRefund[msg.sender] = block.timestamp;
        totalGasRefunded[msg.sender] += refund;
        
        // Mint FLUX tokens as refund
        fluxToken.mint(msg.sender, refund);
        
        emit GasRefunded(msg.sender, refund);
    }
    
    function updateGasRefundLimit(uint256 newLimit) external onlyOwner {
        require(newLimit > 0, "Invalid limit");
        MAX_GAS_REFUND = newLimit;
        emit GasRefundLimitUpdated(newLimit);
    }
    
    function updateGasRefundMultiplier(uint256 newMultiplier) external onlyOwner {
        require(newMultiplier > 0, "Invalid multiplier");
        GAS_REFUND_MULTIPLIER = newMultiplier;
        emit GasRefundMultiplierUpdated(newMultiplier);
    }
    
    function getStakerInfo(address staker) external view returns (
        bool registered,
        uint256 stake,
        uint256 lastRefund,
        uint256 totalRefunded
    ) {
        return (
            isStaker[staker],
            stakerStakes[staker],
            lastGasRefund[staker],
            totalGasRefunded[staker]
        );
    }
    
    function calculateGasRefund(uint256 gasUsed) external view returns (uint256) {
        uint256 refund = (gasUsed + 21000) * tx.gasprice * GAS_REFUND_MULTIPLIER;
        return refund > MAX_GAS_REFUND ? MAX_GAS_REFUND : refund;
    }
    
    function isEligibleForRefund(address user) external view returns (bool) {
        return isStaker[user] && block.timestamp >= lastGasRefund[user] + 1 hours;
    }
} 