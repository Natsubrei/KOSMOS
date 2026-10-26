#!/usr/bin/env bash
# Opens a shell in the guest, or runs the arguments as one guest command.
set -euo pipefail
cd "$(dirname "$0")"
. ./env.sh
exec ssh -i "$ssh_key" -p "$ssh_port" \
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
  kosmos@127.0.0.1 "$@"
