# WARD

**WARD Assists Restarting Daemons** — lightweight service watchdog for OpenRC.

Periodically checks service status via `rc-service` and automatically restarts failed services with configurable cooldown and max restart limits.

## Features

- Periodic service status checks
- Automatic restart on failure
- Configurable cooldown to prevent restart loops
- Max restart limit per service
- YAML configuration
- Logging to file or stdout
- Graceful shutdown on SIGINT/SIGTERM
- Hot reload configuration (SIGHUP)
- Log rotation support (SIGUSR1)
- PID file for process tracking
- Telegram notifications
- Prometheus metrics (per-service)
- CLI commands (status, list, init)

## Installation

### From source

```sh
go build -o ward ./cmd/ward/
sudo install -m 755 ward /usr/local/bin/ward
```

### Using install script

```sh
sudo ./install.sh
```

### Using ebuild (Gentoo overlay)

#### Adding the overlay

1. Create the overlay directory structure:

```sh
sudo mkdir -p /var/db/repos/ward/{app-admin/ward/files,metadata}
```

2. Create `layout.conf`:

```sh
sudo tee /var/db/repos/ward/metadata/layout.conf <<'EOF'
masters = gentoo
auto-sync = false
repo-name = ward
thin-manifests = true
EOF
```

3. Copy the ebuild and auxiliary files from the repository:

```sh
# From the cloned repository
cp app-admin/ward/ward-0.1.0.ebuild /var/db/repos/ward/app-admin/ward/
cp app-admin/ward/files/* /var/db/repos/ward/app-admin/ward/files/
```

4. Register the overlay in portage:

```sh
sudo tee /etc/portage/repos.conf/ward.conf <<'EOF'
[ward]
location = /var/db/repos/ward
masters = gentoo
auto-sync = false
priority = 50
EOF
```

5. Generate the Manifest:

```sh
cd /var/db/repos/ward/app-admin/ward
sudo ebuild ward-0.1.0.ebuild manifest
```

6. Install:

```sh
emerge app-admin/ward
```

#### Service management

```sh
# Add to autostart
sudo rc-update add ward default

# Start
sudo rc-service ward start

# Stop
sudo rc-service ward stop

# Reload configuration
sudo rc-service ward reload

# Check status
ward status
```

## Configuration

Default path: `/etc/ward/config.yaml`

```yaml
check_interval: 10s
log_file: /var/log/ward.log

notification:
  enabled: false
  telegram:
    bot_token: ""
    chat_id: ""

metrics:
  enabled: false
  listen: ":9090"

services:
  - name: nginx
    restart_on_fail: true
    max_restarts: 3
    restart_cooldown: 30s

  - name: redis
    restart_on_fail: true
    max_restarts: 5
    restart_cooldown: 10s

  - name: sshd
    restart_on_fail: false
```

### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `check_interval` | duration | 10s | How often to check services |
| `log_file` | string | stdout | Log file path |
| `notification.enabled` | bool | false | Enable Telegram notifications |
| `notification.telegram.bot_token` | string | | Telegram bot token |
| `notification.telegram.chat_id` | string | | Telegram chat ID |
| `metrics.enabled` | bool | false | Enable Prometheus metrics |
| `metrics.listen` | string | | Listen address (e.g. `:9090`) |
| `services[].name` | string | | OpenRC service name |
| `services[].restart_on_fail` | bool | false | Auto-restart on failure |
| `services[].max_restarts` | int | 3 | Max restart attempts |
| `services[].restart_cooldown` | duration | 30s | Min time between restarts |

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `/etc/ward/config.yaml` | Config file path |
| `-pid` | `/run/ward.pid` | PID file path |

## CLI Commands

```sh
# Show status of all services
ward status

# List monitored services
ward list

# Generate default config
ward init > /etc/ward/config.yaml

# Print version
ward -version

# Show help
ward -h
```

## Notification

### Telegram

1. Create bot via [@BotFather](https://t.me/BotFather)
2. Get bot token
3. Add bot to chat/group
4. Get chat_id via `https://api.telegram.org/bot<TOKEN>/getUpdates`

```yaml
notification:
  enabled: true
  telegram:
    bot_token: "123456:ABC-DEF..."
    chat_id: "-1001234567890"
```

Notifications sent on:
- Service restart: `Restarted (attempt 2/3)`
- Max restarts exceeded: `Exceeded max restarts (3), no longer restarting`
- Restart failed: `Restart failed: exit status 1`

## Metrics

```yaml
metrics:
  enabled: true
  listen: ":9090"
```

Endpoints:
- `GET /metrics` — Prometheus metrics
- `GET /health` — Health check

Example Prometheus config:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "ward"
    static_configs:
      - targets: ["localhost:9090"]
```

Available metrics:
- `ward_service_status` — Current status (1=running, 0=stopped)
- `ward_service_restarts_total` — Total restarts
- `ward_service_restart_failures_total` — Failed restarts
- `ward_service_restart_count` — Current restart count
- `ward_service_max_restarts` — Configured max restarts
- `ward_service_last_restart_timestamp` — Unix timestamp of last restart
- `ward_service_cooldown_seconds` — Configured cooldown
- `ward_service_restart_on_fail` — Auto-restart enabled (1=yes, 0=no)
- `ward_config_services` — Number of configured services
- `ward_check_interval_seconds` — Check interval

## Log Rotation

Ward supports log rotation via `SIGUSR1` without restart. Send the signal to reopen the log file:

```sh
kill -USR1 $(cat /run/ward.pid)
```

Example logrotate config (installed via ebuild to `/etc/logrotate.d/ward`):

```
/var/log/ward.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        kill -USR1 $(cat /run/ward.pid) 2>/dev/null || true
    endscript
}
```

## Usage

```sh
# Run with default config
sudo ward

# Run with custom config
sudo ward -config /path/to/config.yaml
```

## OpenRC Service

```sh
sudo rc-service ward start
sudo rc-service ward stop
sudo rc-service ward reload
sudo rc-update add ward default
```

## Building

```sh
# Build
go build -o ward ./cmd/ward/

# Build with version
go build -ldflags "-s -w -X main.version=v0.1.0" -o ward ./cmd/ward/

# Cross-compile
GOOS=linux GOARCH=arm64 go build -o ward-arm64 ./cmd/ward/
```

## License

[MIT](LICENSE)
