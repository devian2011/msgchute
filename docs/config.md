# Notification service configuration

### .env

```dotenv
APP_HTTP_PORT=:8080

APP_HTTP_WITH_SWAGGER=true

APP_CONFIG_PATH=./config/config.yml

# -------
# Store
# -------
POSTGRES_DB=notifier
POSTGRES_USER=notifier
POSTGRES_PASSWORD=notifier
POSTGRES_HOST=db.notifier.local
POSTGRES_PORT=15432

APP_DB_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
APP_DB_DRIVER=pgx
APP_DB_MIGRATION_DIR=./migrations

```

### config.yaml

Application supports many transports. You can use multiple smtp, file and other transports.

```yaml
http:
  port: ${APP_HTTP_PORT}
  withSwagger: false
  config:
    readTimeout: ${APP_HTTP_READ_TIMEOUT|1s}
    readHeaderTimeout: ${APP_HTTP_READ_HEADER_TIMEOUT|1s}
    writeTimeout: ${APP_HTTP_WRITE_TIMEOUT|2s}
  tls:
    cert: ${APP_HTTP_TLS_CERT}
    key: ${APP_HTTP_TLS_KEY}
db:
  dsn: ${APP_DB_DSN}
  driver: ${APP_DB_DRIVER|pgx}
  migrationDir: ${APP_DB_MIGRATION_DIR|./migrations}
providers:
  maxBufferSize: 1000
  fetchTaskTimeout: 2s
  pluginMap:
    telegram: ./plugins/dist/providers/tg
    smtp: ./plugins/dist/providers/smtp
  providers:
    gmail:
      provider: smtp
      settings:
        workers:
          min: 2
          max: 6
        breaker:
          windowSize: 10m
          minRequests: 100
          failureThreshold: 0.5
          timeout: 5m
      params:
        host: smtp.gmail.com
        port: 587
        user: user
        password: password
        from: mail@example.com
    tg:
      provider: telegram
      settings:
        workers:
          min: 2
          max: 2
        breaker:
          windowSize: 10m
          minRequests: 100
          failureThreshold: 0.3
          timeout: 5m
      params:
        token:

```
