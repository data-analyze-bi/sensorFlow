# Self-hosted Product Analytics Deployment

SensorFlow runs as a self-managed analytics stack: a Go ingestion service receives events, ClickHouse stores them, Redis supports ingestion caching, and Apache Superset provides dashboards and SQL exploration.

## Production checklist

- Set non-default database, Redis, and Superset credentials.
- Set a strong `SUPERSET_SECRET_KEY`.
- Put the ingestion endpoint behind HTTPS and a reverse proxy.
- Restrict ClickHouse, Redis, and Superset ports to trusted networks.
- Back up ClickHouse data and test restoration.
- Never commit tokens, passwords, customer data, or license files.

See the [Docker deployment guide](../deploy/docker/README.md) for service configuration.
