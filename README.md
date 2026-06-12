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

```sh
emerge app-admin/ward
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
