#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:-quay.io/wkulhanek/bmw-loxone-bridge:latest}"

podman manifest rm "$IMAGE" 2>/dev/null || true
podman rmi "$IMAGE" 2>/dev/null || true
podman manifest create "$IMAGE"

podman build --platform linux/amd64,linux/arm64 --manifest "$IMAGE" .

echo "Multi-arch manifest created: ${IMAGE}"
echo "Push with: podman manifest push ${IMAGE}"
