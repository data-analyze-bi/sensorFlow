# Getting Started with SensorFlow

SensorFlow is an open-source, self-hosted product analytics platform. This guide starts the ClickHouse, Redis, and Apache Superset data stack, then launches the Go event ingestion service.

## Prerequisites

- Docker Engine or Docker Desktop with Docker Compose
- Go 1.17 or newer
- A SensorFlow license installed at `binaries/sensors-payload-decoder` as described in the main README

## Start the stack

```bash
cd deploy/docker
docker compose up -d --build
docker compose ps
```

The default services are ClickHouse, Redis, and Apache Superset. Configure passwords and `SUPERSET_SECRET_KEY` before using this setup in production.

## Start ingestion

From the repository root:

```bash
go mod download
go run main.go
```

The ingestion service listens on port `8081` by default. See [Sensors Data SDK integration](sensors-sdk.md) to send events.
