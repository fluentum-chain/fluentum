// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";

interface IERC20 {
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
    function transfer(address to, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
    function burn(uint256 amount) external;
}

contract FluentumGasStation is Ownable, ReentrancyGuard {
    IERC20 public immutable FLUMX_TOKEN;
    uint256 public minStake = 500 ether; // 500 FLUMX

    mapping(address => uint256) public stakedAmount;
    mapping(address => bool) public isStaker;
    address[] public stakers;
    mapping(address => uint256) private stakerIndex; // For efficient removal

    event Staked(address indexed user, uint256 amount);
    event Unstaked(address indexed user, uint256 amount);
    event GasFeeCharged(address indexed user, uint256 fee);
    event Subsidized(address indexed staker, bytes txData);

    constructor(address fluxTokenAddress) {
        FLUMX_TOKEN = IERC20(fluxTokenAddress);
    }

    function stake(uint256 amount) external nonReentrant {
        require(amount >= minStake, "Insufficient stake");
        FLUMX_TOKEN.transferFrom(msg.sender, address(this), amount);
        stakedAmount[msg.sender] += amount;
        if (!isStaker[msg.sender]) {
            isStaker[msg.sender] = true;
            stakerIndex[msg.sender] = stakers.length;
            stakers.push(msg.sender);
        }
        emit Staked(msg.sender, amount);
    }

    function unstake(uint256 amount) external nonReentrant {
        require(stakedAmount[msg.sender] >= amount, "Not enough staked");
        stakedAmount[msg.sender] -= amount;
        if (stakedAmount[msg.sender] < minStake && isStaker[msg.sender]) {
            isStaker[msg.sender] = false;
            // Efficiently remove from stakers array
            uint256 idx = stakerIndex[msg.sender];
            uint256 lastIdx = stakers.length - 1;
            if (idx != lastIdx) {
                address lastStaker = stakers[lastIdx];
                stakers[idx] = lastStaker;
                stakerIndex[lastStaker] = idx;
            }
            stakers.pop();
            delete stakerIndex[msg.sender];
        }
        FLUMX_TOKEN.transfer(msg.sender, amount);
        emit Unstaked(msg.sender, amount);
    }

    function processTransaction(address user, bytes calldata txData) external nonReentrant {
        if (isStaker[user]) {
            // Execute with gas subsidies
            (bool success, ) = address(this).call{gas: gasleft()}(txData);
            require(success, "Transaction failed");
            emit Subsidized(user, txData);
        } else {
            // Charge normal gas fees
            uint256 gasFee = calculateGasFee(txData);
            FLUMX_TOKEN.transferFrom(user, address(this), gasFee);
            (bool success, ) = address(this).call(txData);
            require(success, "Transaction failed");
            emit GasFeeCharged(user, gasFee);
        }
    }

    function calculateGasFee(bytes calldata /*txData*/) public pure returns (uint256) {
        // Implement your gas fee logic here
        return 1 ether; // Placeholder
    }

    function _redistributeFees() internal onlyOwner {
        uint256 fees = FLUMX_TOKEN.balanceOf(address(this));
        uint256 toBurn = fees / 2;
        uint256 toDistribute = fees - toBurn;
        FLUMX_TOKEN.burn(toBurn);
        _distributeToStakers(toDistribute);
    }

    function _distributeToStakers(uint256 amount) internal {
        uint256 totalStakers = stakers.length;
        if (totalStakers == 0) return;
        uint256 share = amount / totalStakers;
        for (uint i = 0; i < totalStakers; i++) {
            FLUMX_TOKEN.transfer(stakers[i], share);
        }
    }

    // Optional: allow owner to update minStake
    function setMinStake(uint256 newMinStake) external onlyOwner {
        minStake = newMinStake;
    }
} 