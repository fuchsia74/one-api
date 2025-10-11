#!/usr/bin/env bash
set -euo pipefail

# Runtime permission fix for bind-mounted /data.
# If /data is a bind mount with root-owned content, chown it (best effort) before dropping privileges.
USER_NAME=oneapi
USER_ID=${ONEAPI_UID:-10001}
GROUP_ID=${ONEAPI_GID:-10001}
APP_BIN=/one-api
DATA_DIR=/data
DEFAULT_LOG_DIR="$DATA_DIR/logs"

# Parse CLI arguments to discover custom log directory so we can prepare it
cli_args=("$@")
resolved_log_dir=""
i=0
while [ $i -lt ${#cli_args[@]} ]; do
  arg="${cli_args[$i]}"
  case "$arg" in
    --log-dir)
      next_index=$((i + 1))
      if [ $next_index -lt ${#cli_args[@]} ]; then
        resolved_log_dir="${cli_args[$next_index]}"
      fi
      ;;
    --log-dir=*)
      resolved_log_dir="${arg#--log-dir=}"
      ;;
  esac
  i=$((i + 1))
done

[ -n "$resolved_log_dir" ] || resolved_log_dir="$DEFAULT_LOG_DIR"

# Only chown if owned by root (avoid expensive recursive chown every start)
if [ -d "$DATA_DIR" ]; then
  data_dir_owner=$(stat -c %u "$DATA_DIR" 2>/dev/null || echo -1)
  if [ "$data_dir_owner" = "0" ]; then
    echo "Adjusting ownership of $DATA_DIR to $USER_NAME ($USER_ID:$GROUP_ID)" >&2 || true
    chown -R "$USER_ID:$GROUP_ID" "$DATA_DIR" || echo "Warning: could not chown $DATA_DIR" >&2
  fi
else
  mkdir -p "$DATA_DIR"
  chown "$USER_ID:$GROUP_ID" "$DATA_DIR" || true
fi

mkdir -p "$resolved_log_dir" || true
chown "$USER_ID:$GROUP_ID" "$resolved_log_dir" || true

# Drop privileges using gosu
if [ "$(id -u)" = "0" ]; then
  exec gosu "$USER_NAME" "$APP_BIN" "$@"
else
  exec "$APP_BIN" "$@"
fi
