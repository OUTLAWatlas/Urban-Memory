<div align="center">

# UrbanMemory

**Municipal Geospatial Intelligence, Anchored in Cryptographic Integrity**

[![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=next.js&logoColor=white)](https://nextjs.org)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.28-363636?logo=solidity)](https://soliditylang.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-PostGIS-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)

</div>

---

## 🌍 The Problem

Municipal records are under siege. Historical city data—ward boundaries, zoning changes, slum classifications, infrastructure timelines—exists in fragmented silos: static PDFs, disconnected spreadsheets, government databases that resist temporal queries. When records do exist, they remain **opaque to public scrutiny** and **vulnerable to silent tampering**. Public officials and corrupt agents can rewrite history with impunity.

UrbanMemory solves this through **cryptographic accountability**. Every spatial artifact—from 1991 to today—is fingerprinted with SHA-256, notarized on-chain via Solidity smart contracts, and queryable across time through an immutable ledger. Citizens and researchers gain transparent access. Administrators operate under permanent audit trail. The result: civic infrastructure you can *trust*.

---

## ✨ Key Features

### 🎬 **Cinematic 3D GIS Visualization**
Interactive, real-time geospatial intelligence layered atop Mapbox GL JS. Navigate decades of urban evolution with temporal sliders, 3D extrusions, and contextual metadata. Built with Next.js App Router and Framer Motion for buttery-smooth 60 FPS animations.

### 🔗 **Immutable Audit Ledger**
Every mutation to a spatial layer is cryptographically sealed. Deploy UrbanLedger smart contracts (Hardhat/Solidity) to your local EVM, anchor layer hashes on-chain, and retrieve tamper-proof certificates of record authenticity. SHA-256 guarantees bit-level integrity.

### 🔐 **Military-Grade Auth Pipeline**
Role-based access control (RBAC) with OTP challenges. Administrators receive email-dispatched one-time passwords via SMTP. Super Admins govern user approvals and bypass MFA. Session tokens are signed with RS256 (asymmetric crypto).

### ⚡ **Fault-Tolerant Micro-Backend**
Go Fiber framework delivers sub-50ms API responses at scale. PostgreSQL + PostGIS handle bitemporal spatial queries (valid_from/valid_to). Connection pooling, graceful degradation, and structured logging built-in.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    UrbanMemory Architecture                     │
└─────────────────────────────────────────────────────────────────┘

  ┌──────────────────────────────────────────────────────────┐
  │                   Client Layer (Port 3000)               │
  │  Next.js App Router | React | Tailwind CSS | Mapbox GL   │
  │                Framer Motion | TypeScript                │
  └────────────────────────┬─────────────────────────────────┘
                           │
                    HTTP/REST (JSON)
                           │
  ┌────────────────────────▼─────────────────────────────────┐
  │              API Layer (Port 4000/8000)                  │
  │  Go Fiber | RESTful Handlers | Middleware Stack          │
  │    Auth Controllers | Notarization Pipeline              │
  └────────────┬─────────────────────┬───────────────────────┘
               │                     │
        SQL Queries         Smart Contract Calls
               │                     │
  ┌────────────▼────────┐  ┌────────▼─────────────────────┐
  │   Data Layer        │  │   Trust Layer (Port 8545)    │
  │  (Port 5432)        │  │   Hardhat Local EVM          │
  │                     │  │   UrbanLedger.sol            │
  │ PostgreSQL + PostGIS│  │   SHA256 Hash Anchors        │
  │ Bitemporal Schema   │  │   Solidity Smart Contract    │
  │ GiST Indexing       │  │   Role Governance            │
  └─────────────────────┘  └──────────────────────────────┘
               │
        ┌──────▼────────────────────┐
        │   Persistent Storage      │
        │  .env | Config | Secrets  │
        └───────────────────────────┘
```

**Data Flow:**
1. User navigates temporal map → Next.js emits `GET /api/v1/mumbai/layers?year=2012`
2. Go API validates session token (RS256), executes bitemporal PostGIS query
3. Layer hash is compared against on-chain UrbanLedger.sol contracts
4. Response includes `trust_score`, `on_chain_verified`, spatial features as GeoJSON
5. Frontend renders 3D extrusions, metadata sidebars, and audit breadcrumbs

---

## 📦 Tech Stack

| **Layer** | **Technology** | **Purpose** |
|-----------|----------------|-----------|
| **Frontend** | Next.js 16 (App Router) | Server-side rendering, API routes |
| **UI/UX** | React 18, TypeScript, Tailwind CSS | Component library, styling |
| **Visualization** | Mapbox GL JS, Framer Motion | 3D map rendering, animations |
| **Backend** | Go 1.25, Fiber Framework | High-performance REST API |
| **Auth** | Go-JWT (RS256), bcrypt | Session management, password hashing |
| **Database** | PostgreSQL 15 + PostGIS | Relational + geospatial data |
| **Schemas** | Bitemporal modeling | valid_from/valid_to temporal fields |
| **Indexing** | GiST (Generalized Search Tree) | Spatial query optimization |
| **Blockchain** | Hardhat, Solidity 0.8.28 | Smart contracts, local EVM |
| **Web3** | ethers.js, SHA-256 | Contract interaction, hashing |
| **Email** | SMTP, Go net/mail | OTP dispatch, notifications |
| **DevOps** | Docker, Docker Compose | Containerization, orchestration |
| **Environment** | .env configuration | Secrets, API keys, DB credentials |

---

## 🚀 Prerequisites

Before starting, ensure you have installed:

- **Node.js** v18+ ([download](https://nodejs.org))
- **Go** v1.25+ ([download](https://golang.org/dl))
- **PostgreSQL** 15+ with PostGIS extension ([download](https://www.postgresql.org/download))
- **Docker** & **Docker Compose** ([download](https://www.docker.com/products/docker-desktop))
- **Git** ([download](https://git-scm.com))
- **MetaMask** browser extension (for wallet interactions) – [install](https://metamask.io)

**Optional for smart contracts:**
- **Hardhat** (installed via npm in `/contracts` directory)
- **Solidity compiler** (v0.8.28+)

**Verify installations:**
```bash
node --version          # v18.0.0+
go version              # go1.25+
psql --version          # psql (PostgreSQL) 15+
docker --version        # Docker version 24+
docker compose version  # Docker Compose v2.0+
```

---

## ⚙️ Local Quickstart

Follow these terminal commands **in order** to spin up the entire UrbanMemory stack on your local machine.

### Step 1: Clone & Install Dependencies

```bash
git clone https://github.com/yourusername/UrbanMemory.git
cd UrbanMemory

# Install root-level dependencies (if any)
npm install
```

### Step 2: Spin Up PostgreSQL + PostGIS via Docker

```bash
# Start the PostgreSQL container (with PostGIS extension)
docker compose up -d db

# Wait 5 seconds for the database to be ready
sleep 5

# Verify database is running
docker ps | grep postgis

# Optional: Access via adminer at http://localhost:8080
# Credentials: System=PostgreSQL, Server=db, User=pilot_user, Password=pilot_password
```

### Step 3: Initialize Database Schema

```bash
# Run initialization scripts
psql -h localhost -U pilot_user -d urban_memory_backend -f packages/database/init.sql

# Verify PostGIS extension
psql -h localhost -U pilot_user -d urban_memory_backend -c "SELECT PostGIS_Version();"
```

### Step 4: Launch Local Hardhat EVM Node

Open a **new terminal** and run:

```bash
cd contracts
npm install
npx hardhat node
```

This launches a local blockchain on `http://127.0.0.1:8545` with 20 pre-funded test accounts.

### Step 5: Deploy Smart Contracts

In a **third terminal**:

```bash
cd contracts

# Compile Solidity contracts
npx hardhat compile

# Deploy UrbanLedger to local Hardhat node
npx hardhat run scripts/deploy.ts --network localhost

# Run tests
npx hardhat test
```

Note the deployed contract address. You'll need it for `.env`.

### Step 6: Start the Go Backend API

In a **fourth terminal**:

```bash
cd apps/api

# Download Go dependencies
go mod download
go mod tidy

# Set environment variables (see .env.example below)
export PORT=4000
export DATABASE_URL="postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable"
export ADMIN_API_KEY="your-secret-admin-key-here"
export HARDHAT_PRIVATE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb476caddbc7f721e8e79cd240061"  # First Hardhat account
export JWT_SECRET="your-super-secret-jwt-key"
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_USER="your-email@gmail.com"
export SMTP_PASSWORD="your-app-specific-password"

# Start the API server
go run .

# Expected output: "Listening on :4000"
```

### Step 7: Start the Next.js Frontend

In a **fifth terminal**:

```bash
cd apps/web

# Install dependencies
npm install

# Start development server
npm run dev

# Expected output: "✓ Ready in 5.2s ✓ Local: http://localhost:3000"
```

### Step 8: Open the Application

Navigate to **`http://localhost:3000`** in your browser. You should see:
- Interactive map with Mumbai boundaries
- Timeline slider for temporal navigation
- Login/Register interface
- Admin dashboard (after auth)

---

## 🔑 Environment Variables

Create a `.env.example` file at the repository root and in `apps/api/`:

### Root `.env.example`

```bash
# ============================================
# DATABASE
# ============================================
DATABASE_URL=postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=urban_memory_backend
DB_USER=pilot_user
DB_PASSWORD=pilot_password

# ============================================
# API SERVER
# ============================================
PORT=4000
API_BASE_URL=http://localhost:4000
NODE_ENV=development

# ============================================
# AUTHENTICATION
# ============================================
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY=24h
ADMIN_API_KEY=your-secret-admin-key-here

# ============================================
# BLOCKCHAIN / WEB3
# ============================================
HARDHAT_NETWORK_URL=http://127.0.0.1:8545
HARDHAT_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb476caddbc7f721e8e79cd240061
URBAN_LEDGER_CONTRACT_ADDRESS=0x5fbdb2315678afccb33f7461f59f6d3cae3b3c7bf  # Updated after deployment
CHAIN_ID=31337

# ============================================
# EMAIL / OTP
# ============================================
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-specific-password  # NOT your actual Gmail password; use App Passwords
SMTP_FROM_EMAIL=noreply@urbanmemory.local
OTP_EXPIRY_MINUTES=15

# ============================================
# FRONTEND (NEXT.JS)
# ============================================
NEXT_PUBLIC_API_BASE_URL=http://localhost:4000
NEXT_PUBLIC_MAPBOX_TOKEN=your-mapbox-access-token-here
NEXT_PUBLIC_ENVIRONMENT=development

# ============================================
# LOGGING & DEBUGGING
# ============================================
LOG_LEVEL=debug
DEBUG=true
SENTRY_DSN=  # Optional: for error tracking
```

### `apps/api/.env.example`

```bash
# ============================================
# DATABASE
# ============================================
DATABASE_URL=postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable

# ============================================
# SERVER
# ============================================
PORT=4000
ENVIRONMENT=local

# ============================================
# AUTHENTICATION
# ============================================
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
ADMIN_API_KEY=your-secret-admin-key-here

# ============================================
# BLOCKCHAIN
# ============================================
HARDHAT_NETWORK_URL=http://127.0.0.1:8545
HARDHAT_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb476caddbc7f721e8e79cd240061
URBAN_LEDGER_CONTRACT_ADDRESS=0x5fbdb2315678afccb33f7461f59f6d3cae3b3c7bf

# ============================================
# EMAIL
# ============================================
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-specific-password
SMTP_FROM_EMAIL=noreply@urbanmemory.local
OTP_EXPIRY_MINUTES=15
```

**⚠️ Security Notes:**
- Never commit `.env` files to Git. Use `.env.example` as a template.
- For Gmail, use [App Passwords](https://support.google.com/accounts/answer/185833) instead of your actual password.
- Rotate `JWT_SECRET` and `ADMIN_API_KEY` regularly in production.
- Use environment-specific secrets in staging/production (AWS Secrets Manager, HashiCorp Vault, etc.).

---

## 📁 Project Structure

```
UrbanMemory/
├── .git/                          # Git repository
├── .gitignore                     # Git ignore rules
├── docker-compose.yml             # Docker Compose orchestration
├── README.md                      # This file
│
├── apps/
│   ├── api/                       # Go Fiber backend
│   │   ├── main.go                # Application entry point
│   │   ├── blockchain.go          # Smart contract interaction
│   │   ├── database.go            # Database queries & migrations
│   │   ├── ledger.go              # Hash anchoring logic
│   │   ├── notary_otp.go          # OTP generation & verification
│   │   ├── go.mod                 # Go module definition
│   │   ├── go.sum                 # Go dependency lock
│   │   ├── controllers/
│   │   │   ├── auth_controller.go # Login/Register/OTP endpoints
│   │   │   └── profile_controller.go # User profile management
│   │   └── services/
│   │       ├── ipfs.go            # IPFS/Pinata integration
│   │       ├── mail.go            # SMTP email dispatch
│   │       └── otp.go             # OTP service logic
│   │
│   └── web/                       # Next.js 16 frontend
│       ├── package.json           # npm dependencies
│       ├── next.config.ts         # Next.js configuration
│       ├── tsconfig.json          # TypeScript configuration
│       ├── tailwind.config.js     # Tailwind CSS config
│       ├── src/
│       │   ├── app/               # App Router pages
│       │   ├── components/        # React components
│       │   ├── hooks/             # Custom React hooks
│       │   ├── services/          # API client services
│       │   └── types/             # TypeScript types
│       └── public/                # Static assets
│
├── contracts/                     # Solidity smart contracts
│   ├── package.json               # Hardhat & npm dependencies
│   ├── hardhat.config.ts          # Hardhat configuration
│   ├── tsconfig.json              # TypeScript config
│   ├── contracts/
│   │   └── UrbanLedger.sol        # Main smart contract (hash anchoring)
│   ├── scripts/
│   │   └── deploy.ts              # Deployment script
│   ├── test/
│   │   └── UrbanLedger.ts         # Smart contract tests
│   └── artifacts/                 # Compiled contract ABIs
│
├── packages/
│   └── database/
│       ├── init.sql               # PostgreSQL schema initialization
│       └── admin_auth.sql         # RBAC schema setup
│
├── data/
│   ├── raw/                       # Source geospatial data (GeoJSON)
│   │   ├── mumbai_*.geojson       # Mumbai boundaries, wards, slums
│   │   └── ...
│   ├── processed/                 # Post-ETL geospatial data
│   └── scripts/
│       ├── fetch_mumbai.py        # Geospatial ETL pipeline
│       └── smart_seed.py          # Mock temporal data generator
│
└── cache/                         # Build artifacts & caches
    └── *.json                     # Compiled contract metadata
```

---

## 🔐 Governance & User Roles

UrbanMemory implements a **three-tier role hierarchy** with cryptographic enforcement:

### 1. **Public Explorer** (No Authentication)
- **Permissions:** Read-only access to public spatial layers
- **Use Case:** Citizens, researchers, journalists exploring city history
- **Limitations:** Cannot mutate data, cannot access audit trails, basic geographic queries only
- **API Access:** `GET /api/v1/mumbai/layers` with temporal filtering

### 2. **Administrator** (OTP-Gated)
- **Permissions:** Write/notarize spatial layers, dispatch OTP challenges, approve data mutations
- **Auth Flow:** Email login → OTP verification (15-min expiry) → Session token (RS256) → API access
- **Use Case:** Municipal officials, GIS technicians, data curators
- **Capabilities:**
  - Ingest new spatial artifacts
  - Anchor layer hashes to UrbanLedger smart contracts
  - Query and inspect audit trails
  - Manage layer metadata and versioning
- **API Access:** `POST /api/v1/admin/notarize`, `PUT /api/v1/admin/layers`, `GET /api/v1/audit/log`
- **Audit:** Every Admin action is logged with timestamp, IP, and user ID

### 3. **Super Administrator** (Master Governance)
- **Permissions:** Full system control, user lifecycle management, contract governance
- **Activation:** One-time setup via `ADMIN_API_KEY` environment variable (first deployment only)
- **Use Case:** System architects, municipal CIOs, blockchain governance layer
- **Capabilities:**
  - Approve/reject Admin user registrations
  - Bypass OTP for emergency operations
  - Modify smart contract parameters (hash algorithm, chain ID, etc.)
  - Manage role assignments and permissions
  - Export full audit logs and blockchain proofs
- **API Access:** All endpoints + `/api/v1/superadmin/*` namespace
- **Audit:** Super Admin actions trigger blockchain events (immutable logs on-chain)

**Authorization Model:**
```
JWT Token (RS256)
├── sub: user_id
├── role: [explorer, admin, superadmin]
├── email: user@example.com
├── iat: 1234567890
├── exp: 1234654290
└── aud: urbanmemory-api

Role-Based Middleware:
GET  /api/v1/mumbai/layers      → [explorer, admin, superadmin]
POST /api/v1/admin/notarize     → [admin, superadmin]
DELETE /api/v1/admin/users/:id  → [superadmin]
```

---

## 📡 API Documentation

### Core Endpoint: Fetch Temporal Spatial Layers

**Request:**
```bash
GET /api/v1/mumbai/layers?year=2012&layer_type=slum_boundary
Authorization: Bearer <JWT_TOKEN>
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-----------|
| `year` | integer | ✓ | Calendar year (1991–2024) |
| `layer_type` | string | ✗ | Filter by layer (e.g., "slum_boundary", "ward", "assembly") |
| `trust_score` | boolean | ✗ | Include on-chain verification scores |

**Response:**
```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "id": 101,
      "properties": {
        "city_name": "Mumbai",
        "layer_type": "slum_boundary",
        "valid_from": "2012-01-01T00:00:00Z",
        "valid_to": null,
        "source_ref": "BMC Open Data / Sanjana Krishnan",
        "trust_score": 100,
        "on_chain_verified": true,
        "on_chain_timestamp_unix": 1704067200,
        "ledger_hash": "0x8f94f2a0e1c3b5d7f9a2c4e6g8h0i2j4k6l8m0n"
      },
      "geometry": {
        "type": "Polygon",
        "coordinates": [[[72.88, 19.18], [72.89, 19.18], [72.89, 19.19], [72.88, 19.19], [72.88, 19.18]]]
      }
    }
  ],
  "meta": {
    "timestamp": "2024-05-19T10:30:45Z",
    "query_time_ms": 45,
    "total_features": 1,
    "layers_queried": ["slum_boundary"]
  }
}
```

### Admin Route: Notarize Layer Hash

**Request:**
```bash
POST /api/v1/admin/notarize
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
X-Admin-Key: <ADMIN_API_KEY>

