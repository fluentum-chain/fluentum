const { ethers } = require("hardhat");

module.exports = async ({ getNamedAccounts, deployments }) => {
  const { deploy } = deployments;
  const { deployer } = await getNamedAccounts();

  // Deploy FLUXToken
  const fluxToken = await deploy("FLUXToken", {
    from: deployer,
    args: [
      "0x1234567890abcdef1234567890abcdef12345678", // Initial treasury address
      "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd", // Team vesting contract
      "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", // Airdrop distributor
    ],
    log: true,
  });

  // Deploy FluxGovernance
  const fluxGovernance = await deploy("FluxGovernance", {
    from: deployer,
    args: [fluxToken.address],
    log: true,
  });

  // Deploy EmissionScheduler
  const emissionScheduler = await deploy("EmissionScheduler", {
    from: deployer,
    args: [], // Add constructor args if needed
    log: true,
  });

  // Deploy QuantumValidatorRegistry
  const quantumValidatorRegistry = await deploy("QuantumValidatorRegistry", {
    from: deployer,
    args: [], // Add constructor args if needed
    log: true,
  });

  // Log deployed addresses
  console.log("FLUXToken deployed to:", fluxToken.address);
  console.log("FluxGovernance deployed to:", fluxGovernance.address);
  console.log("EmissionScheduler deployed to:", emissionScheduler.address);
  console.log("QuantumValidatorRegistry deployed to:", quantumValidatorRegistry.address);
}; 