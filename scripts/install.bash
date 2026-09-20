#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/docker"
ENV_FILE="$COMPOSE_DIR/.env"

random_secret() {
  openssl rand -hex "${1:-24}"
}

prompt() {
  local message="$1" default_value="${2:-}" value
  if [[ -n "$default_value" ]]; then
    read -r -p "$message [$default_value]: " value
    printf '%s' "${value:-$default_value}"
  else
    read -r -p "$message: " value
    printf '%s' "$value"
  fi
}

prompt_secret() {
  local message="$1" value
  read -r -s -p "$message (留空表示无密码): " value
  printf '\n' >&2
  printf '%s' "$value"
}

port_open() {
  nc -z 127.0.0.1 "$1" >/dev/null 2>&1
}

yes_no() {
  local message="$1" answer
  read -r -p "$message [y/N]: " answer
  [[ "$answer" =~ ^[Yy]$ ]]
}

urlencode() {
  local input="$1" output="" character hex index
  for ((index = 0; index < ${#input}; index++)); do
    character="${input:index:1}"
    case "$character" in
      [a-zA-Z0-9.~_-]) output+="$character" ;;
      *) printf -v hex '%%%02X' "'$character"; output+="$hex" ;;
    esac
  done
  printf '%s' "$output"
}

dotenv_quote() {
  local value="${1//\\/\\\\}"
  value="${value//\'/\\\'}"
  printf "'%s'" "$value"
}

wait_for_service() {
  local service="$1" attempts=60 status
  while ((attempts > 0)); do
    status="$(docker compose ps --format json "$service" 2>/dev/null || true)"
    if [[ "$status" == *'"Health":"healthy"'* ]]; then
      return 0
    fi
    sleep 2
    ((attempts--))
  done
  echo "$service 未在预期时间内就绪，请运行 docker compose logs $service" >&2
  return 1
}

validate_external_redis() {
	local host_port="$1" password="$2" host port
	host="${host_port%:*}"
	port="${host_port##*:}"
  local args=(run --rm --add-host host.docker.internal:host-gateway redis:7-alpine redis-cli -h "$host" -p "$port")
  if [[ -n "$password" ]]; then
    args+=(-a "$password")
  fi
  args+=(ping)
  docker "${args[@]}" | grep -q PONG || { echo "无法连接已有 Redis，请检查地址和密码。" >&2; exit 1; }
}

initialize_external_clickhouse() {
	local host_port="$1" username="$2" password="$3" host port
	host="${host_port%:*}"
	port="${host_port##*:}"
  local args=(run --rm --add-host host.docker.internal:host-gateway
    -v "$COMPOSE_DIR/clickhouse/init_clickhouse.sql:/init_clickhouse.sql:ro"
    clickhouse/clickhouse-server:23.8 clickhouse-client -h "$host" --port "$port" --user "$username")
  if [[ -n "$password" ]]; then
    args+=(--password "$password")
  fi
  args+=(--multiquery --queries-file /init_clickhouse.sql)
  docker "${args[@]}" || { echo "无法初始化已有 ClickHouse，请检查地址、凭证和建库权限。" >&2; exit 1; }
}

command -v docker >/dev/null || { echo "请先安装 Docker。" >&2; exit 1; }
docker compose version >/dev/null || { echo "请先安装 Docker Compose v2。" >&2; exit 1; }
command -v openssl >/dev/null || { echo "缺少 openssl，无法安全生成密码。" >&2; exit 1; }
command -v nc >/dev/null || { echo "缺少 nc，无法检测已有服务。" >&2; exit 1; }

use_external_redis=false
if port_open 6379 && yes_no "检测到本机 Redis (6379)，是否复用"; then
  use_external_redis=true
fi

use_external_clickhouse=false
if { port_open 9000 || port_open 8123; } && yes_no "检测到本机 ClickHouse (9000/8123)，是否复用"; then
  use_external_clickhouse=true
fi

superset_secret="$(random_secret 32)"
superset_password="$(random_secret 16)"

if $use_external_redis; then
  redis_host="$(prompt 'Redis 地址（容器可访问地址）' 'host.docker.internal:6379')"
  redis_password="$(prompt_secret 'Redis 密码')"
else
  redis_host="redis:6379"
  redis_password="$(random_secret 16)"
fi

if $use_external_clickhouse; then
  clickhouse_host="$(prompt 'ClickHouse native 地址（容器可访问地址）' 'host.docker.internal:9000')"
  clickhouse_http_host="$(prompt 'ClickHouse HTTP 地址（不含协议）' 'host.docker.internal:8123')"
  clickhouse_user="$(prompt 'ClickHouse 用户名' 'default')"
  clickhouse_password="$(prompt_secret 'ClickHouse 密码')"
  clickhouse_db="$(prompt 'ClickHouse 数据库' 'sensors')"
else
  clickhouse_host="clickhouse:9000"
  clickhouse_http_host="clickhouse:8123"
  clickhouse_user="default"
  clickhouse_password="$(random_secret 16)"
  clickhouse_db="sensors"
fi

clickhouse_uri="clickhousedb://$(urlencode "$clickhouse_user"):$(urlencode "$clickhouse_password")@${clickhouse_http_host}/${clickhouse_db}"

compose_profiles=""
if ! $use_external_redis; then
  compose_profiles="redis-bundled"
fi
if ! $use_external_clickhouse; then
  if [[ -n "$compose_profiles" ]]; then
    compose_profiles+=","
  fi
  compose_profiles+="clickhouse-bundled"
fi

umask 077
cat >"$ENV_FILE" <<EOF
COMPOSE_PROFILES=$compose_profiles
REDIS_HOST=$(dotenv_quote "$redis_host")
REDIS_PASSWORD=$(dotenv_quote "$redis_password")
CLICKHOUSE_HOST=$(dotenv_quote "$clickhouse_host")
CLICKHOUSE_USER=$(dotenv_quote "$clickhouse_user")
CLICKHOUSE_PASSWORD=$(dotenv_quote "$clickhouse_password")
CLICKHOUSE_DB=$(dotenv_quote "$clickhouse_db")
CLICKHOUSE_SQLALCHEMY_URI=$(dotenv_quote "$clickhouse_uri")
SUPERSET_SECRET_KEY=$(dotenv_quote "$superset_secret")
SUPERSET_ADMIN_USERNAME=admin
SUPERSET_ADMIN_PASSWORD=$(dotenv_quote "$superset_password")
EOF

cd "$COMPOSE_DIR"

if $use_external_redis; then
  validate_external_redis "$redis_host" "$redis_password"
else
  docker compose up -d redis
  wait_for_service redis
fi
if $use_external_clickhouse; then
  initialize_external_clickhouse "$clickhouse_host" "$clickhouse_user" "$clickhouse_password"
else
  docker compose up -d clickhouse
  wait_for_service clickhouse
fi

clickhouse_client=(clickhouse-client --host "${clickhouse_host%:*}" --port "${clickhouse_host##*:}" --user "$clickhouse_user")
if [[ -n "$clickhouse_password" ]]; then
  clickhouse_client+=(--password "$clickhouse_password")
fi
if $use_external_clickhouse; then
  docker run --rm --add-host host.docker.internal:host-gateway \
    -v "$COMPOSE_DIR/clickhouse/demo_data.sql:/demo_data.sql:ro" \
    clickhouse/clickhouse-server:23.8 "${clickhouse_client[@]}" --multiquery --queries-file /demo_data.sql
else
  docker compose exec -T clickhouse "${clickhouse_client[@]}" --multiquery --queries-file /dev/stdin <clickhouse/demo_data.sql
fi

docker compose up -d --build superset
docker compose ps

cat <<EOF

SensorFlow 演示环境已启动（尚未启动真实埋点接收服务）。
Superset:  http://127.0.0.1:8088
演示看板: http://127.0.0.1:8088/superset/dashboard/sensorflow-events-overview/
Superset 用户名: admin
Superset 初始密码: $superset_password
私有配置已保存到 deploy/docker/.env（权限 600）。

下一步：先查看 Superset 演示数据和图表。
需要接入真实神策 SDK 数据时，请到 https://sensorflow.site/ 获取许可证，
下载许可证后运行：./activate.sh
EOF
