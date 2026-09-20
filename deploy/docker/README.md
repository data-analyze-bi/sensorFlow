# Docker deployment

This stack is for licensed SensorFlow customers and starts Go ingestion, Redis, ClickHouse, and Apache Superset.

Before startup, install the supplied executable decoder at `../../binaries/licenses/current/decoder` and its verification/license file in the same version directory. The licenses directory is mounted read-only.

Run `../../install.sh` from the repository root. It detects existing Redis and ClickHouse services, prompts only when an existing service may be reused, generates missing credentials, writes the private `.env`, and starts the required Compose services.

```bash
cd ../..
./install.sh
```

For a new machine, the installer generates Redis, ClickHouse, and Superset credentials. Existing Redis or ClickHouse credentials are collected interactively and passed consistently to ingestion and Superset.

Supported overrides include `REDIS_HOST`, `REDIS_TYPE`, `REDIS_PASSWORD`, `CLICKHOUSE_HOST`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB`, `SUPERSET_SECRET_KEY`, `SUPERSET_ADMIN_PASSWORD`, `CLICKHOUSE_SQLALCHEMY_URI`, and the published port variables.

All published ports bind to `127.0.0.1` by default. Set your own credentials before shared or production deployment, configure TLS and backups, and keep decoder/license files outside Git.

Stop without deleting named volumes:

```bash
docker compose down
```

中文版本：[README.zh-CN.md](README.zh-CN.md)