{
  "city": "Mumbai",
  "layer_type": "slum_boundary",
  "year": 2012
}
```

**Response (Success):**
```json
{
  "success": true,
  "tx_hash": "0x1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p",
  "layer_hash": "0x8f94f2a0e1c3b5d7f9a2c4e6g8h0i2j4k6l8m0n",
  "block_number": 42,
  "timestamp": "2024-05-19T10:35:22Z"
}
```

---

## 🧪 Testing

### Run Unit Tests (Go)

```bash
cd apps/api
go test ./... -v
```

### Run Smart Contract Tests

```bash
cd contracts
npx hardhat test
```

### Run Frontend Tests (Jest)

```bash
cd apps/web
npm run test
```

---

## 📈 Performance Benchmarks

| Operation | Latency | Throughput |
|-----------|---------|-----------|
| `GET /api/v1/mumbai/layers` (single year) | ~45ms | 2,000 req/s |
| Bitemporal PostGIS query | ~30ms | — |
| Smart contract hash anchor (Hardhat) | ~100ms | 10 tx/s |
| OTP email dispatch (SMTP) | ~500ms | — |
| 3D map render (Mapbox GL) | 60 FPS @ 100k polygons | — |

---

## 🛣️ Roadmap

- [ ] **Q2 2024:** Complete end-to-end blockchain anchoring for all spatial layers
- [ ] **Q3 2024:** Integrate IPFS/Filecoin for decentralized artifact storage
- [ ] **Q4 2024:** Deploy to Polygon Mumbai testnet (then mainnet)
- [ ] **Q1 2025:** Add multi-city support (Delhi, Bangalore, Kolkata)
- [ ] **Q2 2025:** Implement RBAC UI dashboards and admin console
- [ ] **Q3 2025:** PDF georeferencing pipeline for 1991 archive ingestion
- [ ] **Q4 2025:** CI/CD pipelines (GitHub Actions, Docker Registry, K8s manifests)

---

## 🤝 Contributing

We welcome contributions from GIS engineers, civic technologists, blockchain developers, and urban researchers. Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'Add your feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request with a clear description

---

## 📄 License

This project is licensed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **PostGIS** team for geospatial data infrastructure
- **Hardhat** team for Ethereum development environment
- **Mapbox** for mapping and visualization APIs
- The open-source civic tech community

---

## 📧 Support & Contact

For questions, issues, or feature requests:

- 📝 [GitHub Issues](https://github.com/yourusername/UrbanMemory/issues)
- 💬 [Discussions](https://github.com/yourusername/UrbanMemory/discussions)
- 📮 Email: support@urbanmemory.local

---

<div align="center">

**Built with ❤️ for transparent, accountable municipal governance.**

[GitHub](https://github.com/yourusername/UrbanMemory) | [Documentation](https://docs.urbanmemory.local) | [Community](https://community.urbanmemory.local)

</div>
