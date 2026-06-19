# Production Deployment Specification (Ubuntu 24.04 LTS)

This document establishes the production deployment architecture and infrastructure instructions for executing the core synchronization data loader daemon and automated job schedulers on Ubuntu 24.04 LTS utilizing native Linux `systemd` unit services and persistent timers.

## Prerequisites

The bare-metal or virtual machine target runtime deployment environment requires the following underlying platform layers:

- **Runtime Engine**: Go v1.24 or higher
- **Database Engine**: PostgreSQL v16 or higher (Active cluster instance)
- **Init System**: systemd v255 or higher

## Binary Compilation Pipelines

1. Navigate directly to the synchronized deployment application deployment directory:
   ```bash
   cd /opt/marketplace-data-loader
   ```

2. Compile optimized production-grade static binary assets targeting the REST API server daemon and the localized execution utility:
   ```bash
   go build -o bin/server ./cmd/server
   go build -o bin/sync ./cmd/sync
   ```

## Infrastructure Environment Variables

Initialize your isolated production runtime environment state configuration parameters schema:
```bash
cp .env.example .env
```

Modify variables inside the `.env` footprint configuration file to match explicit production infrastructure specifications. Ensure `APP_ENV` parameter is explicitly mapped to `production` and `DB_POOL_MAX` is configured based on database node capabilities.

## Database Schema Initialization

Create the high-concurrency production database cluster instance, and execute automated state migration scripts:
```bash
createdb marketplace
psql -d marketplace -f migrations/001_init.up.sql
```

## Linux systemd Unit Lifecycle Configuration

### 1. REST API Daemon Unit Service Block

Provision the structural network service daemon context tracking runtime configurations at `/etc/systemd/system/marketplace-api.service`:

```ini
[Unit]
Description=Marketplace Monitor Platform REST API Engine
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=app
Group=app
WorkingDirectory=/opt/marketplace-data-loader
ExecStart=/opt/marketplace-data-loader/bin/server
Restart=on-failure
RestartSec=10s
LimitNOFILE=65536
EnvironmentFile=/opt/marketplace-data-loader/.env

[Install]
WantedBy=multi-user.target
```

### 2. Parameterized Job Execution Template Service

Provision an atomic, stateless parameterized unit mapping at `/etc/systemd/system/marketplace-sync@.service` to process individual synchronization pipeline entities on-demand or via triggers:

```ini
[Unit]
Description=Marketplace Monitor Task Runner (%i)
After=network.target postgresql.service

[Service]
Type=oneshot
User=app
Group=app
WorkingDirectory=/opt/marketplace-data-loader
ExecStart=/opt/marketplace-data-loader/bin/sync --entity=%i
StandardOutput=journal
StandardError=journal
EnvironmentFile=/opt/marketplace-data-loader/.env
```

## Persistent Automated Processing (systemd Timers)

Background highload extraction processing loops are orchestrated cleanly via independent decoupled `systemd` timer targets.

### Target Timer Component Template

Create an exact discrete timer configuration file mapping every target entity payload. 

Reference configuration sample located at `/etc/systemd/system/marketplace-sync-ozon_orders.timer`:

```ini
[Unit]
Description=Trigger Ozon Orders Synchronization Pipeline Every 30 Minutes

[Timer]
OnCalendar=*:0/30
Persistent=true
Unit=marketplace-sync@ozon_orders.service

[Install]
WantedBy=timers.target
```

### Infrastructure Schedulers List

Replicate the exact configuration syntax skeleton layout block above across the following individual destination timer units, tweaking the `OnCalendar` cron expression token according to target marketplace API rate-limit boundaries:

- `/etc/systemd/system/marketplace-sync-ozon_orders.timer` (Runs every 30 minutes)
- `/etc/systemd/system/marketplace-sync-ozon_stocks.timer` (Runs every 15 minutes)
- `/etc/systemd/system/marketplace-sync-wb_orders.timer`   (Runs every 30 minutes)
- `/etc/systemd/system/marketplace-sync-wb_remains.timer`  (Runs every 15 minutes)
- `/etc/systemd/system/marketplace-sync-wb_cards.timer`    (Runs once daily)
- `/etc/systemd/system/marketplace-sync-ms_stocks.timer`   (Runs every 10 minutes)

### Pipeline Initialization

Reload the structural systemd supervisor daemon engine configurations metadata to register newly created unit paths, then activate the persistent task loops:

```bash
# Reload internal daemon state mapping
systemctl daemon-reload

# Mount and activate core REST API service
systemctl enable --now marketplace-api.service

# Mount and activate full automated scheduler timer blocks
systemctl enable --now marketplace-sync-ozon_orders.timer
systemctl enable --now marketplace-sync-ozon_stocks.timer
systemctl enable --now marketplace-sync-wb_orders.timer
systemctl enable --now marketplace-sync-wb_remains.timer
systemctl enable --now marketplace-sync-wb_cards.timer
systemctl enable --now marketplace-sync-ms_stocks.timer
```

## System Verification & Infrastructure Diagnostics

### API Endpoint Health Probe

Verify active server daemon interface response payloads:
```bash
curl -I http://localhost:3000/api/health
```

### Manual Job Validation Sequence

Verify transaction processing pipeline stability and PostgreSQL storage mutations by firing a direct immediate manual synchronization task runner block:
```bash
/opt/marketplace-data-loader/bin/sync --entity=ozon_orders
```

### Telemetry & Log Inspection

All background process tasks emit highload structured logs to standard I/O channels. To tail logs emitted by the main API server or individual scheduler execution runners, utilize `journalctl`:

```bash
# Audit real-time REST API server connection streams
journalctl -u marketplace-api.service -f

# Audit localized history context data from specific batch tasks
journalctl -u marketplace-sync@ozon_orders.service --since "5 min ago"
```
