// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "./interfaces/IMoveVM.sol";
import "./libraries/MoveTypes.sol";

contract MoveRuntime is Ownable {
    // State
    address public moveVM;
    mapping(bytes32 => bool) public executedTransactions;
    
    // Events
    event MoveTransactionExecuted(bytes32 indexed txHash, bool success);
    event MoveVMUpdated(address indexed oldVM, address indexed newVM);
    
    constructor(address _moveVM) {
        moveVM = _moveVM;
    }
    
    function _executeMove(bytes memory txData) internal returns (bytes memory) {
        // Decode Move transaction
        (
            bytes32 moduleId,
            bytes32 functionId,
            bytes memory args,
            bytes memory typeArgs
        ) = abi.decode(txData, (bytes32, bytes32, bytes, bytes));
        
        // Check if transaction was already executed
        bytes32 txHash = keccak256(abi.encodePacked(moduleId, functionId, args, typeArgs));
        require(!executedTransactions[txHash], "Transaction already executed");
        
        // Execute Move transaction
        (bool success, bytes memory result) = moveVM.call(
            abi.encodeWithSelector(
                IMoveVM.execute.selector,
                moduleId,
                functionId,
                args,
                typeArgs
            )
        );
        require(success, "Move execution failed");
        
        // Mark transaction as executed
        executedTransactions[txHash] = true;
        
        emit MoveTransactionExecuted(txHash, success);
        
        return result;
    }
    
    function updateMoveVM(address _moveVM) external onlyOwner {
        address oldVM = moveVM;
        moveVM = _moveVM;
        emit MoveVMUpdated(oldVM, _moveVM);
    }
    
    function getMoveTransactionStatus(bytes32 txHash) external view returns (bool) {
        return executedTransactions[txHash];
    }
    
    function verifyMoveTransaction(
        bytes32 moduleId,
        bytes32 functionId,
        bytes memory args,
        bytes memory typeArgs
    ) external view returns (bool) {
        bytes32 txHash = keccak256(abi.encodePacked(moduleId, functionId, args, typeArgs));
        return executedTransactions[txHash];
    }
} 