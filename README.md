# Marketplace Data Loader

High-performance data synchronization service for Ozon, Wildberries, and MoySklad.  
Exposes a REST API for frontend dashboards and monitoring.

## Stack

- Go 1.24+history
- PostgreSQL 16+
- Docker (optional)

## Features

- **Ozon** – orders (FBO/FBS) and FBO stock
- **Wildberries** – orders, stock, product cards (full sync)
- **MoySklad** – warehouse stock snapshots and product aggregates
- **REST API** – filtered lists, statistics, charts
- **Sync logging** – every job is persisted with status, counts, and duration

## Quick Start

git clone <repo-url>
cd marketplace-data-loader-backend
go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync

createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql

cp .env.example .env # fill in all tokens

./bin/server # start API
./bin/sync --entity=ozon_orders # or others:

# available entities: ozon_orders, ozon_stocks, wb_orders, wb_remains, wb_cards, ms_stocks

Or with Docker Compose:
docker-compose up -d

## API Endpoints

### Ozon

- GET /api/ozon/orders – orders (FBO/FBS)
- GET /api/ozon/orders/stats – stats by scheme
- GET /api/ozon/remains – FBO stock list
- GET /api/ozon/remains/stats – stock stats with top brands

### Wildberries

- GET /api/wb/orders – orders list (filter by date, pagination)
- GET /api/wb/orders/stats – aggregated stats
- GET /api/wb/remains – stock levels by warehouse
- GET /api/wb/cards – product cards (search, pagination)
- GET /api/wb/cards/stats – card statistics

### MoySklad

- GET /api/moysklad/stocks – detailed stock by store and product
- GET /api/moysklad/aggregates – aggregated totals per product
- GET /api/moysklad/stores – warehouse list

### System

- GET /api/health – service health
- GET /api/sync/logs – sync job history
- GET /api/dashboard/stats – summary for dashboard
- GET /api/charts/orders-daily – daily WB order chart
- GET /api/charts/ozon-orders-daily – daily Ozon order chart

## Environment Variables

See .env.example:

DB_HOST=localhost
DB_PORT=5432
DB_NAME=marketplace
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable
DB_POOL_MAX=20

# Ozon

OZON_CLIENT_ID=xxx
OZON_API_KEY=xxx

# Wildberries

WB_API_TOKEN=xxx

# MoySklad

MS_TOKEN=xxx

PORT=3000
APP_ENV=development

## Deployment

See deploy.md for systemd services and timer setup.
