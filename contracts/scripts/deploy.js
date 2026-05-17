import { network } from "hardhat";
import fs from "fs";
import path from "path";

function writeDeploymentAddress(address) {
  const deploymentFile = path.join(process.cwd(), "deployed-address.json");
  const deploymentData = {
    address,
    timestamp: new Date().toISOString(),
    network: "localhost",
  };
  fs.writeFileSync(deploymentFile, JSON.stringify(deploymentData, null, 2));
  console.log(`📝 Deployment address saved to: ${deploymentFile}`);
}

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
  writeDeploymentAddress(address);
}

main().catch((error) => {
  console.error("❌ Deployment failed:", error);
  process.exitCode = 1;
});