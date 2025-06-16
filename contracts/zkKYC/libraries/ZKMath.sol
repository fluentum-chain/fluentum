// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library ZKMath {
    function hashToField(bytes memory data) internal pure returns (uint256) {
        return uint256(keccak256(data)) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
    }
    
    function hashToGroup(bytes memory data) internal pure returns (uint256[2] memory) {
        uint256 h = hashToField(data);
        return [h, h + 1];
    }
    
    function scalarMul(
        uint256[2] memory point,
        uint256 scalar
    ) internal pure returns (uint256[2] memory) {
        uint256[2] memory result;
        result[0] = (point[0] * scalar) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        result[1] = (point[1] * scalar) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        return result;
    }
    
    function pointAdd(
        uint256[2] memory a,
        uint256[2] memory b
    ) internal pure returns (uint256[2] memory) {
        uint256[2] memory result;
        result[0] = (a[0] + b[0]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        result[1] = (a[1] + b[1]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        return result;
    }
    
    function pointSub(
        uint256[2] memory a,
        uint256[2] memory b
    ) internal pure returns (uint256[2] memory) {
        uint256[2] memory result;
        result[0] = (a[0] - b[0]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        result[1] = (a[1] - b[1]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        return result;
    }
    
    function pointNeg(
        uint256[2] memory point
    ) internal pure returns (uint256[2] memory) {
        uint256[2] memory result;
        result[0] = (21888242871839275222246405745257275088548364400416034343698204186575808495617 - point[0]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        result[1] = (21888242871839275222246405745257275088548364400416034343698204186575808495617 - point[1]) % 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        return result;
    }
    
    function isOnCurve(
        uint256[2] memory point
    ) internal pure returns (bool) {
        uint256 x = point[0];
        uint256 y = point[1];
        uint256 p = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
        
        // Check if point satisfies y^2 = x^3 + 3
        uint256 y2 = (y * y) % p;
        uint256 x3 = (x * x * x) % p;
        uint256 x3plus3 = (x3 + 3) % p;
        
        return y2 == x3plus3;
    }
    
    function modExp(
        uint256 base,
        uint256 exponent,
        uint256 modulus
    ) internal pure returns (uint256) {
        uint256 result = 1;
        base = base % modulus;
        
        while (exponent > 0) {
            if (exponent % 2 == 1) {
                result = (result * base) % modulus;
            }
            base = (base * base) % modulus;
            exponent = exponent / 2;
        }
        
        return result;
    }
} 