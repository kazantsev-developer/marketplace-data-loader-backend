# Marketplace Data Loader

A high-performance data synchronization service designed to aggregate orders, stock levels, and product metadata from Ozon, Wildberries, and MoySklad. The service exposes a REST API for consumption by frontend monitoring dashboards.

## Tech Stack

- Runtime: Go 1.24 or higher
- Database: PostgreSQL 16 or higher
- Containerization: Docker / Docker Compose

## Features

- Ozon Integration: Synchronizes FBO/FBS orders and real-time FBO stock data.
- Wildberries Integration: Executes full synchronization of orders, stock levels, and product cards.
- MoySklad Integration: Captures warehouse stock snapshots and constructs product aggregates.
- REST API Layer: Provides filtered lists, aggregated statistics, and time-series data for charts.
- Execution Logging: Persists comprehensive metadata for every synchronization job, including status, processed item counts, and precise duration.

## Quick Start

### Repository Initialization

git clone <repo-url>
cd marketplace-data-loader-backend

### Compilation

Compile the API server and the synchronization CLI binary:

go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync

### Database Initialization

Ensure a PostgreSQL instance is running, create the target database, and apply the initial schema migrations:

createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql

### Configuration

Copy the example environment configuration file and populate it with valid credentials:

cp .env.example .env

### Execution

Start the REST API server:

./bin/server

Execute a manual synchronization job for a specific entity:

./bin/sync --entity=ozon_orders

Supported arguments for the `--entity` flag:

- `ozon_orders`
- `ozon_stocks`
- `wb_orders`
- `wb_remains`
- `wb_cards`
- `ms_stocks`

### Containerized Execution

To run the application ecosystem via Docker Compose:

docker compose up -d

## API Endpoints

### Ozon

- `GET /api/ozon/orders` — Retrieves FBO/FBS orders with filtering and pagination.
- `GET /api/ozon/orders/stats` — Provides aggregated order statistics broken down by fulfillment scheme.
- `GET /api/ozon/remains` — Returns the current FBO stock list.
- `GET /api/ozon/remains/stats` — Provides stock metrics filtered by top brands.

### Wildberries

- `GET /api/wb/orders` — Retrieves orders with date filtering and pagination.
- `GET /api/wb/orders/stats` — Returns aggregated sales and volume statistics.
- `GET /api/wb/remains` — Returns stock availability grouped by warehouse.
- `GET /api/wb/cards` — Lists product cards with text search and pagination.
- `GET /api/wb/cards/stats` — Provides product card distribution analytics.

### MoySklad

- `GET /api/moysklad/stocks` — Provides inventory status filtered by store and product.
- `GET /api/moysklad/aggregates` — Returns structural product aggregates.
- `GET /api/moysklad/stores` — Returns a list of active warehouses.

### System

- `GET /api/health` — Returns current service health status.
- `GET /api/sync/logs` — Retrieves historical execution logs of synchronization jobs.
- `GET /api/dashboard/stats` — Fetches global summary metrics for the dashboard view.
- `GET /api/charts/orders-daily` — Outputs daily time-series data for Wildberries orders.
- `GET /api/charts/ozon-orders-daily` — Outputs daily time-series data for Ozon orders.

## Environment Variables

Configuration parameters must be defined within the `.env` file based on the following reference structure:

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

## Deployment

Refer to `deploy.md` for production architecture guidelines, including native systemd service provisioning and automated execution via systemd timers.
