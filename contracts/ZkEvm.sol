// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "./interfaces/IZkVerifier.sol";
import "./libraries/StateTree.sol";

contract ZkEvm is Ownable, ReentrancyGuard {
    using StateTree for bytes32;

    // State
    bytes32 public currentStateRoot;
    mapping(bytes32 => bool) public processedProofs;
    
    // Configuration
    address public zkVerifier;
    uint256 public maxGasLimit;
    uint256 public proofTimeout;
    
    // Events
    event StateRootUpdated(bytes32 indexed oldRoot, bytes32 indexed newRoot);
    event ProofVerified(bytes32 indexed proofHash);
    event ExecutionCompleted(bytes32 indexed txHash, bool success);
    
    constructor(
        address _zkVerifier,
        uint256 _maxGasLimit,
        uint256 _proofTimeout
    ) {
        zkVerifier = _zkVerifier;
        maxGasLimit = _maxGasLimit;
        proofTimeout = _proofTimeout;
    }
    
    function _executeEVM(bytes memory txData) internal returns (bytes memory) {
        // Decode transaction data
        (address to, uint256 value, bytes memory data) = abi.decode(
            txData,
            (address, uint256, bytes)
        );
        
        // Execute transaction
        (bool success, bytes memory result) = to.call{value: value, gas: maxGasLimit}(data);
        require(success, "EVM execution failed");
        
        return result;
    }
    
    function _commitState(bytes memory result) internal {
        // Extract new state root and proof from result
        (bytes32 newStateRoot, bytes memory proof) = abi.decode(
            result,
            (bytes32, bytes)
        );
        
        // Verify proof hasn't been processed
        bytes32 proofHash = keccak256(proof);
        require(!processedProofs[proofHash], "Proof already processed");
        
        // Verify ZK proof
        require(
            IZkVerifier(zkVerifier).verifyProof(
                currentStateRoot,
                newStateRoot,
                proof
            ),
            "Invalid ZK proof"
        );
        
        // Update state
        bytes32 oldRoot = currentStateRoot;
        currentStateRoot = newStateRoot;
        processedProofs[proofHash] = true;
        
        emit StateRootUpdated(oldRoot, newStateRoot);
        emit ProofVerified(proofHash);
    }
    
    function updateConfig(
        address _zkVerifier,
        uint256 _maxGasLimit,
        uint256 _proofTimeout
    ) external onlyOwner {
        zkVerifier = _zkVerifier;
        maxGasLimit = _maxGasLimit;
        proofTimeout = _proofTimeout;
    }
    
    function getStateProof(bytes32 stateRoot) external view returns (bytes memory) {
        return StateTree.getProof(stateRoot);
    }
    
    function verifyStateTransition(
        bytes32 oldRoot,
        bytes32 newRoot,
        bytes memory proof
    ) external view returns (bool) {
        return IZkVerifier(zkVerifier).verifyProof(oldRoot, newRoot, proof);
    }
} 