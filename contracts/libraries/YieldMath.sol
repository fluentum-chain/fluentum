// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library YieldMath {
    uint256 public constant PRECISION = 1e18;
    uint256 public constant MAX_APY = 1000 * PRECISION; // 1000%
    uint256 public constant MAX_RISK = 100 * PRECISION; // 100%
    
    function calculateAPY(
        uint256 principal,
        uint256 interest,
        uint256 timeElapsed
    ) internal pure returns (uint256) {
        if (timeElapsed == 0) return 0;
        return (interest * PRECISION * 365 days) / (principal * timeElapsed);
    }
    
    function calculateRiskAdjustedReturn(
        uint256 apy,
        uint256 risk
    ) internal pure returns (uint256) {
        return (apy * (PRECISION - risk)) / PRECISION;
    }
    
    function calculateOptimalAllocation(
        uint256[] memory apys,
        uint256[] memory risks,
        uint256 totalAmount
    ) internal pure returns (uint256[] memory) {
        uint256 length = apys.length;
        require(length == risks.length, "Length mismatch");
        
        uint256[] memory riskAdjustedReturns = new uint256[](length);
        uint256 totalRiskAdjusted = 0;
        
        // Calculate risk-adjusted returns
        for (uint256 i = 0; i < length; i++) {
            riskAdjustedReturns[i] = calculateRiskAdjustedReturn(apys[i], risks[i]);
            totalRiskAdjusted += riskAdjustedReturns[i];
        }
        
        // Calculate optimal allocations
        uint256[] memory allocations = new uint256[](length);
        for (uint256 i = 0; i < length; i++) {
            allocations[i] = (totalAmount * riskAdjustedReturns[i]) / totalRiskAdjusted;
        }
        
        return allocations;
    }
    
    function calculateHarvestAmount(
        uint256 principal,
        uint256 apy,
        uint256 timeElapsed
    ) internal pure returns (uint256) {
        return (principal * apy * timeElapsed) / (PRECISION * 365 days);
    }
    
    function calculateSlippage(
        uint256 expected,
        uint256 actual
    ) internal pure returns (uint256) {
        if (expected == 0) return 0;
        return (expected > actual) ? 
            ((expected - actual) * PRECISION) / expected :
            ((actual - expected) * PRECISION) / expected;
    }
    
    function calculateUtilization(
        uint256 totalDeposits,
        uint256 totalCapacity
    ) internal pure returns (uint256) {
        if (totalCapacity == 0) return 0;
        return (totalDeposits * PRECISION) / totalCapacity;
    }
    
    function calculateCompoundedAPY(
        uint256 baseAPY,
        uint256 compoundFrequency
    ) internal pure returns (uint256) {
        if (compoundFrequency == 0) return baseAPY;
        return PRECISION * (
            (PRECISION + (baseAPY / compoundFrequency)) ** compoundFrequency - PRECISION
        );
    }
} 