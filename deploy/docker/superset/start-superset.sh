#!/usr/bin/env bash
set -euo pipefail

superset db upgrade

superset fab create-admin \
  --username "${SUPERSET_ADMIN_USERNAME:-admin}" \
  --firstname Admin \
  --lastname User \
  --email "${SUPERSET_ADMIN_EMAIL:-admin@example.com}" \
  --password "${SUPERSET_ADMIN_PASSWORD:?SUPERSET_ADMIN_PASSWORD is required}" || true

superset init
python /app/pythonpath/setup_clickhouse.py
python /app/pythonpath/create_sensorflow_dashboard.py

exec gunicorn \
  --bind 0.0.0.0:8088 \
  --workers 2 \
  --worker-class gthread \
  --threads 4 \
  --timeout 120 \
  "superset.app:create_app()"
