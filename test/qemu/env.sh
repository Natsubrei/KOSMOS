# shellcheck shell=bash
# Shared paths and ports for the host-side helpers. Source it, do not run it.
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
state=$repo_root/.runbin/qemu
ssh_key=$state/ssh/id_ed25519
ssh_port=${KOSMOS_SSH_PORT:-2222}
server_port=${KOSMOS_SERVER_PORT:-8080}

# guest_ssh <args...> runs one command in the guest over the forwarded port.
guest_ssh() {
  ssh -i "$ssh_key" -p "$ssh_port" \
      -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      -o ConnectTimeout=3 -o LogLevel=ERROR kosmos@127.0.0.1 "$@"
}
