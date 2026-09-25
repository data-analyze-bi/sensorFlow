# Getting Started with SensorFlow

SensorFlow is an open-source, self-hosted product analytics platform. This guide installs the demo first, then activates real event ingestion.

## Prerequisites

- Docker Engine or Docker Desktop with Docker Compose v2
- A server IP reachable from your browser (TCP 8088 must be allowed)

## Install and view the demo

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
./install.sh
```

The installer generates missing credentials, prints the Superset credentials in red, and outputs a server-IP URL. Allow TCP 8088 and open that URL to view the demo dashboard. Caddy is optional and is only needed for a domain name and HTTPS.

## Activate real ingestion

After verifying the demo, run this in the repository root:

```bash
./activate.sh
```

Configure your SDK with the URL printed by the activation command, then send an `integration_test` event. See [Sensors Data SDK integration](sensors-sdk.md).
