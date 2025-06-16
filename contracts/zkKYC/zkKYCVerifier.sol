// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IZKVerifier.sol";
import "./interfaces/IKYCRegistry.sol";
import "./libraries/ZKMath.sol";

contract zkKYCVerifier is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;

    // Constants
    uint256 public constant MIN_KYC_LEVEL = 1;
    uint256 public constant MAX_KYC_LEVEL = 3;
    uint256 public constant VERIFICATION_TIMEOUT = 1 hours;
    
    // State
    IZKVerifier public immutable zkVerifier;
    IKYCRegistry public immutable kycRegistry;
    mapping(bytes32 => bool) public verifiedProofs;
    mapping(address => uint256) public lastVerification;
    mapping(address => uint256) public kycLevel;
    mapping(address => bool) public isWhitelisted;
    
    // Events
    event KYCVerified(
        address indexed user,
        uint256 kycLevel,
        bytes32 proofHash
    );
    event KYCLevelUpdated(
        address indexed user,
        uint256 oldLevel,
        uint256 newLevel
    );
    event WhitelistUpdated(
        address indexed user,
        bool status
    );
    event VerificationTimeoutUpdated(uint256 newTimeout);
    
    constructor(
        address _zkVerifier,
        address _kycRegistry
    ) {
        require(_zkVerifier != address(0), "Invalid verifier");
        require(_kycRegistry != address(0), "Invalid registry");
        
        zkVerifier = IZKVerifier(_zkVerifier);
        kycRegistry = IKYCRegistry(_kycRegistry);
    }
    
    function verifyTransaction(
        bytes calldata zkProof,
        uint256 minBalance,
        uint256 minKycLevel
    ) external nonReentrant returns (bool) {
        require(minKycLevel >= MIN_KYC_LEVEL && minKycLevel <= MAX_KYC_LEVEL, "Invalid KYC level");
        require(block.timestamp >= lastVerification[msg.sender] + VERIFICATION_TIMEOUT, "Verification too soon");
        
        // Verify ZK proof
        bytes32 proofHash = keccak256(zkProof);
        require(!verifiedProofs[proofHash], "Proof already used");
        
        bool isValid = zkVerifier.verifyProof(
            zkProof,
            [minBalance, minKycLevel]
        );
        require(isValid, "Invalid proof");
        
        // Update state
        verifiedProofs[proofHash] = true;
        lastVerification[msg.sender] = block.timestamp;
        kycLevel[msg.sender] = minKycLevel;
        
        emit KYCVerified(msg.sender, minKycLevel, proofHash);
        
        return true;
    }
    
    function updateKYCLevel(
        address user,
        uint256 newLevel
    ) external onlyOwner {
        require(newLevel >= MIN_KYC_LEVEL && newLevel <= MAX_KYC_LEVEL, "Invalid level");
        require(user != address(0), "Invalid user");
        
        uint256 oldLevel = kycLevel[user];
        kycLevel[user] = newLevel;
        
        emit KYCLevelUpdated(user, oldLevel, newLevel);
    }
    
    function updateWhitelist(
        address user,
        bool status
    ) external onlyOwner {
        require(user != address(0), "Invalid user");
        
        isWhitelisted[user] = status;
        
        emit WhitelistUpdated(user, status);
    }
    
    function updateVerificationTimeout(
        uint256 newTimeout
    ) external onlyOwner {
        require(newTimeout > 0, "Invalid timeout");
        
        VERIFICATION_TIMEOUT = newTimeout;
        
        emit VerificationTimeoutUpdated(newTimeout);
    }
    
    function verifyKYCStatus(
        address user,
        uint256 requiredLevel
    ) external view returns (bool) {
        return kycLevel[user] >= requiredLevel && isWhitelisted[user];
    }
    
    function getVerificationInfo(
        address user
    ) external view returns (
        uint256 level,
        uint256 lastVerified,
        bool whitelisted
    ) {
        return (
            kycLevel[user],
            lastVerification[user],
            isWhitelisted[user]
        );
    }
    
    function isProofVerified(
        bytes32 proofHash
    ) external view returns (bool) {
        return verifiedProofs[proofHash];
    }
} 