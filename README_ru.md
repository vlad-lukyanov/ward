# WARD

**WARD Assists Restarting Daemons** — лёгкий сторожевой сервис для OpenRC.

Периодически проверяет статус сервисов через `rc-service` и автоматически перезапускает упавшие с настраиваемой задержкой и лимитом попыток.

## Возможности

- Периодическая проверка статуса сервисов
- Автоматический перезапуск при падении
- Защита от restart-спирали через cooldown
- Лимит перезапусков на каждый сервис
- YAML-конфигурация
- Логирование в файл или stdout
- Корректное завершение по SIGINT/SIGTERM
- Hot reload конфигурации (SIGHUP)
- Ротация логов без перезапуска (SIGUSR1)
- PID-файл для отслеживания процесса
- Telegram уведомления
- Prometheus метрики (по каждому сервису)
- CLI команды (status, list, init)

## Установка

### Из исходников

```sh
go build -o ward ./cmd/ward/
sudo install -m 755 ward /usr/local/bin/ward
```

### Через скрипт установки

```sh
sudo ./install.sh
```

### Через ebuild (Gentoo overlay)

```sh
emerge app-admin/ward
```

## Конфигурация

Путь по умолчанию: `/etc/ward/config.yaml`

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

### Параметры

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `check_interval` | duration | 10s | Интервал проверки сервисов |
| `log_file` | string | stdout | Путь к файлу логов |
| `notification.enabled` | bool | false | Включить Telegram уведомления |
| `notification.telegram.bot_token` | string | | Токен Telegram бота |
| `notification.telegram.chat_id` | string | | ID чата Telegram |
| `metrics.enabled` | bool | false | Включить Prometheus метрики |
| `metrics.listen` | string | | Адрес для прослушивания (напр. `:9090`) |
| `services[].name` | string | | Имя сервиса в OpenRC |
| `services[].restart_on_fail` | bool | false | Автоперезапуск при падении |
| `services[].max_restarts` | int | 3 | Макс. количество перезапусков |
| `services[].restart_cooldown` | duration | 30s | Минимальный интервал между перезапусками |

### Флаги CLI

| Флаг | По умолчанию | Описание |
|------|--------------|----------|
| `-config` | `/etc/ward/config.yaml` | Путь к конфигу |
| `-pid` | `/run/ward.pid` | Путь к PID-файлу |

## CLI команды

```sh
# Показать статус всех сервисов
ward status

# Список отслеживаемых сервисов
ward list

# Сгенерировать конфиг по умолчанию
ward init > /etc/ward/config.yaml

# Показать версию
ward -version

# Показать справку
ward -h
```

## Уведомления

### Telegram

1. Создайте бота через [@BotFather](https://t.me/BotFather)
2. Получите токен бота
3. Добавьте бота в чат/группу
4. Узнайте chat_id через `https://api.telegram.org/bot<TOKEN>/getUpdates`

```yaml
notification:
  enabled: true
  telegram:
    bot_token: "123456:ABC-DEF..."
    chat_id: "-1001234567890"
```

Уведомления отправляются при:
- Перезапуске сервиса: `Restarted (attempt 2/3)`
- Превышении лимита перезапусков: `Exceeded max restarts (3), no longer restarting`
- Неудачном перезапуске: `Restart failed: exit status 1`

## Метрики

```yaml
metrics:
  enabled: true
  listen: ":9090"
```

Эндпоинты:
- `GET /metrics` — Prometheus метрики
- `GET /health` — Проверка работоспособности

Пример конфига Prometheus:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "ward"
    static_configs:
      - targets: ["localhost:9090"]
```

Доступные метрики:
- `ward_service_status` — Текущий статус (1=работает, 0=остановлен)
- `ward_service_restarts_total` — Общее количество перезапусков
- `ward_service_restart_failures_total` — Количество неудачных перезапусков
- `ward_service_restart_count` — Текущее количество перезапусков
- `ward_service_max_restarts` — Настроенный лимит перезапусков
- `ward_service_last_restart_timestamp` — Unix timestamp последнего перезапуска
- `ward_service_cooldown_seconds` — Настроенный cooldown
- `ward_service_restart_on_fail` — Автоперезапуск включен (1=да, 0=нет)
- `ward_config_services` — Количество настроенных сервисов
- `ward_check_interval_seconds` — Интервал проверки

## Ротация логов

Ward поддерживает ротацию логов через `SIGUSR1` без перезапуска. Отправьте сигнал для переоткрытия лог-файла:

```sh
kill -USR1 $(cat /run/ward.pid)
```

Пример конфига logrotate (устанавливается ebuild'ом в `/etc/logrotate.d/ward`):

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

## Использование

```sh
# Запуск с конфигом по умолчанию
sudo ward

# Запуск с произвольным конфигом
sudo ward -config /path/to/config.yaml
```

## Сервис OpenRC

```sh
sudo rc-service ward start
sudo rc-service ward stop
sudo rc-service ward reload
sudo rc-update add ward default
```

## Сборка

```sh
# Обычная сборка
go build -o ward ./cmd/ward/

# Сборка с версией
go build -ldflags "-s -w -X main.version=v0.1.0" -o ward ./cmd/ward/

# Кросс-компиляция
GOOS=linux GOARCH=arm64 go build -o ward-arm64 ./cmd/ward/
```

## Лицензия

[MIT](LICENSE)
