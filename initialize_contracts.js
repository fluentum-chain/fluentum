const { ethers } = require("hardhat");

async function main() {
  const flux = await ethers.getContract("FLUXToken");
  const governance = await ethers.getContract("FluxGovernance");

  // Configure relationships
  console.log("Setting governance contract on FLUXToken...");
  await (await flux.setGovernanceContract(governance.address)).wait();
  console.log("Setting token contract on FluxGovernance...");
  await (await governance.setTokenContract(flux.address)).wait();

  // Initialize emission schedule
  console.log("Initializing emission schedule on FLUXToken...");
  await (await flux.initializeEmission(25000, 0)).wait(); // 25k FLUX/block, start block 0

  console.log("Post-deployment setup complete.");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  }); 