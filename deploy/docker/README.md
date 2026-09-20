# Docker deployment

This stack is for licensed SensorFlow customers and starts Go ingestion, Redis, ClickHouse, and Apache Superset.

Before startup, install the supplied executable decoder at `../../binaries/licenses/current/decoder` and its verification/license file in the same version directory. The licenses directory is mounted read-only.

```bash
test -x ../../binaries/licenses/current/decoder
docker compose up -d --build
docker compose ps
docker compose logs ingestion
```

Supported overrides include `REDIS_PASSWORD`, `CLICKHOUSE_PASSWORD`, `SUPERSET_SECRET_KEY`, `SUPERSET_ADMIN_PASSWORD`, `CLICKHOUSE_SQLALCHEMY_URI`, `SENSORFLOW_PORT`, `CLICKHOUSE_HTTP_PORT`, `CLICKHOUSE_NATIVE_PORT`, `REDIS_PORT`, and `SUPERSET_PORT`.

All published ports bind to `127.0.0.1` by default. Replace every default credential before shared or production deployment, configure TLS and backups, and keep decoder/license files outside Git.

Stop without deleting named volumes:

```bash
docker compose down
```

中文版本：[README.zh-CN.md](README.zh-CN.md)
