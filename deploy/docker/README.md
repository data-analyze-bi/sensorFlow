# Docker deployment

The deployment has two stages. Stage 1 starts Redis, ClickHouse, and Apache Superset, imports labeled demo events, and creates the demo dashboard without requiring a license. Stage 2 installs the purchased decoder/license and starts Go ingestion last.

Run `../../install.sh` from the repository root. It detects existing Redis and ClickHouse services, generates missing credentials, writes the private `.env`, imports demo data, and starts Superset. It does not request a decoder or start ingestion.

```bash
cd ../..
./install.sh
```

After evaluating the dashboard, download the decoder/license from `sensorflow.site` and activate real ingestion:

```bash
./activate.sh
```

For a new machine, the installer generates Redis, ClickHouse, and Superset credentials. Existing Redis or ClickHouse credentials are collected interactively and passed consistently to Superset and, after activation, ingestion.

Supported overrides include `REDIS_HOST`, `REDIS_TYPE`, `REDIS_PASSWORD`, `CLICKHOUSE_HOST`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DB`, `SUPERSET_SECRET_KEY`, `SUPERSET_ADMIN_PASSWORD`, `CLICKHOUSE_SQLALCHEMY_URI`, and the published port variables.

All published ports bind to `127.0.0.1` by default. Set your own credentials before shared or production deployment, configure TLS and backups, and keep decoder/license files outside Git.

Stop without deleting named volumes:

```bash
docker compose down
```

中文版本：[README.zh-CN.md](README.zh-CN.md)
