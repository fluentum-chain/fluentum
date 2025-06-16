// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./ZkEvm.sol";
import "./MoveRuntime.sol";
import "./interfaces/IWasmRuntime.sol";
import "./libraries/VMType.sol";

contract FluentumExecution is ZkEvm, MoveRuntime {
    using VMType for VMType.Type;
    
    // State
    mapping(address => VMType.Type) public contractVM;
    address public wasmRuntime;
    
    // Events
    event ContractVMUpdated(address indexed contract, VMType.Type vmType);
    event WasmRuntimeUpdated(address indexed oldRuntime, address indexed newRuntime);
    event TransactionExecuted(
        address indexed to,
        VMType.Type vmType,
        bool success
    );
    
    constructor(
        address _zkVerifier,
        address _moveVM,
        address _wasmRuntime,
        uint256 _maxGasLimit,
        uint256 _proofTimeout
    )
        ZkEvm(_zkVerifier, _maxGasLimit, _proofTimeout)
        MoveRuntime(_moveVM)
    {
        wasmRuntime = _wasmRuntime;
    }
    
    function executeTransaction(bytes calldata txData) external nonReentrant {
        // Decode transaction
        (address to, bytes memory data) = abi.decode(txData, (address, bytes));
        
        // Get VM type for contract
        VMType.Type vm = contractVM[to];
        require(vm != VMType.Type.None, "Unknown VM type");
        
        bytes memory result;
        bool success;
        
        // Execute based on VM type
        if (vm == VMType.Type.EVM) {
            result = _executeEVM(abi.encode(to, 0, data));
            success = true;
        } else if (vm == VMType.Type.Move) {
            result = _executeMove(data);
            success = true;
        } else if (vm == VMType.Type.Wasm) {
            (success, result) = _executeWasm(data);
        }
        
        require(success, "Transaction execution failed");
        
        // Verify and commit state
        _commitState(result);
        
        emit TransactionExecuted(to, vm, success);
    }
    
    function _executeWasm(bytes memory data) internal returns (bool, bytes memory) {
        return IWasmRuntime(wasmRuntime).execute(data);
    }
    
    function setContractVM(address contract, VMType.Type vmType) external onlyOwner {
        require(vmType != VMType.Type.None, "Invalid VM type");
        contractVM[contract] = vmType;
        emit ContractVMUpdated(contract, vmType);
    }
    
    function updateWasmRuntime(address _wasmRuntime) external onlyOwner {
        address oldRuntime = wasmRuntime;
        wasmRuntime = _wasmRuntime;
        emit WasmRuntimeUpdated(oldRuntime, _wasmRuntime);
    }
    
    function getContractVM(address contract) external view returns (VMType.Type) {
        return contractVM[contract];
    }
    
    function verifyExecution(
        address to,
        bytes memory data,
        bytes memory proof
    ) external view returns (bool) {
        VMType.Type vm = contractVM[to];
        require(vm != VMType.Type.None, "Unknown VM type");
        
        if (vm == VMType.Type.EVM) {
            return verifyStateTransition(
                currentStateRoot,
                keccak256(abi.encodePacked(to, data)),
                proof
            );
        } else if (vm == VMType.Type.Move) {
            return verifyMoveTransaction(
                bytes32(0), // moduleId
                bytes32(0), // functionId
                data,
                bytes("")  // typeArgs
            );
        } else {
            return IWasmRuntime(wasmRuntime).verifyExecution(data, proof);
        }
    }
} 