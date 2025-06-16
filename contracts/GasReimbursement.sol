// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./interfaces/IFluentumVerifier.sol";
import "./interfaces/IRelayer.sol";
import "./libraries/GasMath.sol";

contract GasReimbursement is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;
    
    // Constants
    uint256 public constant MIN_GAS_REIMBURSEMENT = 21000;
    uint256 public constant MAX_GAS_REIMBURSEMENT = 1000000;
    uint256 public constant REIMBURSEMENT_MULTIPLIER = 110; // 110%
    
    // State
    IFluentumVerifier public immutable verifier;
    IRelayer public immutable relayer;
    mapping(bytes32 => bool) public executedTransactions;
    mapping(address => uint256) public userGasRefunds;
    
    // Events
    event TransactionExecuted(
        bytes32 indexed txId,
        address indexed user,
        bytes payload,
        uint256 gasUsed,
        uint256 gasRefunded
    );
    event GasRefunded(
        address indexed user,
        uint256 amount
    );
    event VerifierUpdated(address indexed verifier);
    event RelayerUpdated(address indexed relayer);
    
    constructor(address _verifier, address _relayer) {
        require(_verifier != address(0), "Invalid verifier");
        require(_relayer != address(0), "Invalid relayer");
        
        verifier = IFluentumVerifier(_verifier);
        relayer = IRelayer(_relayer);
    }
    
    function executeWithGas(
        bytes32 txId,
        address user,
        bytes calldata payload,
        bytes calldata proof
    ) external nonReentrant {
        require(!executedTransactions[txId], "Transaction already executed");
        require(user != address(0), "Invalid user");
        
        // Verify Fluentum proof
        require(
            verifier.verifyProof(
                txId,
                user,
                payload,
                proof
            ),
            "Invalid proof"
        );
        
        // Mark transaction as executed
        executedTransactions[txId] = true;
        
        // Execute transaction
        uint256 gasStart = gasleft();
        (bool success, ) = address(this).call(payload);
        require(success, "Execution failed");
        uint256 gasUsed = gasStart - gasleft();
        
        // Calculate gas refund
        uint256 gasRefund = GasMath.calculateGasRefund(
            gasUsed,
            tx.gasprice,
            REIMBURSEMENT_MULTIPLIER
        );
        
        // Update user's gas refund
        userGasRefunds[user] += gasRefund;
        
        emit TransactionExecuted(
            txId,
            user,
            payload,
            gasUsed,
            gasRefund
        );
    }
    
    function claimGasRefund() external nonReentrant {
        uint256 refund = userGasRefunds[msg.sender];
        require(refund > 0, "No refund available");
        
        userGasRefunds[msg.sender] = 0;
        
        // Transfer refund
        (bool success, ) = payable(msg.sender).call{value: refund}("");
        require(success, "Transfer failed");
        
        emit GasRefunded(msg.sender, refund);
    }
    
    function updateVerifier(address _verifier) external onlyOwner {
        require(_verifier != address(0), "Invalid verifier");
        
        verifier = IFluentumVerifier(_verifier);
        
        emit VerifierUpdated(_verifier);
    }
    
    function updateRelayer(address _relayer) external onlyOwner {
        require(_relayer != address(0), "Invalid relayer");
        
        relayer = IRelayer(_relayer);
        
        emit RelayerUpdated(_relayer);
    }
    
    function getGasRefund(
        address user
    ) external view returns (uint256) {
        return userGasRefunds[user];
    }
    
    function isTransactionExecuted(
        bytes32 txId
    ) external view returns (bool) {
        return executedTransactions[txId];
    }
    
    receive() external payable {}
} 