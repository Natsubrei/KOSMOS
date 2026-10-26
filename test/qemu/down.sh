#!/usr/bin/env bash
# Stops the stack and removes the container; the cached image and runner key
# stay in .runbin/qemu so the next up.sh boots in seconds.
set -euo pipefail
cd "$(dirname "$0")"
docker compose down
