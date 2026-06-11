#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="quay.io/wkulhanek/bmw-loxone-bridge:latest"
CONTAINER="bmw-loxone-bridge"

podman stop "$CONTAINER" 2>/dev/null || true
podman rm "$CONTAINER" 2>/dev/null || true

podman run -d \
  --name "$CONTAINER" \
  --restart=always \
  --env-file "${SCRIPT_DIR}/.env" \
  -p 8400:8400 \
  -v "${SCRIPT_DIR}/data:/data:Z" \
  "$IMAGE"

echo "Started ${CONTAINER} — check logs with: podman logs -f ${CONTAINER}"
