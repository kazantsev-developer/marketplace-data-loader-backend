# Deployment Guide (Ubuntu 24.04)

This document provides technical instructions for deploying the backend data loader synchronization service on Ubuntu 24.04 LTS using systemd services and timers.

## Prerequisites

The deployment environment requires the following software versions:

- Go 1.24 or higher
- PostgreSQL 16 or higher
- systemd

## Build Instructions

1. Navigate to the application root directory:

cd /opt/marketplace-data-loader

2. Compile production binaries for the API server and execution utility:

go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync

## Environment Configuration

Initialize the production environment configuration file:

cp .env.example .env

Modify the parameters inside `.env` to match production specifications. Ensure `APP_ENV` is set to `production` and `DB_POOL_MAX` is tuned for expected concurrency levels.

## Database Initialization

Create the production database structure and execute schema definition migrations:

createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql

## systemd Unit Configuration

### API Service Block

Create the systemd unit file for the API daemon at `/etc/systemd/system/marketplace-api.service`:

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

### Parameterized Synchronization Template

Create a parameterized unit file at `/etc/systemd/system/marketplace-sync@.service` to process individual synchronization entities:

[Unit]
Description=Marketplace Synchronization Task (%i)
After=network.target postgresql.service

[Service]
Type=oneshot
User=app
WorkingDirectory=/opt/marketplace-data-loader
ExecStart=/opt/marketplace-data-loader/bin/sync --entity=%i
StandardOutput=journal
StandardError=journal

## Scheduled Execution (systemd Timers)

Automated background processing is managed via independent systemd timer files.

### Template Definition

Create a timer configuration file for each separate sync entity.

Example configuration file `/etc/systemd/system/marketplace-sync-ozon_orders.timer`:

[Unit]
Description=Trigger Ozon Orders Synchronization Every 30 Minutes

[Timer]
OnCalendar=\*:0/30
Persistent=true

[Install]
WantedBy=timers.target

### Target Execution List

Replicate the configuration block above for the following destination timer files, adjusting the `OnCalendar` expression as demanded by API rate limits:

- `/etc/systemd/system/marketplace-sync-ozon_orders.timer`
- `/etc/systemd/system/marketplace-sync-ozon_stocks.timer`
- `/etc/systemd/system/marketplace-sync-wb_orders.timer`
- `/etc/systemd/system/marketplace-sync-wb_remains.timer`
- `/etc/systemd/system/marketplace-sync-wb_cards.timer`
- `/etc/systemd/system/marketplace-sync-ms_stocks.timer`

### Service Initialization

Reload the systemd manager configuration to recognize the new unit definitions, then enable and start the execution loop:

systemctl daemon-reload

systemctl enable marketplace-api.service
systemctl start marketplace-api.service

systemctl enable --now marketplace-sync-ozon_orders.timer
systemctl enable --now marketplace-sync-ozon_stocks.timer
systemctl enable --now marketplace-sync-wb_orders.timer
systemctl enable --now marketplace-sync-wb_remains.timer
systemctl enable --now marketplace-sync-wb_cards.timer
systemctl enable --now marketplace-sync-ms_stocks.timer

## Verification and Diagnostics

### API Health Check

curl -I http://localhost:3000/api/health

### Manual Job Validation

Verify job processing and DB write sequences by running a direct manual sync execution:

/opt/marketplace-data-loader/bin/sync --entity=ozon_orders

### Log Inspection

All service components emit structured JSON logs to stdout/stderr. To monitor the API server or individual synchronization runners, use `journalctl`:

journalctl -u marketplace-api.service -f
journalctl -u marketplace-sync@ozon_orders.service --since "5 min ago"
