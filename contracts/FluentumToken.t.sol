// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "ds-test/test.sol";
import "./FluentumToken.sol";

contract FLUMXTokenTest is DSTest {
    FLUMXToken token;
    address owner = address(this);
    address treasury = address(0xBEEF);
    address validator = address(0xCAFE);
    address notValidator = address(0xBAD);

    function setUp() public {
        token = new FLUMXToken();
        token.setQuantumTreasury(treasury);
    }

    function testInitialSupply() public {
        assertEq(token.totalSupply(), token.INITIAL_SUPPLY());
        assertEq(token.decimals(), 9);
    }

    function testMaxSupply() public {
        assertEq(token.MAX_SUPPLY(), 1_000_000_000 * 10**9);
    }

    function testEmitToTreasuryOwner() public {
        uint256 before = token.totalSupply();
        token.emitToTreasury();
        assertEq(token.totalSupply(), before + token.emissionRate());
        assertEq(token.balanceOf(treasury), token.emissionRate());
    }

    function testEmitToTreasuryValidator() public {
        token.addQuantumValidator(validator);
        // prank as validator
        (bool success, ) = address(token).call(abi.encodeWithSignature("emitToTreasury()"));
        assertTrue(success, "Validator should be able to emit");
    }

    function testEmitToTreasuryNotValidator() public {
        (bool success, ) = notValidator.call(abi.encodeWithSignature("emitToTreasury()"));
        assertFalse(success, "Not validator should not be able to emit");
    }

    function testAddRemoveQuantumValidator() public {
        token.addQuantumValidator(validator);
        assertTrue(token.quantumValidators(validator));
        assertEq(token.quantumValidatorCount(), 1);
        token.removeQuantumValidator(validator);
        assertFalse(token.quantumValidators(validator));
        assertEq(token.quantumValidatorCount(), 0);
    }

    function testSetEmissionRate() public {
        token.setEmissionRate(12345);
        assertEq(token.emissionRate(), 12345);
    }

    function testSetQuantumTreasury() public {
        token.setQuantumTreasury(address(0x1234));
        assertEq(token.quantumTreasury(), address(0x1234));
    }

    function testEmitToTreasuryMaxSupply() public {
        // Mint up to max supply
        uint256 toMint = token.MAX_SUPPLY() - token.totalSupply();
        token.setEmissionRate(toMint);
        token.emitToTreasury();
        assertEq(token.totalSupply(), token.MAX_SUPPLY());
        // Next emission should revert
        (bool success, ) = address(token).call(abi.encodeWithSignature("emitToTreasury()"));
        assertFalse(success, "Should not emit above max supply");
    }
} 