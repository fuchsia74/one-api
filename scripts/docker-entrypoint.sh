#!/usr/bin/env bash
set -euo pipefail

# Runtime permission fix for bind-mounted /data.
# If /data is a bind mount with root-owned content, chown it (best effort) before dropping privileges.
USER_NAME=oneapi
USER_ID=${ONEAPI_UID:-10001}
GROUP_ID=${ONEAPI_GID:-10001}
APP_BIN=/one-api
DATA_DIR=/data
LOG_DIR="$DATA_DIR/logs"

# Only chown if owned by root (avoid expensive recursive chown every start)
if [ "$(stat -c %u "%DATA_DIR%" || echo 0)" = "0" ]; then
  echo "Adjusting ownership of $DATA_DIR to $USER_NAME ($USER_ID:$GROUP_ID)" >&2 || true
  chown -R "$USER_ID:$GROUP_ID" "$DATA_DIR" || echo "Warning: could not chown $DATA_DIR" >&2
fi

mkdir -p "$LOG_DIR" || true
chown "$USER_ID:$GROUP_ID" "$LOG_DIR" || true

# Drop privileges using gosu
if [ "$(id -u)" = "0" ]; then
  exec gosu "$USER_NAME" "$APP_BIN" "$@"
else
  exec "$APP_BIN" "$@"
fi
