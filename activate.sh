#!/bin/sh
set -eu

if ! command -v bash >/dev/null 2>&1; then
  echo "激活程序需要 Bash，请先安装 Bash。" >&2
  exit 1
fi

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec bash "$root_dir/scripts/activate.bash" "$@"
