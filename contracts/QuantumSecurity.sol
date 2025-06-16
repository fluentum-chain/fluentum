// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "./interfaces/IQuantumVerifier.sol";
import "./libraries/QuantumMath.sol";

contract QuantumSecurity is ReentrancyGuard, Ownable {
    // Constants
    uint256 public constant MAX_SIGNATURE_SIZE = 2700; // Dilithium3 signature size
    uint256 public constant MAX_PUBLIC_KEY_SIZE = 1472; // Dilithium3 public key size
    uint256 public constant MAX_MESSAGE_SIZE = 32; // SHA-256 hash size
    uint256 public constant SIGNATURE_TIMEOUT = 1 hours;
    
    // State
    IQuantumVerifier public immutable verifier;
    mapping(address => bytes) public quantumPublicKeys;
    mapping(bytes32 => bool) public usedSignatures;
    mapping(address => uint256) public lastSignatureTimestamp;
    
    // Events
    event PublicKeyRegistered(
        address indexed user,
        bytes publicKey
    );
    event PublicKeyUpdated(
        address indexed user,
        bytes publicKey
    );
    event SignatureVerified(
        address indexed user,
        bytes32 messageHash,
        bool isValid
    );
    event VerifierUpdated(address indexed verifier);
    
    constructor(address _verifier) {
        require(_verifier != address(0), "Invalid verifier");
        verifier = IQuantumVerifier(_verifier);
    }
    
    function registerPublicKey(
        bytes calldata publicKey,
        bytes calldata signature
    ) external nonReentrant {
        require(publicKey.length <= MAX_PUBLIC_KEY_SIZE, "Public key too large");
        require(signature.length <= MAX_SIGNATURE_SIZE, "Signature too large");
        
        // Verify signature of public key
        bytes32 messageHash = keccak256(
            abi.encodePacked(
                "Register public key:",
                msg.sender,
                publicKey
            )
        );
        
        require(
            verifier.verifySignature(
                publicKey,
                messageHash,
                signature
            ),
            "Invalid signature"
        );
        
        // Register public key
        quantumPublicKeys[msg.sender] = publicKey;
        
        emit PublicKeyRegistered(msg.sender, publicKey);
    }
    
    function updatePublicKey(
        bytes calldata newPublicKey,
        bytes calldata signature
    ) external nonReentrant {
        require(newPublicKey.length <= MAX_PUBLIC_KEY_SIZE, "Public key too large");
        require(signature.length <= MAX_SIGNATURE_SIZE, "Signature too large");
        require(quantumPublicKeys[msg.sender].length > 0, "No public key registered");
        
        // Verify signature of new public key
        bytes32 messageHash = keccak256(
            abi.encodePacked(
                "Update public key:",
                msg.sender,
                newPublicKey
            )
        );
        
        require(
            verifier.verifySignature(
                newPublicKey,
                messageHash,
                signature
            ),
            "Invalid signature"
        );
        
        // Update public key
        quantumPublicKeys[msg.sender] = newPublicKey;
        
        emit PublicKeyUpdated(msg.sender, newPublicKey);
    }
    
    function verifySignature(
        address user,
        bytes32 messageHash,
        bytes calldata signature
    ) external nonReentrant returns (bool) {
        require(signature.length <= MAX_SIGNATURE_SIZE, "Signature too large");
        require(quantumPublicKeys[user].length > 0, "No public key registered");
        
        // Check signature timeout
        require(
            block.timestamp <= lastSignatureTimestamp[user] + SIGNATURE_TIMEOUT,
            "Signature expired"
        );
        
        // Check signature reuse
        bytes32 sigHash = keccak256(signature);
        require(!usedSignatures[sigHash], "Signature already used");
        usedSignatures[sigHash] = true;
        
        // Verify signature
        bool isValid = verifier.verifySignature(
            quantumPublicKeys[user],
            messageHash,
            signature
        );
        
        if (isValid) {
            lastSignatureTimestamp[user] = block.timestamp;
        }
        
        emit SignatureVerified(user, messageHash, isValid);
        
        return isValid;
    }
    
    function verifyBatch(
        address[] calldata users,
        bytes32[] calldata messageHashes,
        bytes[] calldata signatures
    ) external nonReentrant returns (bool[] memory) {
        require(
            users.length == messageHashes.length &&
            messageHashes.length == signatures.length,
            "Length mismatch"
        );
        
        bool[] memory results = new bool[](users.length);
        
        for (uint256 i = 0; i < users.length; i++) {
            results[i] = this.verifySignature(
                users[i],
                messageHashes[i],
                signatures[i]
            );
        }
        
        return results;
    }
    
    function updateVerifier(address _verifier) external onlyOwner {
        require(_verifier != address(0), "Invalid verifier");
        
        verifier = IQuantumVerifier(_verifier);
        
        emit VerifierUpdated(_verifier);
    }
    
    function getPublicKey(
        address user
    ) external view returns (bytes memory) {
        return quantumPublicKeys[user];
    }
    
    function isSignatureUsed(
        bytes32 sigHash
    ) external view returns (bool) {
        return usedSignatures[sigHash];
    }
    
    function getLastSignatureTimestamp(
        address user
    ) external view returns (uint256) {
        return lastSignatureTimestamp[user];
    }
} 