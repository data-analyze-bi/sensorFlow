# Helpers for installer-owned, single-line Compose dotenv settings. Never eval/source .env.
dotenv_has() {
  grep -q "^$1=" "$ENV_FILE"
}

dotenv_get() {
  local value
  value="$(sed -n "s/^$1=//p" "$ENV_FILE" | tail -n 1)"
  value="${value%$'\r'}"
  if [[ "$value" == \'*\' ]]; then
    value="${value:1:${#value}-2}"
    value="${value//\\\'/\'}"
    value="${value//\\\\/\\}"
  elif [[ "$value" == \"*\" ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s' "$value"
}

dotenv_default() {
  local key="$1" value="$2" temporary
  [[ -z "$(dotenv_get "$key")" ]] || return 0
  if dotenv_has "$key" && [[ -z "$value" ]]; then return 0; fi
  temporary="$(mktemp "${ENV_FILE}.XXXXXX")"
  chmod 600 "$temporary"
  # Remove empty definitions so there is exactly one effective value.
  sed "/^${key}=/d" "$ENV_FILE" > "$temporary"
  printf '\n%s=%s\n' "$key" "$(dotenv_quote "$value")" >> "$temporary"
  mv "$temporary" "$ENV_FILE"
}
