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

install_decoder() {
  local decoder_target="$ROOT_DIR/binaries/licenses/customer/decoder"
  local license_dir="$ROOT_DIR/binaries/licenses/customer"
  local decoder_path="" verify_path="" candidate
  local candidates=()

  if [[ -x "$ROOT_DIR/binaries/licenses/current/decoder" ]]; then
    echo "检测到已安装 decoder。"
    return 0
  fi

  while IFS= read -r candidate; do
    candidates+=("$candidate")
  done < <(find "$ROOT_DIR" "$HOME/Downloads" -maxdepth 3 -type f \
    \( -name 'sensors-payload-decoder-*' -o -name 'sensors-payload-decoder' -o -name 'decoder' \) \
    ! -path '*/binaries/licenses/customer/decoder' 2>/dev/null | sort -u)

  if ((${#candidates[@]} == 1)); then
    decoder_path="${candidates[0]}"
    echo "检测到 decoder：$decoder_path"
  elif ((${#candidates[@]} > 1)); then
    echo "检测到多个 decoder 候选："
    printf '  %s\n' "${candidates[@]}"
  fi

  while [[ ! -f "$decoder_path" ]]; do
    decoder_path="$(prompt '请输入运营网站下载的 decoder 文件路径')"
    decoder_path="${decoder_path/#\~/$HOME}"
    if [[ ! -f "$decoder_path" ]]; then
      echo "文件不存在：$decoder_path" >&2
      decoder_path=""
    fi
  done

  mkdir -p "$license_dir"
  cp "$decoder_path" "$decoder_target"
  chmod 700 "$decoder_target"

  local decoder_source_dir
  decoder_source_dir="$(cd "$(dirname "$decoder_path")" && pwd)"
  while IFS= read -r candidate; do
    verify_path="$candidate"
    break
  done < <(find "$decoder_source_dir" -maxdepth 1 -type f -name '*.verify.json' 2>/dev/null | sort)

  if [[ -z "$verify_path" ]]; then
    verify_path="$(prompt '请输入运营网站下载的 verify.json 路径')"
    verify_path="${verify_path/#\~/$HOME}"
  fi
  [[ -f "$verify_path" ]] || { echo "验证文件不存在：$verify_path" >&2; exit 1; }
  cp "$verify_path" "$license_dir/decoder.verify.json"
  chmod 600 "$license_dir/decoder.verify.json"

  ln -sfn customer "$ROOT_DIR/binaries/licenses/current"
}

command -v docker >/dev/null || { echo "请先安装 Docker。" >&2; exit 1; }
docker compose version >/dev/null || { echo "请先安装 Docker Compose v2。" >&2; exit 1; }
[[ -f "$ENV_FILE" ]] || { echo "尚未完成基础环境安装，请先运行 ./install.sh。" >&2; exit 1; }

install_decoder

cd "$COMPOSE_DIR"
docker compose --profile ingestion up -d --build ingestion
docker compose ps ingestion

cat <<EOF

SensorFlow 真实埋点接收服务已启动。
Ingestion: http://127.0.0.1:8081
请将神策 SDK serverUrl 指向：https://你的域名/sensors/send/?token=YOUR_TOKEN
EOF
