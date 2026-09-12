<p align="center">
  <img src="assets/logo-sensorflow.svg" alt="SensorFlow — self-hosted event analytics" width="720">
</p>

# SensorFlow

Self-hosted event analytics from collection to visualization:

**Sensors Data SDKs -> ingestion service -> ClickHouse -> Apache Superset**

![SensorFlow operations overview](assets/operations-overview.svg)

Designed for simple deployment, low server overhead, predictable pricing, Sensors Data SDK compatibility, and fully customizable Apache Superset BI.

Languages: [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **Third-party SDK notice:** Sensors Data SDKs are not distributed by this repository. SensorFlow is independent from and is not endorsed or certified by Sensors Data. Obtain SDKs from official sources and comply with their licenses and terms. See [Third-Party SDK and Trademark Notice](THIRD_PARTY_NOTICES.md).

## Components

- Official Sensors Data SDKs collect events from Android, iOS, Web, mini programs, and server applications.
- The Go ingestion service validates and processes incoming events.
- ClickHouse stores event and user data for analytical queries.
- Apache Superset provides SQL exploration, charts, and dashboards.
- Redis supports ingestion caching.

## Start the data stack

Install Docker Engine or Docker Desktop with Docker Compose, then run:

```bash
cd deploy/docker
docker compose up -d --build
docker compose ps
```

Compose starts only the infrastructure. Default endpoints:

- ClickHouse HTTP: `http://127.0.0.1:8123`
- ClickHouse native protocol: `127.0.0.1:9000`
- Redis: `127.0.0.1:6379`
- Superset: `http://127.0.0.1:8088`

Set database, Redis, Superset administrator, and `SUPERSET_SECRET_KEY` environment variables before starting Compose. See the Docker deployment guide for the required values.

## Start ingestion

Go 1.17 or newer is required. The Go service runs separately and listens on `8081` by default. Review the Redis and ClickHouse settings in `conf/app.conf`, then run:

```bash
go mod download
go run main.go
```

## Connect an application

This repository does not publish a custom client SDK. Use the official Sensors Data SDK for your platform and set its `serverUrl` to:

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

See [Sensors Data SDK integration](docs/sensors-sdk.md) for platform coverage and verification steps.

## Authorization and managed services

For authorization, hosted operations, and commercial plans, visit [sensorflow.site/plans](https://sensorflow.site/plans). Deployment and account documentation is available at [sensorflow.site/docs](https://sensorflow.site/docs).

## Query in Superset

Open `http://127.0.0.1:8088`, select the configured ClickHouse database in SQL Lab, and run:

```sql
SELECT event, count() AS event_count
FROM sensors.event
GROUP BY event
ORDER BY event_count DESC
LIMIT 20;
```

The deployment also provisions an event overview dashboard with event totals, unique devices, trends, and top events.

## Verify and build

```bash
go test ./...
go build -o sensors main.go
```

See [FAQ](FAQ.md) for common issues and [Docker deployment](deploy/docker/README.md) for infrastructure details. Never commit real passwords, tokens, or customer data. See [LICENSE](LICENSE).
