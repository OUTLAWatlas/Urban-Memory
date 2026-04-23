# Smart Contract Deployment Guide

## Prerequisites
- **Node.js 22.10.0+** LTS (Hardhat 3.x requirement)
- Hardhat initialized and dependencies installed ✅

## Contracts Setup (Post-Node Upgrade)

### 1. Compile the Smart Contracts
```bash
cd contracts
npm run compile
```

This compiles:
- `UrbanLedger.sol` — The main ledger (hash anchoring for GIS data)
- `Counter.sol` — Template example (can be removed)

### 2. Deploy to Local Hardhat Network
```bash
npm run deploy
```

**Expected Output:**
```
🚀 Initializing deployment...
✅ UrbanLedger successfully deployed to address: 0x5FbDB2315678afccb333f8432F3aC7dDf1d2E7Ae
🔒 Save this address! Your Go backend will need it.
```

### 3. Deploy to Polygon Testnet (Sepolia)
Edit `hardhat.config.ts` and set environment variables:
```bash
export SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
export SEPOLIA_PRIVATE_KEY=your_private_key
npm run deploy -- --network sepolia
```

## Integration with Go Backend

Once deployed, configure your Go API to:
1. Read the contract address from deploy output
2. Use ethers.js or similar to interact with `UrbanLedger` contract
3. Call `commitLayerHash()` when ingesting GIS data
4. Call `verifyLayer()` to validate historical records

## Script Files
- `scripts/deploy.js` — Deployment orchestration
- `hardhat.config.ts` — Network and compiler configuration
- `test/` — Test suite (run with `npm run test`)

## Troubleshooting

**"Node.js not supported"**
→ Upgrade to Node.js 22.10.0 LTS: https://nodejs.org/

**"No Hardhat config found"**
→ Run from the `contracts/` directory

**"Contract not found"**
→ Ensure `UrbanLedger.sol` is in `contracts/contracts/` folder
