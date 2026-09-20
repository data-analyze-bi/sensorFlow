# Docker deployment

This stack is for licensed SensorFlow customers and starts Go ingestion, Redis, ClickHouse, and Apache Superset.

Before startup, install the supplied executable decoder at `../../binaries/licenses/current/decoder` and its verification/license file in the same version directory. The licenses directory is mounted read-only.

```bash
test -x ../../binaries/licenses/current/decoder
cp .env.example .env
# Set your own SUPERSET_SECRET_KEY and SUPERSET_ADMIN_PASSWORD in .env.
docker compose up -d --build
docker compose ps
docker compose logs ingestion
```

The stack does not define customer passwords. Empty `REDIS_PASSWORD` and `CLICKHOUSE_PASSWORD` values start the bundled local databases without authentication; non-empty values enable authentication and are passed to ingestion. When ClickHouse authentication is enabled, also set `CLICKHOUSE_SQLALCHEMY_URI` with the same URL-encoded credentials for Superset.

Supported overrides include `REDIS_HOST`, `REDIS_TYPE`, `REDIS_PASSWORD`, `CLICKHOUSE_HOST`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB`, `SUPERSET_SECRET_KEY`, `SUPERSET_ADMIN_PASSWORD`, `CLICKHOUSE_SQLALCHEMY_URI`, and the published port variables.

All published ports bind to `127.0.0.1` by default. Set your own credentials before shared or production deployment, configure TLS and backups, and keep decoder/license files outside Git.

Stop without deleting named volumes:

```bash
docker compose down
```

中文版本：[README.zh-CN.md](README.zh-CN.md)
