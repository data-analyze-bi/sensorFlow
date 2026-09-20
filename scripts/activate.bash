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

install_license() {
  local license_target="$ROOT_DIR/binaries/sensors-payload-decoder"
  local verify_target="$ROOT_DIR/binaries/sensors-payload-decoder.verify.json"
  local license_path="" verify_path="" candidate
  local candidates=()

  if [[ -x "$license_target" ]]; then
    echo "检测到已安装许可证。"
    return 0
  fi

  while IFS= read -r candidate; do
    candidates+=("$candidate")
  done < <(find "$ROOT_DIR" "$HOME/Downloads" -maxdepth 3 -type f \
    \( -name 'sensorflow-license-*' -o -name 'sensors-payload-decoder-*' -o -name 'sensors-payload-decoder' \) \
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
  cp "$license_path" "$license_target"
  chmod 700 "$license_target"

  local license_source_dir
  license_source_dir="$(cd "$(dirname "$license_path")" && pwd)"
  while IFS= read -r candidate; do
    verify_path="$candidate"
    break
  done < <(find "$license_source_dir" -maxdepth 1 -type f -name '*.verify.json' 2>/dev/null | sort)

  if [[ -z "$verify_path" ]]; then
    verify_path="$(prompt '请输入运营网站下载的 verify.json 路径')"
    verify_path="${verify_path/#\~/$HOME}"
  fi
  [[ -f "$verify_path" ]] || { echo "验证文件不存在：$verify_path" >&2; exit 1; }
  cp "$verify_path" "$verify_target"
  chmod 600 "$verify_target"
}

command -v docker >/dev/null || { echo "请先安装 Docker。" >&2; exit 1; }
docker compose version >/dev/null || { echo "请先安装 Docker Compose v2。" >&2; exit 1; }
[[ -f "$ENV_FILE" ]] || { echo "尚未完成基础环境安装，请先运行 ./install.sh。" >&2; exit 1; }

install_license

cd "$COMPOSE_DIR"
docker compose --profile ingestion up -d --build ingestion
docker compose ps ingestion

cat <<EOF

SensorFlow 真实埋点接收服务已启动。
Ingestion: http://127.0.0.1:8081
请将神策 SDK serverUrl 指向：https://你的域名/sensors/send/?token=YOUR_TOKEN
EOF
