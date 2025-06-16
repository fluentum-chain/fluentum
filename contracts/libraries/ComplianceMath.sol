// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library ComplianceMath {
    uint256 public constant PRECISION = 1e18;
    uint256 public constant MAX_RISK_SCORE = 100;
    uint256 public constant MIN_CONFIDENCE = 80;
    
    function calculateUSDValue(
        uint256 amount,
        uint256 price
    ) internal pure returns (uint256) {
        return (amount * price) / PRECISION;
    }
    
    function calculateRiskScore(
        uint256 amount,
        uint256 usdValue,
        uint256 sanctionsScore,
        uint256 pepScore,
        uint256 amountScore
    ) internal pure returns (uint256) {
        uint256 weightedScore = (
            sanctionsScore * 40 +
            pepScore * 30 +
            amountScore * 30
        ) / 100;
        
        return weightedScore > MAX_RISK_SCORE ? MAX_RISK_SCORE : weightedScore;
    }
    
    function calculateConfidence(
        uint256 dataQuality,
        uint256 providerReliability,
        uint256 timeSinceUpdate
    ) internal pure returns (uint256) {
        uint256 timeFactor = timeSinceUpdate > 30 days ? 80 : 100;
        
        return (dataQuality * providerReliability * timeFactor) / (100 * 100);
    }
    
    function calculateAmountScore(
        uint256 amount,
        uint256 threshold
    ) internal pure returns (uint256) {
        if (amount <= threshold) return 0;
        
        uint256 excess = amount - threshold;
        uint256 score = (excess * 100) / threshold;
        
        return score > MAX_RISK_SCORE ? MAX_RISK_SCORE : score;
    }
    
    function calculateSanctionsScore(
        bool isSanctioned,
        uint256 matchConfidence
    ) internal pure returns (uint256) {
        if (!isSanctioned) return 0;
        return (matchConfidence * MAX_RISK_SCORE) / 100;
    }
    
    function calculatePEPScore(
        bool isPEP,
        uint256 matchConfidence
    ) internal pure returns (uint256) {
        if (!isPEP) return 0;
        return (matchConfidence * MAX_RISK_SCORE) / 100;
    }
    
    function calculateBatchRiskScore(
        uint256[] memory riskScores,
        uint256[] memory amounts
    ) internal pure returns (uint256) {
        require(riskScores.length == amounts.length, "Length mismatch");
        
        uint256 totalAmount = 0;
        uint256 weightedScore = 0;
        
        for (uint256 i = 0; i < amounts.length; i++) {
            totalAmount += amounts[i];
            weightedScore += riskScores[i] * amounts[i];
        }
        
        if (totalAmount == 0) return 0;
        return weightedScore / totalAmount;
    }
    
    function calculateJurisdictionRisk(
        string memory jurisdiction,
        uint256 baseRisk
    ) internal pure returns (uint256) {
        // Implementation depends on jurisdiction risk mapping
        return baseRisk;
    }
    
    function calculateTimeBasedRisk(
        uint256 baseRisk,
        uint256 timeSinceLastScreening
    ) internal pure returns (uint256) {
        if (timeSinceLastScreening <= 1 days) return baseRisk;
        if (timeSinceLastScreening <= 7 days) return (baseRisk * 110) / 100;
        if (timeSinceLastScreening <= 30 days) return (baseRisk * 120) / 100;
        return (baseRisk * 150) / 100;
    }
} 