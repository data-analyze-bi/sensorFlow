#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/docker"
ENV_FILE="$COMPOSE_DIR/.env"

random_secret() {
  od -An -N "${1:-24}" -tx1 /dev/urandom | tr -d ' \n'
}

# SSH_CONNECTION contains client IP, client port, server IP, server port.
# Prefer the address used for this SSH session; no external IP lookup is needed.
detect_public_host() {
  local client client_port server server_port detected
  if [[ -n "${SSH_CONNECTION:-}" ]]; then
    read -r client client_port server server_port <<< "$SSH_CONNECTION"
    if [[ -n "$server" ]]; then printf '%s' "$server"; return; fi
  fi
  if command -v hostname >/dev/null 2>&1; then
    detected="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"
    if [[ -n "$detected" ]]; then printf '%s' "$detected"; return; fi
  fi
  printf '%s' localhost
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
  (exec 3<>"/dev/tcp/127.0.0.1/$1") >/dev/null 2>&1
}

yes_no() {
  local message="$1" answer
  [[ -t 0 ]] || return 1
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
  local service="$1" attempts=180 status
  while ((attempts > 0)); do
    status="$(docker compose ps --format json "$service" 2>/dev/null || true)"
    if [[ "$status" == *'"Health":"healthy"'* ]]; then
      return 0
    fi
    if [[ "$status" == *'"State":"exited"'* || "$status" == *'"Health":"unhealthy"'* ]]; then
      echo "$service 启动失败，请在 deploy/docker 运行 docker compose logs $service" >&2
      return 1
    fi
    sleep 2
    attempts=$((attempts - 1))
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
docker info >/dev/null 2>&1 || { echo "无法连接 Docker，请启动 Docker 并确认当前用户有访问权限。" >&2; exit 1; }
source "$ROOT_DIR/scripts/install-env.bash"
umask 077
touch "$ENV_FILE"
chmod 600 "$ENV_FILE"

# Existing configuration is data, never shell code. Preserve credentials on retries.
existing_profiles="$(dotenv_get COMPOSE_PROFILES)"


use_external_redis=false
if [[ -n "$(dotenv_get REDIS_HOST)" && "$(dotenv_get REDIS_HOST)" != redis:6379 ]]; then
  use_external_redis=true
elif [[ -z "$existing_profiles" ]] && port_open 6379 && yes_no "检测到本机 Redis (6379)，是否复用"; then
  use_external_redis=true
fi

use_external_clickhouse=false
if [[ -n "$(dotenv_get CLICKHOUSE_HOST)" && "$(dotenv_get CLICKHOUSE_HOST)" != clickhouse:9000 ]]; then
  use_external_clickhouse=true
elif [[ -z "$existing_profiles" ]] && { port_open 9000 || port_open 8123; } && yes_no "检测到本机 ClickHouse (9000/8123)，是否复用"; then
  use_external_clickhouse=true
fi

superset_secret="$(dotenv_get SUPERSET_SECRET_KEY)"
superset_secret="${superset_secret:-$(random_secret 32)}"
superset_password="$(dotenv_get SUPERSET_ADMIN_PASSWORD)"
superset_password="${superset_password:-$(random_secret 16)}"
ingestion_token="$(dotenv_get SENSORFLOW_INGESTION_TOKEN)"
ingestion_token="${ingestion_token:-$(random_secret 24)}"

if $use_external_redis && [[ -n "$(dotenv_get REDIS_HOST)" ]]; then
  redis_host="$(dotenv_get REDIS_HOST)"
  redis_password="$(dotenv_get REDIS_PASSWORD)"
elif $use_external_redis; then
  redis_host="$(prompt 'Redis 地址（容器可访问地址）' 'host.docker.internal:6379')"
  redis_password="$(prompt_secret 'Redis 密码')"
else
  redis_host="redis:6379"
  if dotenv_has REDIS_PASSWORD; then redis_password="$(dotenv_get REDIS_PASSWORD)"; else redis_password="$(random_secret 16)"; fi
fi

if $use_external_clickhouse && [[ -n "$(dotenv_get CLICKHOUSE_HOST)" ]]; then
  clickhouse_host="$(dotenv_get CLICKHOUSE_HOST)"
  clickhouse_http_host="$(dotenv_get CLICKHOUSE_HTTP_HOST)"
  clickhouse_http_host="${clickhouse_http_host:-${clickhouse_host%:*}:8123}"
  clickhouse_user="$(dotenv_get CLICKHOUSE_USER)"
  clickhouse_password="$(dotenv_get CLICKHOUSE_PASSWORD)"
  clickhouse_db="$(dotenv_get CLICKHOUSE_DB)"
  clickhouse_user="${clickhouse_user:-default}"
  clickhouse_db="${clickhouse_db:-sensors}"
elif $use_external_clickhouse; then
  clickhouse_host="$(prompt 'ClickHouse native 地址（容器可访问地址）' 'host.docker.internal:9000')"
  clickhouse_http_host="$(prompt 'ClickHouse HTTP 地址（不含协议）' 'host.docker.internal:8123')"
  clickhouse_user="$(prompt 'ClickHouse 用户名' 'default')"
  clickhouse_password="$(prompt_secret 'ClickHouse 密码')"
  clickhouse_db="$(prompt 'ClickHouse 数据库' 'sensors')"
else
  clickhouse_host="clickhouse:9000"
  clickhouse_http_host="clickhouse:8123"
  clickhouse_user="$(dotenv_get CLICKHOUSE_USER)"
  clickhouse_user="${clickhouse_user:-default}"
  if dotenv_has CLICKHOUSE_PASSWORD; then clickhouse_password="$(dotenv_get CLICKHOUSE_PASSWORD)"; else clickhouse_password="$(random_secret 16)"; fi
  clickhouse_db="$(dotenv_get CLICKHOUSE_DB)"
  clickhouse_db="${clickhouse_db:-sensors}"
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

# Fill only missing/empty required values; leave unrelated settings untouched.
dotenv_default COMPOSE_PROFILES "${existing_profiles:-$compose_profiles}"
dotenv_default REDIS_HOST "$redis_host"
dotenv_default REDIS_PASSWORD "$redis_password"
dotenv_default CLICKHOUSE_HOST "$clickhouse_host"
dotenv_default CLICKHOUSE_HTTP_HOST "$clickhouse_http_host"
dotenv_default CLICKHOUSE_USER "$clickhouse_user"
dotenv_default CLICKHOUSE_PASSWORD "$clickhouse_password"
dotenv_default CLICKHOUSE_DB "$clickhouse_db"
dotenv_default CLICKHOUSE_SQLALCHEMY_URI "$clickhouse_uri"
dotenv_default SUPERSET_SECRET_KEY "$superset_secret"
dotenv_default SUPERSET_ADMIN_USERNAME admin
dotenv_default SUPERSET_ADMIN_PASSWORD "$superset_password"
dotenv_default SENSORFLOW_INGESTION_TOKEN "$ingestion_token"

public_host="${SENSORFLOW_PUBLIC_HOST:-$(dotenv_get SENSORFLOW_PUBLIC_HOST)}"
public_host="${public_host:-$(detect_public_host)}"
public_port="$(dotenv_get SUPERSET_PORT)"
public_port="${public_port:-8088}"
url_host="$public_host"
if [[ "$url_host" == *:* && "$url_host" != \[*\] ]]; then url_host="[$url_host]"; fi
dotenv_default SENSORFLOW_PUBLIC_HOST "$public_host"
dotenv_default SUPERSET_BIND_ADDRESS 0.0.0.0
dotenv_default SUPERSET_PUBLIC_BASE_URL "http://${url_host}:${public_port}/"
public_url="$(dotenv_get SUPERSET_PUBLIC_BASE_URL)"
public_url="${public_url%/}"

cd "$COMPOSE_DIR"
# Use the private file consistently, including when shell variables are set.
unset SUPERSET_SECRET_KEY SUPERSET_ADMIN_PASSWORD SENSORFLOW_INGESTION_TOKEN
unset REDIS_HOST REDIS_PASSWORD CLICKHOUSE_HOST CLICKHOUSE_USER CLICKHOUSE_PASSWORD
unset CLICKHOUSE_DB CLICKHOUSE_SQLALCHEMY_URI COMPOSE_PROFILES SUPERSET_ADMIN_USERNAME
unset SUPERSET_PUBLIC_BASE_URL SUPERSET_PORT SUPERSET_BIND_ADDRESS
docker compose config --quiet

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
wait_for_service superset
docker compose ps

cat <<EOF

SensorFlow 演示环境已启动（尚未启动真实埋点接收服务）。
Superset:  $public_url
演示看板: $public_url/superset/dashboard/sensorflow-events-overview/
EOF
# Bold red credentials stand out in a terminal; NO_COLOR keeps plain-text logs clean.
if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
  printf '\033[1;31mSuperset 用户名: %s\nSuperset 密码: %s\033[0m\n' "$(dotenv_get SUPERSET_ADMIN_USERNAME)" "$superset_password"
else
  printf 'Superset 用户名: %s\nSuperset 密码: %s\n' "$(dotenv_get SUPERSET_ADMIN_USERNAME)" "$superset_password"
fi
cat <<EOF
私有配置已保存到 deploy/docker/.env（权限 600），重复安装会保留已有凭证。

请在浏览器打开上面的地址并登录查看图表。服务器/云防火墙需允许 TCP ${public_port}。
如服务器位于 NAT 后或显示的地址不正确，在 deploy/docker/.env 中设置
SENSORFLOW_PUBLIC_HOST 和 SUPERSET_PUBLIC_BASE_URL 为实际可访问的地址。
绑定域名与启用 HTTPS：参阅 deploy/docker/README.zh-CN.md「绑定域名」。

需要接入真实神策 SDK 数据时，请到 https://sensorflow.site/ 获取许可证，
下载许可证后运行：./activate.sh
EOF
