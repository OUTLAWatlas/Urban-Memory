# Hardhat Smart Contracts Setup

## Status
✅ **Hardhat v3.4.1 initialized**  
✅ **Dependencies installed**  
❌ **Node.js version mismatch** — compilation blocked

## The Issue
Hardhat 3.x requires **Node.js 22.10.0 LTS or later** (even major version number).  
Your current environment: **Node.js 20.18.0** (unsupported)

Error on `npm run compile`:
```
TypeError: this[#dependenciesMap].values(...).flatMap is not a function
```

## Solution: Upgrade Node.js

### Option 1: Direct Upgrade (Recommended)
Download Node.js 22 LTS from: https://nodejs.org/  
Install and verify:
```bash
node --version  # Should show v22.x.x or later
npm --version
```

### Option 2: Using NVM (Node Version Manager)
```bash
# On macOS/Linux
nvm install 22
nvm use 22

# On Windows, use nvm-windows
# Download: https://github.com/coreybutler/nvm-windows/releases
```

### Option 3: Using .nvmrc (for team consistency)
We've created `.nvmrc` in this folder. After upgrading, run:
```bash
nvm use  # Automatically switch to Node 22
```

## After Upgrading Node.js

Clear node_modules and reinstall:
```bash
rm -r node_modules package-lock.json  # On Windows: rmdir /s node_modules
npm install
```

Then compile:
```bash
npm run compile
```

## Contract Files
- `contracts/UrbanLedger.sol` — Your custom smart contract (stub)
- `contracts/Lock.sol` — Hardhat template example
- `ignition/` — Deployment scripts
- `test/` — Mocha test suite
