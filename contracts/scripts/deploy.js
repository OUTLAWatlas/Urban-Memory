import { network } from "hardhat";

async function main() {
  console.log("🚀 Initializing deployment...");

  const { ethers } = await network.create();

  // Deploy the compiled smart contract
  const ledger = await ethers.deployContract("UrbanLedger");

  // Wait for the transaction to be mined
  await ledger.waitForDeployment();

  const address = await ledger.getAddress();
  console.log(`✅ UrbanLedger successfully deployed to address: ${address}`);
  console.log("🔒 Save this address! Your Go backend will need it.");
}

main().catch((error) => {
  console.error("❌ Deployment failed:", error);
  process.exitCode = 1;
});