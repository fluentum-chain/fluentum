// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library StakingMath {
    function calculateRewardShare(
        uint256 totalRewards,
        uint256 userStake,
        uint256 totalStake
    ) internal pure returns (uint256) {
        if (totalStake == 0) return 0;
        return (totalRewards * userStake) / totalStake;
    }
    
    function calculateCommission(
        uint256 amount,
        uint256 commissionRate
    ) internal pure returns (uint256) {
        return (amount * commissionRate) / 100;
    }
    
    function calculateEffectiveStake(
        uint256 selfStake,
        uint256 delegatedStake
    ) internal pure returns (uint256) {
        return selfStake + delegatedStake;
    }
    
    function calculateVotingPower(
        uint256 stake,
        uint256 totalStake
    ) internal pure returns (uint256) {
        if (totalStake == 0) return 0;
        return (stake * 100) / totalStake;
    }
    
    function calculateSlashingPenalty(
        uint256 stake,
        uint256 severity
    ) internal pure returns (uint256) {
        require(severity <= 100, "Invalid severity");
        return (stake * severity) / 100;
    }
    
    function calculateUnbondingPeriod(
        uint256 stake,
        uint256 minStake
    ) internal pure returns (uint256) {
        if (stake <= minStake) return 7 days;
        if (stake <= minStake * 2) return 14 days;
        return 21 days;
    }
    
    function calculateRewardMultiplier(
        uint256 stake,
        uint256 minStake
    ) internal pure returns (uint256) {
        if (stake <= minStake) return 100;
        if (stake <= minStake * 2) return 120;
        if (stake <= minStake * 5) return 150;
        return 200;
    }
} 