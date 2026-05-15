# UrbanMemory 🌆🧠

**A Temporal Urban Operating System utilizing Hybrid Blockchain-GIS Architectures for Decentralized City Archiving.**

![License](https://img.shields.io/badge/License-Not%20Specified-lightgrey)
![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)
![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=next.js)

---

## Abstract

UrbanMemory is a full-stack geospatial intelligence platform designed to combat **Administrative Amnesia**: the persistent loss, fragmentation, and manipulability of historical city data across wards, slum boundaries, zoning records, and civic infrastructure timelines. In many municipalities, these records remain trapped in static PDFs, disconnected spreadsheets, and siloed systems that cannot be queried temporally or verified cryptographically.

This project rethinks city memory as a **bitemporal, queryable, and verifiable system**. UrbanMemory delivers an interactive 4D map experience where users can move across time (1991-2024), inspect historical layers, and understand urban evolution through spatial deltas rather than flat snapshots.

At the trust layer, UrbanMemory introduces a **hybrid blockchain-GIS architecture**: spatial artifacts are fingerprinted with SHA-256, anchored via Polygon PoS smart contracts, and optionally persisted in decentralized storage (IPFS/Pinata). The result is a chain-of-custody model for civic truth, enabling resilient public archives that are resistant to silent tampering.

---

## System Architecture 🏗️

### Service Communication Flow

```text
[Next.js Web App :3000]
        |
        | HTTP (Temporal + Spatial Requests)
        v
[Go Fiber API :8000 (target) / :4000 (current local default)]
        |
        | SQL + PostGIS Operators
        v
[PostgreSQL + PostGIS :5432]

Parallel Trust Plane:
[API/Data Pipeline] --> [IPFS via Pinata] --> [Polygon PoS Smart Contract Hash Anchors]
```

### Runtime Topology

- **Frontend** renders interactive geospatial layers with MapLibre GL (WebGL).
- **Backend API** executes bitemporal spatial queries and returns lightweight GeoJSON payloads.
- **Database** stores geometry with temporal fields (`valid_from`, `valid_to`) and GiST spatial indexing.
- **Data Pipeline** normalizes raw GeoJSON and ingests city artifacts into PostGIS.
- **Web3 Layer** secures provenance through immutable hash commitments.

---

## Key Features ✨

- **Temporal Delta Engine**
  - Timeline slider dynamically requests year-scoped spatial states and visualizes urban change over time.
- **Deep Dive Inspector**
  - Interactive polygon inspection with temporal metadata, source references, and simulated density context.
- **Bitemporal Spatial Queries**
  - Database-side filtering dramatically reduces browser-side GeoJSON processing overhead.
- **Decentralized Chain of Custody**
  - Hash-anchored records enforce tamper-evident historical audit trails.
- **High-Concurrency API Layer**
  - Fiber-based Go API architecture targeting low-latency response patterns (~50ms class workloads).

---

## Tech Stack 🧰

### Frontend
- Next.js
- React
- TypeScript
- Tailwind CSS
- MapLibre GL / react-map-gl

### Backend
- Go
- Fiber
- RESTful API architecture

### Database & Geospatial
- PostgreSQL
- PostGIS
- GiST spatial indexing
- Bitemporal modeling with `valid_from`, `valid_to`

### Data Engineering
- Python
- GeoPandas
- SQLAlchemy
- GeoJSON ETL workflows

### Blockchain / Web3 (Hybrid Layer)
- IPFS (Pinata)
- Polygon PoS
- SHA-256 content verification hashes

### DevOps & Operations
- Docker
- Docker Compose
- Containerized service orchestration

---

## Getting Started (Local Development) 🚀

### Prerequisites

Install the following before bootstrapping:

- Docker + Docker Compose
- Node.js 20+
- Go 1.25+
- Python 3.10+

**Smart Contracts (optional):**
- Node.js 22.10.0+ LTS (Hardhat requirement; separate environment)
- See [contracts/SETUP.md](contracts/SETUP.md) for Web3 development guidance.

### 1) Clone and Enter the Repository

```bash
git clone <your-repo-url>
cd UrbanMemory
```

### 2) Start Core Infrastructure

The current Compose file boots PostGIS and Adminer.

```bash
# Preferred modern command
docker compose up -d

# Legacy equivalent
docker-compose up -d
```

Exposed services:

- `db` -> `localhost:5432`
- `adminer` -> `localhost:8080`

### 3) Run the API Service

```powershell
cd apps/api
go mod tidy
$env:PORT="4000"
$env:DATABASE_URL="postgres://pilot_user:pilot_password@localhost:5432/urban_memory_backend?sslmode=disable"
go run .
```

If you are using Command Prompt (cmd.exe), use `set PORT=4000` and `set DATABASE_URL=...` instead.

On Windows, if you see a cgo compiler error like "cc1.exe: sorry, unimplemented: 64-bit mode not compiled in", use:

```powershell
$env:CGO_ENABLED="0"
go run .
```

If you are using Command Prompt (cmd.exe), use `set CGO_ENABLED=0` instead.

> Note: Architecture target is **:8000**. Current web rewrite config points to **:4000** in local development.

### 4) Run the Web Client

```bash
cd apps/web
npm install
npm run dev
```

Open: `http://localhost:3000`

### 5) Generate Mock Temporal Data (`smart_seed.py`)

```bash
cd data/scripts
python smart_seed.py
```

This generates a synthetic GeoJSON file for local experiments.

### 6) Ingest Spatial Data into PostGIS

```bash
# from repository root
pip install geopandas sqlalchemy psycopg2-binary
python data/scripts/fetch_mumbai.py
```

### 7) (Optional) Deploy Smart Contracts

If working on blockchain integration (UrbanLedger, hash anchoring):

```bash
cd contracts
npm install
npm run compile
npm run test
```

Requires **Node.js 22.10.0+** LTS. See [contracts/SETUP.md](contracts/SETUP.md) for details.

---

## API Documentation 📘

### Main Endpoint

`GET /api/v1/mumbai/layers?year=YYYY`

Returns all spatial artifacts valid for the supplied calendar year using bitemporal constraints.

Optional trust query parameters:

- `trust_layer_type` – layer to verify against the on-chain hash for that year.
- `layer_type` – filter returned features to a specific layer type.

Trust fields returned in the same payload:

- `trust_score` (`0` or `100`)
- `on_chain_verified` (`true` if DB hash matches ledger hash)
- `verified_layer_type`, `on_chain_timestamp_unix`, `on_chain_source`

### Example Request

```bash
curl "http://localhost:4000/api/v1/mumbai/layers?year=2012&trust_layer_type=slum_boundary"
```

### Admin Notarize Route

`POST /api/v1/admin/notarize`

Protected route to commit the current DB hash for a layer/year to the UrbanLedger contract.

```bash
curl -X POST "http://localhost:4000/api/v1/admin/notarize" \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: <ADMIN_API_KEY>" \
  -d '{"city":"Mumbai","layer_type":"slum_boundary","year":2012}'
```

Set `ADMIN_API_KEY` in the API environment before starting the server.

### Example Response (truncated)

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "properties": {
        "id": 101,
        "city_name": "Mumbai",
        "layer_type": "slum_boundary",
        "valid_from": "2012-01-01T00:00:00Z",
        "valid_to": null,
        "source_ref": "Sanjana Krishnan / BMC Open Data"
      },
      "geometry": {
        "type": "Polygon",
        "coordinates": [[[72.88, 19.18], [72.89, 19.18], [72.89, 19.19], [72.88, 19.19], [72.88, 19.18]]]
      }
    }
  ]
}
```

---

## Roadmap 🛣️

- [ ] Complete end-to-end blockchain anchoring for spatial artifact provenance.
- [ ] Add IPFS CID persistence and retrieval workflows in the API layer.
- [ ] Integrate Polygon PoS smart contracts for immutable hash verification.
- [ ] Productionize 1991 archive ingestion via PDF georeferencing and vector extraction.
- [ ] Add CI/CD pipelines for automated linting, testing, and container release workflows.

---

## Repository Layout 📁

```text
apps/
  api/      # Go + Fiber backend
  web/      # Next.js frontend
packages/
  database/ # PostGIS schema bootstrap

data/
  raw/      # source geospatial files
  scripts/  # ETL and mock data generation
```

---

## Engineering Note 🤝

UrbanMemory is built as an open civic infrastructure project with production-oriented standards in reliability, traceability, and spatial data integrity. Contributions are welcome from GIS engineers, civic technologists, urban researchers, and Web3 infrastructure contributors.
