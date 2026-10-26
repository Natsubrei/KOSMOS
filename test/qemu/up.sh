#!/usr/bin/env bash
# Builds the test binaries and starts the guest and server. Returns once the
# guest answers SSH (the compose healthcheck) and prints the boot facts.
set -euo pipefail
cd "$(dirname "$0")"
. ./env.sh

mkdir -p "$state/ssh" "$state/bin" "$state/cache"
if [[ ! -f $ssh_key ]]; then
  ssh-keygen -q -t ed25519 -N '' -C kosmos-qemu -f "$ssh_key"
fi

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -o "$state/bin/kosmos-server" "$repo_root/cmd/server"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -o "$state/bin/kosmos-agent" "$repo_root/cmd/agent"

# Compose passes the host kvm GID through: the container only needs it when it
# drops root, but a mismatch here is the classic silent /dev/kvm denial.
KOSMOS_KVM_GID=$(stat -c %g /dev/kvm) docker compose up -d --build --wait

# The healthcheck turns green when SSH answers; the entrypoint writes the boot
# facts a moment later, so wait for the last line before printing them.
for _ in $(seq 1 60); do
  grep -q '^guest_mem_total_mib=' "$state/run/facts" 2>/dev/null && break
  sleep 1
done
cat "$state/run/facts"
