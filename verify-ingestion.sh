#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/docker"
distinct_id="${1:-}"

[[ -n "$distinct_id" ]] || { echo "用法：./verify-ingestion.sh TEST_DISTINCT_ID" >&2; exit 2; }
[[ "$distinct_id" =~ ^[A-Za-z0-9_-]+$ ]] || { echo "测试用户 ID 格式无效。" >&2; exit 2; }

cd "$COMPOSE_DIR"
docker compose exec -T clickhouse clickhouse-client --query "
SELECT time, event, distinct_id, platform, environment
FROM sensors.event
WHERE distinct_id = '${distinct_id}'
ORDER BY time DESC
LIMIT 20
"
