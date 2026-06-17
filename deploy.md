# Deployment (Ubuntu 24.04)

## Prerequisites

- Go 1.24+
- PostgreSQL 16+
- systemd

## Build

cd /opt/marketplace-data-loader
go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync

## Environment

cp .env.example .env

# Edit with production values

## Database

createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql

## Systemd Services

### API service (/etc/systemd/system/marketplace-api.service)

[Unit]
Description=Marketplace Data Loader API
After=network.target postgresql.service

[Service]
Type=simple
User=app
WorkingDirectory=/opt/marketplace-data-loader
ExecStart=/opt/marketplace-data-loader/bin/server
Restart=on-failure
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target

### Sync service template (/etc/systemd/system/marketplace-sync@.service)

[Unit]
Description=Marketplace Sync %i
After=network.target postgresql.service

[Service]
Type=oneshot
User=app
WorkingDirectory=/opt/marketplace-data-loader
ExecStart=/opt/marketplace-data-loader/bin/sync --entity=%i
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

## Scheduled Sync (systemd Timers)

Create a timer for each entity:

- marketplace-sync-ozon_orders.timer
- marketplace-sync-ozon_stocks.timer
- marketplace-sync-wb_orders.timer
- marketplace-sync-wb_remains.timer
- marketplace-sync-wb_cards.timer
- marketplace-sync-ms_stocks.timer

Example (/etc/systemd/system/marketplace-sync-ozon_orders.timer):

[Unit]
Description=Sync Ozon orders every 30 min

[Timer]
OnCalendar=\*:0/30
Persistent=true

[Install]
WantedBy=timers.target

Enable and start:

systemctl enable marketplace-api
systemctl start marketplace-api
systemctl enable marketplace-sync-ozon_orders.timer
systemctl start marketplace-sync-ozon_orders.timer

# repeat for all timers

## Logging

- API logs: journalctl -u marketplace-api -f
- Sync logs: journalctl -u marketplace-sync@ozon_orders -f
  All logs are JSON – forward to your aggregator (ELK / Loki).

## Verification

curl http://localhost:3000/api/health
/opt/marketplace-data-loader/bin/sync --entity=ozon_orders
journalctl -u marketplace-sync@ozon_orders --since "5 min ago"

## Environment in Production

Set APP_ENV=production and adjust DB_POOL_MAX etc. in .env.
