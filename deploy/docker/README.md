# Docker Deployment

This stack runs MySQL, Redis, ClickHouse, and Apache Superset. Run the Go ingestion service separately from the repository root.

```bash
export MYSQL_ROOT_PASSWORD='replace-with-strong-password'
export MYSQL_PASSWORD='replace-with-strong-password'
export REDIS_PASSWORD='replace-with-strong-password'
export CLICKHOUSE_PASSWORD='replace-with-strong-password'
export SUPERSET_SECRET_KEY='replace-with-long-random-secret'
export SUPERSET_ADMIN_PASSWORD='replace-with-strong-password'
export CLICKHOUSE_SQLALCHEMY_URI='clickhousedb://default:replace-with-url-encoded-password@clickhouse:8123/sensors'
docker compose up -d --build
docker compose ps
```

Wait until all services are healthy. Open Superset at `http://127.0.0.1:8088` and use the administrator password configured above.

The startup process initializes Superset, creates the administrator, configures ClickHouse, and provisions the `sensors.event` dataset and event overview dashboard.

Before production deployment, override `SUPERSET_SECRET_KEY`, `SUPERSET_ADMIN_PASSWORD`, and database credentials with environment variables. Restrict database ports and enable backups.

Stop without deleting data:

```bash
docker compose down
```

中文版本：[README.zh-CN.md](README.zh-CN.md)
