<p align="center">
  <img src="assets/logo-sensorflow.svg" alt="SensorFlow — open-source self-hosted event analytics" width="720">
</p>

# SensorFlow

**Open-source, self-hosted event analytics for licensed SensorFlow customers. Keep existing Sensors Data SDK instrumentation and own the resulting ClickHouse data.**

```text
Sensors Data SDKs ──> Go ingestion ──> ClickHouse ──> Apache Superset
                              └──────> Redis
```

SensorFlow is independent from Sensors Data and is not endorsed or certified by it. Official Sensors Data SDKs are not distributed by this repository.

[中文](README.zh-CN.md) · [Documentation](docs/getting-started.md) · [Website](https://sensorflow.site/) · [Apache-2.0 License](LICENSE)

## Customer deployment

Prerequisites:

- Docker Engine or Docker Desktop with Docker Compose v2
- A valid SensorFlow payload decoder and verification/license file supplied after purchase
- An official Sensors Data SDK client configured for your application

Clone the repository, then install the supplied files under a versioned directory:

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
mkdir -p binaries/licenses/customer
mv ~/Downloads/sensors-payload-decoder-* binaries/licenses/customer/decoder
mv ~/Downloads/*.verify.json binaries/licenses/customer/decoder.verify.json
chmod +x binaries/licenses/customer/decoder
ln -sfn customer binaries/licenses/current
```

Do not commit decoder or license files. Compose mounts `binaries/licenses` read-only into ingestion. Verify the executable exists before startup:

```bash
test -x binaries/licenses/current/decoder
```

Start the stack:

```bash
cd deploy/docker
cp .env.example .env
# Edit .env and set your own SUPERSET_SECRET_KEY and SUPERSET_ADMIN_PASSWORD.
docker compose up -d --build
docker compose ps
```

This starts Go ingestion, Redis, ClickHouse, and Apache Superset. Published ports bind to `127.0.0.1` by default:

- Ingestion: `127.0.0.1:8081`
- ClickHouse HTTP/native: `127.0.0.1:8123` / `127.0.0.1:9000`
- Redis: `127.0.0.1:6379`
- Superset: `127.0.0.1:8088`

## Connect a Sensors Data SDK

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

Open Superset at `http://127.0.0.1:8088`. Set `SUPERSET_ADMIN_PASSWORD` before startup; the stack does not define a customer password.

## Production requirements

The stack does not define Redis, ClickHouse, MySQL, or customer account passwords. Redis and ClickHouse start without passwords only when their variables are empty and their ports remain bound to `127.0.0.1`. Before production:

- Set `REDIS_PASSWORD`, `CLICKHOUSE_PASSWORD`, `SUPERSET_ADMIN_PASSWORD`, `SUPERSET_SECRET_KEY`, and a matching URL-encoded `CLICKHOUSE_SQLALCHEMY_URI`.
- Terminate TLS in a reverse proxy and expose only the required ingestion path.
- Keep database ports private, configure backups, and monitor decoder failures, ingestion latency, and ClickHouse disk usage.
- Pin and review container image versions according to your release policy.
- Validate SDK version, encrypted payload plugins, identity behavior, property types, and event timestamps before moving traffic.

See [Docker deployment](deploy/docker/README.md), [Sensors Data SDK integration](docs/sensors-sdk.md), [FAQ](FAQ.md), and [third-party notices](THIRD_PARTY_NOTICES.md).

## Development checks

```bash
go test ./...
go build -o sensors main.go
docker compose -f deploy/docker/docker-compose.yml config
```

Contributions are described in [CONTRIBUTING.md](CONTRIBUTING.md). Never commit customer data, credentials, decoder binaries, or license files.
