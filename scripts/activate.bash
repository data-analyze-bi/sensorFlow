#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/docker"
ENV_FILE="$COMPOSE_DIR/.env"

prompt() {
  local message="$1" value
  read -r -p "$message: " value
  printf '%s' "$value"
}

dotenv_get() {
  local key="$1" value
  value="$(sed -n "s/^${key}=//p" "$ENV_FILE" | tail -n 1)"
  value="${value#\'}"
  value="${value%\'}"
  printf '%s' "$value"
}

dotenv_set() {
  local key="$1" value="$2" escaped
  escaped="${value//\'/\'\\\'\'}"
  if grep -q "^${key}=" "$ENV_FILE"; then
    sed -i.bak "s|^${key}=.*$|${key}='${escaped}'|" "$ENV_FILE"
    rm -f "$ENV_FILE.bak"
  else
    printf "%s='%s'\n" "$key" "$escaped" >>"$ENV_FILE"
  fi
}

install_license() {
  local license_target="$ROOT_DIR/binaries/sensors-payload-license"
  local verify_target="$ROOT_DIR/binaries/sensors-payload-license.verify.json"
  local license_path="" verify_path="" candidate archive_dir=""
  local candidates=()

  if [[ -x "$license_target" && -f "$verify_target" ]]; then
    echo "检测到已安装许可证。"
    return 0
  fi

  while IFS= read -r candidate; do
    candidates+=("$candidate")
  done < <(find "$ROOT_DIR" "$HOME/Downloads" -maxdepth 3 -type f \
    -name 'sensorflow-license-*' \
    ! -path "$license_target" 2>/dev/null | sort -u)

  if ((${#candidates[@]} == 1)); then
    license_path="${candidates[0]}"
    echo "检测到许可证：$license_path"
  elif ((${#candidates[@]} > 1)); then
    echo "检测到多个许可证候选："
    printf '  %s\n' "${candidates[@]}"
  fi

  while [[ ! -f "$license_path" ]]; do
    license_path="$(prompt '请输入运营网站下载的许可证文件路径')"
    license_path="${license_path/#\~/$HOME}"
    if [[ ! -f "$license_path" ]]; then
      echo "文件不存在：$license_path" >&2
      license_path=""
    fi
  done

  mkdir -p "$ROOT_DIR/binaries"

  if [[ "$license_path" == *.zip ]]; then
    command -v unzip >/dev/null || { echo "请先安装 unzip，或手动解压许可证压缩包。" >&2; exit 1; }
    archive_dir="$(mktemp -d)"
    trap '[[ -z "${archive_dir:-}" ]] || rm -rf "$archive_dir"' RETURN
    unzip -q "$license_path" -d "$archive_dir"
    license_path="$archive_dir/binaries/sensors-payload-license"
    verify_path="$archive_dir/binaries/sensors-payload-license.verify.json"
    [[ -f "$license_path" ]] || { echo "许可证压缩包缺少 binaries/sensors-payload-license。" >&2; exit 1; }
    [[ -f "$verify_path" ]] || { echo "许可证压缩包缺少 binaries/sensors-payload-license.verify.json。" >&2; exit 1; }
  fi

  if [[ ! "$license_path" -ef "$license_target" ]]; then
    cp "$license_path" "$license_target"
  fi
  chmod 700 "$license_target"

  if [[ -z "$verify_path" ]]; then
    local license_source_dir
    license_source_dir="$(cd "$(dirname "$license_path")" && pwd)"
    while IFS= read -r candidate; do
      verify_path="$candidate"
      break
    done < <(find "$license_source_dir" -maxdepth 1 -type f -name '*.verify.json' 2>/dev/null | sort)
  fi

  if [[ -z "$verify_path" ]]; then
    verify_path="$(prompt '请输入运营网站下载的 verify.json 路径')"
    verify_path="${verify_path/#\~/$HOME}"
  fi
  [[ -f "$verify_path" ]] || { echo "验证文件不存在：$verify_path" >&2; exit 1; }
  if [[ ! "$verify_path" -ef "$verify_target" ]]; then
    cp "$verify_path" "$verify_target"
  fi
  chmod 600 "$verify_target"

  if [[ -n "$archive_dir" ]]; then
    rm -rf "$archive_dir"
    archive_dir=""
    trap - RETURN
  fi
}

command -v docker >/dev/null || { echo "请先安装 Docker。" >&2; exit 1; }
docker compose version >/dev/null || { echo "请先安装 Docker Compose v2。" >&2; exit 1; }
[[ -f "$ENV_FILE" ]] || { echo "尚未完成基础环境安装，请先运行 ./install.sh。" >&2; exit 1; }

install_license

ingestion_token="$(dotenv_get SENSORFLOW_INGESTION_TOKEN)"
if [[ -z "$ingestion_token" ]]; then
  ingestion_token="$(openssl rand -hex 24)"
  dotenv_set SENSORFLOW_INGESTION_TOKEN "$ingestion_token"
fi

domain="$(dotenv_get SENSORFLOW_DOMAIN)"
if [[ -z "$domain" && -t 0 ]]; then
  domain="$(prompt '请输入已解析到本机公网 IP 的埋点域名（没有则直接回车，仅启用本机地址）')"
  domain="${domain#http://}"
  domain="${domain#https://}"
  domain="${domain%%/*}"
  [[ -z "$domain" ]] || dotenv_set SENSORFLOW_DOMAIN "$domain"
fi

cd "$COMPOSE_DIR"
docker compose --profile ingestion up -d --build ingestion
if [[ -n "$domain" ]]; then
  docker compose --profile ingestion --profile edge up -d caddy
fi
docker compose ps ingestion

if [[ -n "$domain" ]]; then
  server_url="https://${domain}/sensors/send/?token=${ingestion_token}"
else
  server_url="http://127.0.0.1:${SENSORFLOW_PORT:-8081}/sensors/send/?token=${ingestion_token}"
fi

cat <<EOF

SensorFlow 真实埋点接收服务已启动。
Ingestion: http://127.0.0.1:8081
埋点 Token: $ingestion_token
神策 SDK serverUrl: $server_url

Token 是安装器自动生成的接收密钥，已保存到 deploy/docker/.env，请勿提交到 Git。
$(if [[ -n "$domain" ]]; then printf '%s' "Caddy 已启动，将自动申请 HTTPS 证书。请确保 DNS A/AAAA 记录已指向本服务器，且 80/443 端口可访问。"; else printf '%s' "当前仅可本机访问。如需公网埋点，先将域名 DNS 指向服务器，再重新运行 ./activate.sh 并输入域名。"; fi)
EOF
