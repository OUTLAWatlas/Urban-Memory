# Urban Memory

> A full-stack Web3 spatial engine for Mumbai that preserves historical urban geography as a cryptographically verifiable public record.

![Urban Memory banner placeholder](https://placehold.co/1600x480/png?text=Urban+Memory)

[![Next.js](https://img.shields.io/badge/Next.js-16.1.6-000000?logo=nextdotjs&logoColor=white)](https://nextjs.org)
[![React](https://img.shields.io/badge/React-19.2.3-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6?logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%20%2B%20PostGIS-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.28-363636?logo=solidity)](https://soliditylang.org)
[![Hardhat](https://img.shields.io/badge/Hardhat-3.x-yellow?logo=ethereum&logoColor=black)](https://hardhat.org)
[![IPFS](https://img.shields.io/badge/IPFS%2FPinata-Decentralized%20Storage-65C2CB)](https://www.pinata.cloud)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4.x-38B2AC?logo=tailwindcss&logoColor=white)](https://tailwindcss.com)

**Urban Memory** is an academic-grade geospatial integrity system that reconstructs Mumbai across the epochs 1991, 2000, 2012, and 2024. It solves the core governance problem of municipal data mutation by storing heavy spatial payloads in PostGIS and IPFS while anchoring SHA-256 hashes immutably on an Ethereum ledger.

---

## Project Abstract & Thesis Statement

Traditional GIS systems are optimized for storage and query performance, not for adversarial integrity. Municipal boundary files, zoning overlays, slum extents, and civic records are often maintained in mutable databases, shared file systems, or exported datasets that can be revised without durable proof of prior state. In such an environment, historical geography can be tampered with, selectively redacted, or silently replaced, and the absence of cryptographic traceability makes those actions difficult to detect after the fact.

Urban Memory proposes a decentralized evidentiary model for urban spatial history. The platform treats each GeoJSON layer as a verifiable artifact: the raw geometry is queried from PostGIS, normalized and hashed with SHA-256, pinned to IPFS through Pinata, and recorded on-chain in `UrbanLedger.sol`. The result is a Hybrid Pointer Architecture in which the database remains the operational store, IPFS provides content-addressed distribution, and the Ethereum ledger provides immutable proof.

> The thesis is simple: if a city's past can be rewritten without evidence, then the present is already untrustworthy. Urban Memory converts historical GIS from a mutable data product into a cryptographically auditable record.

---

## System Architecture: The Hybrid Pointer Model

The platform uses a split-storage, proof-on-chain design.

- The frontend requests a historical layer for a given city, year, and layer type.
- The Go API queries PostGIS and reconstructs the GeoJSON payload.
- The backend computes a SHA-256 digest of the live payload.
- The payload is pinned to IPFS, producing a CID.
- The digest and CID are compared against the Smart Contract record.
- The UI renders a trust state, enabling immediate detection of tamper, drift, or missing notarization.

Heavy GeoJSON is intentionally not stored directly on-chain because blockchains are optimized for deterministic state transitions, not large binary payloads. Storing polygons, coordinate rings, and multi-epoch feature collections on Ethereum would be prohibitively expensive, inefficient to query, and structurally mismatched to the storage model. Urban Memory therefore stores the data off-chain and stores only the proof of the data on-chain.

### Hash-Matching Loop

```text
PostGIS GeoJSON -> canonical serialization -> SHA-256 -> IPFS CID -> Solidity record -> runtime verification
```

This loop ensures that the frontend can independently verify the integrity of each loaded epoch. If even one byte of the payload changes, the hash changes, and the on-chain comparison fails.

### Mermaid Flowchart

```mermaid
flowchart LR
    U[User UI\nNext.js + Mapbox GL JS] <--> API[Go API\nFiber v2]
    API <--> PG[(PostGIS Database)]
    API <--> IPFS[Pinata / IPFS]
    API <--> EVM[Hardhat EVM\nUrbanLedger.sol]
    API -->|GeoJSON + verification metadata| U
    U -->|timeline selection| API
    API -->|hash + CID seal| IPFS
    API -->|commit / verify| EVM
```

---

## Repository Structure

```text
UrbanMemory/
|-- README.md
|-- docker-compose.yml
|-- apps/
|   |-- web/
|   |   |-- src/
|   |   |   `-- app/
|   |   |       |-- api/
|   |   |       `-- components/
|   |   |-- package.json
|   |   `-- public/
|   `-- api/
|       |-- main.go
|       |-- database.go
|       |-- blockchain.go
|       |-- ledger.go
|       |-- controllers/
|       |-- services/
|       `-- go.mod
|-- contracts/
|   |-- UrbanLedger.sol
|   |-- hardhat.config.ts
|   |-- scripts/
|   `-- deployed-address.json
|-- data/
|   |-- raw/
|   `-- scripts/
|       `-- story_seed.py
`-- packages/
    `-- database/
        |-- init.sql
        `-- admin_auth.sql
```

### What Each Area Owns

- `/apps/web/` contains the Next.js App Router frontend, Map components, and browser-side API proxy routes.
- `/apps/api/` contains the Fiber backend, auth controllers, spatial PostGIS queries, and Hardhat ABI bindings for ledger verification.
- `/contracts/` contains the Ethereum workspace, the `UrbanLedger.sol` contract, and deployment scripts.
- `/data/scripts/` contains the Python ETL pipeline that injects historical geometries into the local database.
- `/packages/database/` contains SQL schema initialization for spatial artifacts and admin governance tables.

---

## Core Engineering Workflows

### Temporal Delta Engine

The Temporal Delta Engine is the mechanism that makes Mumbai feel like a live historical simulation rather than a static map.

In the frontend, `Map.tsx` listens to timeline changes and issues a fetch to the Go backend for the selected epoch. The API then returns the exact set of features valid for that year, and the UI repaints the map layers rather than interpolating from cached state. This matters because the project is not visualizing abstract time metadata; it is re-rendering the city's spatial geometry epoch by epoch.

### Zero-Trust Ledger Verification

Each payload load is verified at runtime.

The Go backend hashes the live GeoJSON payload with SHA-256, resolves the contract address, and checks the value against the on-chain record in `UrbanLedger.sol`. The UI receives the result and uses it to drive the IntegrityBadge: green for cryptographic agreement, red for mismatch or missing notarization.

This verification step is zero-trust because the database is not assumed to be honest. The system proves integrity every time the payload is requested.

### Administrative Governance

Administrative mutation is gated behind a super-admin workflow.

- A `super_admin` JWT bearer token authorizes privileged routes.
- The backend can initiate a notarization or decentralized seal flow.
- The payload is hashed, pinned to IPFS, and committed to the ledger.
- The resulting chain of custody is auditable and reproducible.

This governance model is intentionally restrictive. It is designed to make historical mutation explicit, reviewable, and cryptographically bounded rather than silently editable.

---

## The Data Pipeline (ETL)

The ETL pipeline lives in `data/scripts/story_seed.py` and is responsible for transforming the historical boundary dataset into a strict spatial seed for PostGIS.

The script uses GeoPandas and Shapely to generate and buffer historical zones such as SRA slums, forests, mills, roads, and redevelopment footprints. Each feature is tagged with `city_name`, `layer_type`, `valid_from`, `valid_to`, and `source_ref`, then written into the `urban_artifacts` table using EPSG:4326 alignment.

Key characteristics of the pipeline:

- Raw `.geojson` assets are used as source material for historical context.
- Geometries are normalized to geographic coordinates in EPSG:4326.
- Temporal validity is encoded with `valid_from` and `valid_to` fields.
- The script injects strict historical zones such as SGNP, Aarey, Dharavi, mills, BKC, and transport corridors.
- The data is appended into PostGIS for spatial querying and later hashed for ledger verification.

The effect is an auditable seed layer for Mumbai's temporal map history rather than a generic spatial dump.

---

## Local Setup & Deployment Guide

### Prerequisites

Install the following before starting:

- Docker and Docker Compose
- Node.js 22+ for the frontend and contracts workspace
- Go 1.25+ for the API
- PostgreSQL client tools if you want to inspect the database manually

### Step 1: Database

Start the PostGIS database and supporting services with Docker Compose.

```bash
docker compose up -d db
```

This launches PostgreSQL with the PostGIS extension and loads `packages/database/init.sql` automatically.

### Step 2: Smart Contracts

From the `contracts/` directory, install dependencies and bring up the local Ethereum EVM.

```bash
cd contracts
npm install
npx hardhat node
```

In a separate terminal, compile and deploy the `UrbanLedger` contract:

```bash
cd contracts
npx hardhat compile
npx hardhat ignition deploy --network hardhat <ignition-module>
```

If your workspace uses the repository's current deployment script instead of Ignition for local testing, the equivalent deployment path is the `scripts/deploy.js` flow exposed through `npm run deploy`.

### Step 3: Backend

Start the Go backend from the API package.

```bash
cd apps/api
go run main.go
```

If you prefer the package-root form used by the Go toolchain in this repository, `go run .` is also valid. The service loads environment variables, connects to PostGIS, configures admin bootstrap state, and exposes the API on port `4000` by default.

### Step 4: Frontend

Start the Next.js frontend.

```bash
cd apps/web
npm install
npm run dev
```

The web app runs on `http://localhost:3000` and proxies API requests through its App Router route handlers.

---

## Environment Variables

### `apps/api/.env`

| Variable | Required | Purpose | Example |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | PostgreSQL/PostGIS connection string | `postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable` |
| `PORT` | No | API listen port | `4000` |
| `DEFAULT_CITY` | No | Default city for layer queries | `Mumbai` |
| `LEDGER_CONTRACT_ADDRESS` | No | Deployed UrbanLedger contract address | `0x5fbdb2315678afccb33f7461f59f6d3cae3b3c7bf` |
| `ETH_NODE_URL` | Yes for ledger verification | Ethereum RPC endpoint for the Hardhat node or testnet | `http://127.0.0.1:8545` |
| `ETH_PRIVATE_KEY` | Yes for sealing | Signer private key used for on-chain commits | `0xac0974bec39a17e36ba4a6b4d238ff944bacb476caddbc7f721e8e79cd240061` |
| `LEDGER_SOURCE_REF` | No | Human-readable provenance stored on-chain | `Urban Memory API` |
| `BLOCKCHAIN_ENABLED` | No | Enables or disables verification and commit flow | `true` |
| `ADMIN_API_KEY` | Yes for bootstrap/admin flows | Administrative secret for privileged operations | `change-me-before-production` |
| `ADMIN_SESSION_SECRET` | Yes for session signing | Secret for admin session material | `change-me-before-production` |
| `SMTP_EMAIL` | Yes for OTP and alerts | Sender email address | `noreply@urbanmemory.local` |
| `SMTP_HOST` | Yes for OTP and alerts | SMTP relay host | `smtp.gmail.com` |
| `SMTP_PORT` | Yes for OTP and alerts | SMTP relay port | `587` |
| `SMTP_APP_PASSWORD` | Yes for OTP and alerts | App password for SMTP | `your-app-password` |
| `SMTP_DISABLE_AUTH` | No | Disables SMTP auth for local test mailers | `false` |
| `PINATA_API_KEY` | Yes for IPFS sealing | Pinata API key | `pinata_api_key` |
| `PINATA_JWT` | Yes for IPFS sealing | Pinata JWT used for authenticated uploads | `pinata_jwt_token` |
| `PINATA_ENDPOINT` | No | Pinata upload endpoint override | `https://api.pinata.cloud/pinning/pinFileToIPFS` |

### `apps/web/.env.local`

| Variable | Required | Purpose | Example |
| --- | --- | --- | --- |
| `NEXT_PUBLIC_API_URL` | Yes | Base URL for the Go API and browser proxy routes | `http://localhost:4000` |
| `NEXT_PUBLIC_MAPBOX_ACCESS_TOKEN` | Yes for map rendering | Mapbox access token for the 3D map interface | `pk.eyJ1IjoieW91ci1rZXkiLCJhIjoiY2x...` |

> Keep the frontend URL explicit so the App Router proxy and the map client resolve against the same backend endpoint.

---

## API Reference

### `GET /api/v1/:city/layers`

Primary endpoint for historical spatial retrieval.

**Purpose:** returns a GeoJSON payload for a city, layer type, and year, along with verification metadata.

**Typical query parameters:**

- `layer_type` or `trust_layer_type`
- `year`

**Example:**

```bash
GET /api/v1/Mumbai/layers?layer_type=slum_boundary&year=2012
```

This endpoint powers the temporal map, the layer comparison logic, and the verification badge.

### `POST /api/v1/admin/seal-decentralized`

Primary notarization endpoint for administrative sealing.

**Purpose:** commits a selected payload through the decentralized sealing workflow, which includes hashing, IPFS upload, and on-chain ledger anchoring.

**Access:** protected by the admin governance layer and intended for `super_admin` use.

**Outcome:** produces a cryptographically auditable record that ties the database snapshot, IPFS CID, and Smart Contract hash together.

---

## Why This Architecture Exists

Urban Memory is not a visualization demo with blockchain attached as decoration. It is a systems argument: if public urban data can be altered without evidence, then public memory itself is vulnerable.

By splitting the system into an operational store, a decentralized content layer, and an immutable ledger, the project demonstrates how modern GIS infrastructure can be made both queryable and defensible. That is the thesis, the product, and the design constraint all at once.
