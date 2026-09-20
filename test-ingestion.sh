#!/bin/sh
set -eu

if ! command -v python3 >/dev/null 2>&1; then
  echo "测试脚本需要 Python 3。" >&2
  exit 1
fi

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec python3 "$root_dir/scripts/test_ingestion.py" "$@"
