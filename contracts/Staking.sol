// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "./interfaces/IFLUMXToken.sol";
import "./libraries/StakingMath.sol";

contract FluentumStaking is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;
    using StakingMath for uint256;

    // Constants
    uint256 public constant MIN_VALIDATOR_STAKE = 50_000 * 10**18; // 50k FLUMX
    uint256 public constant MIN_DELEGATION = 100 * 10**18; // 100 FLUMX
    uint256 public constant MAX_COMMISSION_RATE = 20; // 20%
    uint256 public constant GAS_REFUND_THRESHOLD = 10_000 * 10**18; // 10k FLUMX
    uint256 public constant GAS_REFUND_MULTIPLIER = 100000;
    
    // State
    IFLUMXToken public immutable fluxToken;
    Validator[] public validators;
    mapping(address => uint256) public delegations;
    mapping(uint256 => mapping(address => uint256)) public validatorDelegations;
    mapping(uint256 => uint256) public validatorRewards;
    mapping(uint256 => uint256) public lastRewardUpdate;
    
    // Structs
    struct Validator {
        address owner;
        uint256 stakedAmount;
        uint256 commissionRate;
        uint256 totalDelegated;
        uint256 lastRewardPerToken;
        bool active;
        uint256 activationEpoch;
    }
    
    struct DelegationInfo {
        uint256 amount;
        uint256 rewardDebt;
        uint256 lastClaim;
    }
    
    // Events
    event ValidatorCreated(uint256 indexed validatorId, address indexed owner, uint256 amount, uint256 commission);
    event ValidatorDeactivated(uint256 indexed validatorId);
    event DelegationAdded(uint256 indexed validatorId, address indexed delegator, uint256 amount);
    event DelegationRemoved(uint256 indexed validatorId, address indexed delegator, uint256 amount);
    event RewardsDistributed(uint256 indexed validatorId, uint256 amount);
    event RewardsClaimed(uint256 indexed validatorId, address indexed delegator, uint256 amount);
    event GasRefunded(address indexed user, uint256 amount);
    
    constructor(address _fluxToken) {
        require(_fluxToken != address(0), "Invalid token address");
        fluxToken = IFLUMXToken(_fluxToken);
    }
    
    function createValidator(uint256 amount, uint256 commission) external nonReentrant {
        require(amount >= MIN_VALIDATOR_STAKE, "Insufficient stake amount");
        require(commission <= MAX_COMMISSION_RATE, "Commission too high");
        
        // Transfer stake
        fluxToken.safeTransferFrom(msg.sender, address(this), amount);
        
        // Create validator
        uint256 validatorId = validators.length;
        validators.push(Validator({
            owner: msg.sender,
            stakedAmount: amount,
            commissionRate: commission,
            totalDelegated: 0,
            lastRewardPerToken: 0,
            active: true,
            activationEpoch: block.number
        }));
        
        emit ValidatorCreated(validatorId, msg.sender, amount, commission);
    }
    
    function delegate(uint256 validatorId, uint256 amount) external nonReentrant {
        require(validatorId < validators.length, "Invalid validator");
        require(amount >= MIN_DELEGATION, "Insufficient delegation amount");
        Validator storage validator = validators[validatorId];
        require(validator.active, "Validator inactive");
        
        // Transfer delegation
        fluxToken.safeTransferFrom(msg.sender, address(this), amount);
        
        // Update delegations
        delegations[msg.sender] += amount;
        validatorDelegations[validatorId][msg.sender] += amount;
        validator.totalDelegated += amount;
        
        // Process gas refund for large delegations
        if (amount >= GAS_REFUND_THRESHOLD) {
            _refundGas(msg.sender);
        }
        
        emit DelegationAdded(validatorId, msg.sender, amount);
    }
    
    function undelegate(uint256 validatorId, uint256 amount) external nonReentrant {
        require(validatorId < validators.length, "Invalid validator");
        Validator storage validator = validators[validatorId];
        require(validator.active, "Validator inactive");
        
        uint256 currentDelegation = validatorDelegations[validatorId][msg.sender];
        require(currentDelegation >= amount, "Insufficient delegation");
        
        // Update delegations
        delegations[msg.sender] -= amount;
        validatorDelegations[validatorId][msg.sender] -= amount;
        validator.totalDelegated -= amount;
        
        // Transfer tokens back
        fluxToken.safeTransfer(msg.sender, amount);
        
        emit DelegationRemoved(validatorId, msg.sender, amount);
    }
    
    function distributeRewards(uint256 amount) external onlyOwner {
        require(amount > 0, "Invalid reward amount");
        
        // Transfer rewards
        fluxToken.safeTransferFrom(msg.sender, address(this), amount);
        
        // Distribute to validators based on stake
        uint256 totalStake = _getTotalStake();
        require(totalStake > 0, "No active stake");
        
        for (uint256 i = 0; i < validators.length; i++) {
            Validator storage validator = validators[i];
            if (!validator.active) continue;
            
            uint256 validatorShare = (amount * (validator.stakedAmount + validator.totalDelegated)) / totalStake;
            validatorRewards[i] += validatorShare;
            
            emit RewardsDistributed(i, validatorShare);
        }
    }
    
    function claimRewards(uint256 validatorId) external nonReentrant {
        require(validatorId < validators.length, "Invalid validator");
        Validator storage validator = validators[validatorId];
        require(validator.active, "Validator inactive");
        
        uint256 delegation = validatorDelegations[validatorId][msg.sender];
        require(delegation > 0, "No delegation found");
        
        // Calculate rewards
        uint256 totalRewards = validatorRewards[validatorId];
        uint256 delegatorShare = (totalRewards * delegation) / validator.totalDelegated;
        
        // Apply commission
        uint256 commission = (delegatorShare * validator.commissionRate) / 100;
        uint256 reward = delegatorShare - commission;
        
        // Update state
        validatorRewards[validatorId] -= delegatorShare;
        lastRewardUpdate[validatorId] = block.number;
        
        // Transfer rewards
        fluxToken.safeTransfer(msg.sender, reward);
        fluxToken.safeTransfer(validator.owner, commission);
        
        emit RewardsClaimed(validatorId, msg.sender, reward);
    }
    
    function deactivateValidator(uint256 validatorId) external {
        require(validatorId < validators.length, "Invalid validator");
        Validator storage validator = validators[validatorId];
        require(msg.sender == validator.owner, "Not validator owner");
        require(validator.active, "Already inactive");
        
        validator.active = false;
        
        emit ValidatorDeactivated(validatorId);
    }
    
    function _refundGas(address user) internal {
        uint256 refundAmount = tx.gasprice * GAS_REFUND_MULTIPLIER;
        fluxToken.mint(user, refundAmount);
        
        emit GasRefunded(user, refundAmount);
    }
    
    function _getTotalStake() internal view returns (uint256) {
        uint256 total = 0;
        for (uint256 i = 0; i < validators.length; i++) {
            Validator storage validator = validators[i];
            if (validator.active) {
                total += validator.stakedAmount + validator.totalDelegated;
            }
        }
        return total;
    }
    
    function getValidatorInfo(uint256 validatorId) external view returns (
        address owner,
        uint256 stakedAmount,
        uint256 commissionRate,
        uint256 totalDelegated,
        bool active,
        uint256 activationEpoch
    ) {
        require(validatorId < validators.length, "Invalid validator");
        Validator storage validator = validators[validatorId];
        return (
            validator.owner,
            validator.stakedAmount,
            validator.commissionRate,
            validator.totalDelegated,
            validator.active,
            validator.activationEpoch
        );
    }
    
    function getDelegationInfo(uint256 validatorId, address delegator) external view returns (
        uint256 amount,
        uint256 pendingRewards
    ) {
        require(validatorId < validators.length, "Invalid validator");
        Validator storage validator = validators[validatorId];
        require(validator.active, "Validator inactive");
        
        uint256 delegation = validatorDelegations[validatorId][delegator];
        if (delegation == 0) return (0, 0);
        
        uint256 totalRewards = validatorRewards[validatorId];
        uint256 delegatorShare = (totalRewards * delegation) / validator.totalDelegated;
        uint256 commission = (delegatorShare * validator.commissionRate) / 100;
        
        return (delegation, delegatorShare - commission);
    }
} 