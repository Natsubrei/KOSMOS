#!/usr/bin/env bash
# Builds the agent for the guest and starts it against the relayed server. The
# delivery path is scp over hostfwd: the binary is not baked into the image, so
# rebuilding the agent does not rebuild or reboot anything.
set -euo pipefail
cd "$(dirname "$0")"
. ./env.sh

mkdir -p "$state/bin"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -o "$state/bin/kosmos-agent" "$repo_root/cmd/agent"

scp -i "$ssh_key" -P "$ssh_port" \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
    "$state/bin/kosmos-agent" kosmos@127.0.0.1:/tmp/kosmos-agent
guest_ssh 'sudo install -m 0755 /tmp/kosmos-agent /usr/local/bin/kosmos-agent'
guest_ssh 'sudo pkill -x kosmos-agent || true'
# The redirect must run under sudo, or the log open fails before the agent starts.
guest_ssh "sudo sh -c 'nohup /usr/local/bin/kosmos-agent -server http://10.0.2.2:8080 -interval 1s >>/var/log/kosmos-agent.log 2>&1 &'"
echo "push-agent: kosmos-agent started in the guest (log: /var/log/kosmos-agent.log)"
