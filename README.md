# Marketplace Monitor Platform (Backend)

Core data synchronization engine and high-performance backend monolith for the Marketplace Monitor Platform ecosystem. Engineered with Go 1.24, pgx/v5, scany, and PostgreSQL 16. Provides a concurrent, low-allocation, and low-latency REST API to aggregate and stream real-time orders, warehouse stock telemetry, and product catalog metadata from Ozon, Wildberries, and MoySklad.

## Tech Stack

- **Runtime Environment** — Go 1.24 or higher (Concurrent execution pools & lightweight routines)
- **Database Driver** — pgx/v5 (High-performance, low-allocation PostgreSQL driver & connection pool)
- **Data Mapping** — scany/v2 (Efficient database rows-to-struct scanning utility)
- **Configuration** — envconfig (Atomic environment variable parsing layout)
- **Container Architecture** — Docker / Docker Compose orchestration layers

## Key Architectural Features

- **Ozon API Integration** — Automated extraction of FBO/FBS orders and concurrent real-time FBO inventory availability datasets.
- **Wildberries API Integration** — Full pipeline synchronization mapping transactional order metrics, warehouse remnants, and product cards metadata.
- **MoySklad Client Integration** — Point-in-time enterprise warehouse inventory snapshots capturing raw product aggregate metrics.
- **Transport API Layer** — Optimized endpoints delivering complex multi-tenant filters, analytical summary aggregations, and highload daily charts datasets.
- **Granular Execution Logging** — Explicit metadata persistence tracking every single sync task lifecycle, including exact items count delta, state tracking, and duration parameters.

## Core Setup Instructions

### Repository Initialization

Clone the tracking branch context and navigate into the project workspace boundary:
```bash
git clone https://github.com/kazantsev-developer/marketplace-data-loader-backend.git
cd marketplace-data-loader-backend
```

### Binary Compilation

Compile both the headless REST API network server and the synchronization Command Line Interface (CLI) executable binaries:
```bash
go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync
```

### Database Initialization Ledger

Ensure a local PostgreSQL instance is active, create the application ledger boundaries, and execute initial schema migrations:
```bash
createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql
```

### Infrastructure Configuration

Clone the baseline configuration variables schema layout file and insert functional merchant secret key parameters:
```bash
cp .env.example .env
```

### Execution Lifecycle

#### Run the REST API Server Engine
```bash
./bin/server
```

#### Run Manual Synchronizer Tasks via CLI
Execute an atomic data extraction job targeting specific merchant integration channels:
```bash
./bin/sync --entity=ozon_orders
```

Supported arguments for the deterministic `--entity` flag:
- `ozon_orders` — Sync Ozon FBO/FBS sales records
- `ozon_stocks` — Sync Ozon FBO inventory availability
- `wb_orders`   — Sync Wildberries sales volume metrics
- `wb_remains`  — Sync Wildberries warehouse remnants tracking
- `wb_cards`    — Sync Wildberries product card catalog content
- `ms_stocks`   — Sync MoySklad inventory ledger parameters

### Production Container Orchestration

To initialize stateless backend microservice nodes and database pooling topologies concurrently via Docker Compose:
```bash
docker compose up -d
```

## Core API Endpoints Specification

### Ozon Operations
- **GET** `/api/ozon/orders` — Fetches FBO/FBS orders incorporating filtering parameters and server pagination.
- **GET** `/api/ozon/orders/stats` — Aggregates real-time order statistics broken down by fulfillment scheme.
- **GET** `/api/ozon/remains` — Returns the current FBO stock inventory array list.
- **GET** `/api/ozon/remains/stats` — Provides top brand distribution stock volume analytics.

### Wildberries Operations
- **GET** `/api/wb/orders` — Fetches sales logs with explicit date queries and page boundaries.
- **GET** `/api/wb/orders/stats` — Evaluates aggregated marketplace financial sales and volume analytics.
- **GET** `/api/wb/remains` — Returns inventory stock balances segmented across distributed physical warehouses.
- **GET** `/api/wb/cards` — Full-text searchable product catalog card index tracking pagination.
- **GET** `/api/wb/cards/stats` — Visual item distribution analytics mapping across catalog branches.

### MoySklad Corporate Channels
- **GET** `/api/moysklad/stocks` — Provides inventory parameters cross-referenced by specific store and product keys.
- **GET** `/api/moysklad/aggregates` — Calculates deep enterprise structural product aggregates.
- **GET** `/api/moysklad/stores` — Lists all registered active system warehouses.

### System Control Core
- **GET** `/api/health` — Probes core microservice Liveness/Readiness lifecycle state.
- **GET** `/api/sync/logs` — Comprehensive historical analytics audit database for cron data jobs.
- **GET** `/api/dashboard/stats` — Compiles global consolidated summary index values for the root UI view.
- **GET** `/api/charts/orders-daily` — Daily analytical time-series logs tracking Wildberries sales volume.
- **GET** `/api/charts/ozon-orders-daily` — Daily analytical time-series logs tracking Ozon sales volume.

## Infrastructure Environment Configurations

Variables must be strictly populated inside the local `.env` configuration context file following this schema definition:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=marketplace
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
DB_POOL_MAX=20

OZON_CLIENT_ID=your_client_id
OZON_API_KEY=your_api_key

WB_API_TOKEN=your_wb_token

MS_TOKEN=your_moysklad_token

PORT=3000
APP_ENV=development
```

## Production Deployment Guidelines

For scalable environment provisioning layout operations, inspect the standalone architectural guidelines file `deploy.md`. This contains native instructions for mounting Linux `systemd` service abstractions and automation via `systemd.timer` job triggers.
