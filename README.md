<p align="center">
  <img src="assets/logo-sensorflow.svg" alt="SensorFlow — open-source self-hosted event analytics" width="720">
</p>

# SensorFlow

**Open-source, self-hosted event analytics. Start with a working Superset demo, then activate licensed Sensors Data SDK ingestion when ready.**

```text
Sensors Data SDKs ──> Go ingestion ──> ClickHouse ──> Apache Superset
                              └──────> Redis
```

SensorFlow is independent from Sensors Data and is not endorsed or certified by it. Official Sensors Data SDKs are not distributed by this repository.

[中文](README.zh-CN.md) · [Documentation](docs/getting-started.md) · [Website](https://sensorflow.site/) · [Apache-2.0 License](LICENSE)

## Customer deployment

## Stage 1: deploy and view the demo

Prerequisites:

- Docker Engine or Docker Desktop with Docker Compose v2

Clone the repository and run the installer:

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
./install.sh
```

This stage does not require a license. The installer detects Redis and ClickHouse, starts missing services, imports clearly labeled `demo_*` events, creates the Superset dataset and dashboard, and prints the Superset login. The demo dashboard includes events, users, DAU, new users, purchasers, demo GMV, conversion, hourly/daily trends, funnel stages, channels, pages, countries, operating systems, networks, and app versions. It does not start the real SDK ingestion service.

After installation, allow TCP 8088 in the server/cloud firewall, open the server IP URL printed by the installer, and use the red administrator credentials to view the demo. Once verified, run `./activate.sh` to start real ingestion. Caddy is an optional deployment component for domains and HTTPS.

## Stage 2: activate real event ingestion

Download the SensorFlow license from [sensorflow.site](https://sensorflow.site/), leave the downloaded license files in `~/Downloads`, then run:

```bash
./activate.sh
```

Activation automatically installs the license at `binaries/sensors-payload-license` and its verification file at `binaries/sensors-payload-license.verify.json`, then starts ingestion last. Existing Redis, ClickHouse, Superset, demo data, and private configuration are preserved.

Superset is reachable through the server IP; other ports bind to `127.0.0.1` by default. Go ingestion starts only after activation:

- Ingestion: `127.0.0.1:8081`
- ClickHouse HTTP/native: `127.0.0.1:8123` / `127.0.0.1:9000`
- Redis: `127.0.0.1:6379`
- Superset: `http://YOUR_SERVER_IP:8088` (binds to `0.0.0.0`)

The installer fills missing secrets, preserves existing credentials on retries, and highlights the administrator login in red. See [domain and HTTPS setup](deploy/docker/README.md#domain-and-https).

## Connect a Sensors Data SDK after activation

Keep the existing SDK instrumentation and point its upload URL to SensorFlow:

```text
https://your-sensorflow.example/sensors/send/?token=YOUR_TOKEN
```

Send a uniquely named integration event from the real client, then verify the stored row:

```bash
cd deploy/docker
docker compose exec -T clickhouse clickhouse-client \
  ${CLICKHOUSE_PASSWORD:+--password "$CLICKHOUSE_PASSWORD"} \
  --query "SELECT time, event, distinct_id FROM sensors.event WHERE event = 'integration_test' ORDER BY time DESC LIMIT 10"
```

Open Superset at the installer's `http://YOUR_SERVER_IP:8088` URL. The installer prints the generated administrator password once startup completes.

## Production requirements

The stack does not define Redis, ClickHouse, MySQL, or customer account passwords. Redis and ClickHouse start without passwords only when their variables are empty and their ports remain bound to `127.0.0.1`. Before production:

- Set `REDIS_PASSWORD`, `CLICKHOUSE_PASSWORD`, `SUPERSET_ADMIN_PASSWORD`, `SUPERSET_SECRET_KEY`, and a matching URL-encoded `CLICKHOUSE_SQLALCHEMY_URI`.
- Terminate TLS in a reverse proxy and expose only the required ingestion path.
- Keep database ports private, configure backups, and monitor license processing failures, ingestion latency, and ClickHouse disk usage.
- Pin and review container image versions according to your release policy.
- Validate SDK version, encrypted payload plugins, identity behavior, property types, and event timestamps before moving traffic.

See [Docker deployment](deploy/docker/README.md), [Sensors Data SDK integration](docs/sensors-sdk.md), [FAQ](FAQ.md), and [third-party notices](THIRD_PARTY_NOTICES.md).

## Development checks

```bash
go test ./...
go build -o sensors main.go
docker compose -f deploy/docker/docker-compose.yml config
```

Contributions are described in [CONTRIBUTING.md](CONTRIBUTING.md). Never commit customer data, credentials, or license files.
