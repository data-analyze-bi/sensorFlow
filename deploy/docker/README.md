# One-command Docker installation

With Docker Engine/Desktop and Docker Compose available, run `./install.sh` from the repository root or from `deploy/docker`. No host Go, Python, nc, or openssl installation is needed. Raw `docker compose up` does not generate secrets; use the installer for first-time setup.

The installer fills missing secrets, saves a private `.env` (mode 600), starts Redis, ClickHouse and Superset, imports labeled demo data, and waits for Superset to become healthy. It prints the dashboard URL and highlights the administrator username and password in bold red in a terminal (`NO_COLOR=1` disables color). Reruns preserve existing credentials, custom settings and volumes.

## Access through the server IP

Superset listens on `0.0.0.0:8088` by default. The installer uses the SSH connection's server address for its links. Open `http://YOUR_SERVER_IP:8088` and log in to view the demo dashboard. Allow inbound TCP 8088 in the host/cloud firewall. Redis, ClickHouse and ingestion remain bound to loopback; port 8081 is the ingestion API, not the dashboard.

For NAT or a local installation, specify the reachable address on first installation:

```bash
SENSORFLOW_PUBLIC_HOST=YOUR_SERVER_IP ./install.sh
```

For existing installations, edit `SUPERSET_PUBLIC_BASE_URL`, `SUPERSET_BIND_ADDRESS` and `SUPERSET_PORT` in `deploy/docker/.env`, then rerun the installer. Set the bind address to `127.0.0.1` for local-only access.

## Domain and HTTPS

Use separate subdomains for the Superset UI and SensorFlow ingestion API. Point both DNS records to the server and allow TCP 80/443.

### 1. Superset domain

```dotenv
SUPERSET_DOMAIN=analytics.example.com
SUPERSET_PUBLIC_BASE_URL=https://analytics.example.com/
SUPERSET_BIND_ADDRESS=127.0.0.1
```

Open `https://analytics.example.com` and log in with the installer credentials.

### 2. SensorFlow ingestion domain

```dotenv
SENSORFLOW_DOMAIN=track.example.com
```

After running `./activate.sh`, configure SDKs with `https://track.example.com/sensors/send/?token=YOUR_TOKEN`. This hostname only proxies `/sensors/*` and never exposes Superset. Either hostname may be configured independently; if both are used, they must be different.

From `deploy/docker`, run:

```bash
docker compose up -d superset
docker compose --profile edge up -d caddy
docker compose logs --tail=50 caddy
```

Caddy automatically obtains and renews TLS certificates.

## Activate real ingestion

The demo does not need a license and does not start ingestion. Download your license from `sensorflow.site`, then run `./activate.sh` from the repository root. If configured, the ingestion domain is reused for `https://YOUR_DOMAIN/sensors/send/?token=YOUR_TOKEN`.

## Operations

From `deploy/docker`, use `docker compose ps` and `docker compose logs --tail=100 superset`. `docker compose down` stops services without deleting volumes. Back up `.env` with your database, preserve `SUPERSET_SECRET_KEY`, and never commit credentials or license files.

[中文文档](README.zh-CN.md)
