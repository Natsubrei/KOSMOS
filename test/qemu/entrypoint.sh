#!/usr/bin/env bash
# Boots one unprivileged QEMU guest for agent tests. Everything persistent
# lives in /state (bind-mounted from the host): the pinned image cache, the
# runner's SSH key, and the per-boot overlay, seed, and serial log. The guest
# is reachable only through user-mode networking: hostfwd for SSH on the
# container's port 2222 (published to the host loopback by compose) and the
# QEMU host alias 10.0.2.2:8080 for the server relay.
set -euo pipefail

state=${STATE_DIR:-/state}
image_name=${KOSMOS_GUEST_IMAGE:-openEuler}
cpus=${KOSMOS_GUEST_CPUS:-2}
memory_mb=${KOSMOS_GUEST_MEMORY_MB:-2048}
ssh_port=${KOSMOS_GUEST_SSH_PORT:-2222}
server_addr=${KOSMOS_SERVER_ADDR:-server:8080}
accel=${KOSMOS_QEMU_ACCEL:-kvm}

key=$state/ssh/id_ed25519
if [[ ! -f $key.pub ]]; then
  echo "entrypoint: $key.pub is missing; generate the runner key on the host (test/qemu/up.sh)" >&2
  exit 1
fi

# shellcheck source=/dev/null
. /etc/kosmos/images.env
case $image_name in
  openEuler) url=$OPENEULER_URL; sha=$OPENEULER_SHA256; disk=$OPENEULER_DISK ;;
  ubuntu) url=$UBUNTU_URL; sha=$UBUNTU_SHA256; disk=$UBUNTU_DISK ;;
  *) echo "entrypoint: unknown guest image $image_name (openEuler|ubuntu)" >&2; exit 1 ;;
esac

mkdir -p "$state/cache" "$state/run"
disk_path=$state/cache/$disk
if [[ ! -f $disk_path ]]; then
  fetch=$disk_path
  [[ $url == *.xz ]] && fetch=$disk_path.xz
  echo "entrypoint: fetching $image_name image to $fetch"
  curl -fL --retry 3 --retry-delay 2 -o "$fetch.part" "$url"
  echo "$sha  $fetch.part" | sha256sum -c -
  mv "$fetch.part" "$fetch"
  if [[ $fetch == *.xz ]]; then
    xz -d "$fetch"
  fi
fi

rm -f "$state"/run/disk.qcow2 "$state"/run/seed.iso "$state"/run/serial.log "$state"/run/facts
qemu-img create -q -f qcow2 -F qcow2 -b "$disk_path" "$state/run/disk.qcow2"
sed "s@__SSH_AUTHORIZED_KEY__@$(cat "$key.pub")@" /etc/kosmos/user-data > "$state/run/user-data"
cloud-localds "$state/run/seed.iso" "$state/run/user-data" /etc/kosmos/meta-data

guest_ssh() {
  ssh -i "$key" -p "$ssh_port" \
      -o BatchMode=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
      -o ConnectTimeout=3 -o LogLevel=ERROR kosmos@127.0.0.1 "$@"
}

cpu=host
[[ $accel == tcg ]] && cpu=max

started=$SECONDS
qemu-system-x86_64 \
  -name kosmos-guest \
  -machine q35 \
  -accel "$accel" \
  -cpu "$cpu" \
  -smp "$cpus" \
  -m "$memory_mb" \
  -object rng-random,filename=/dev/urandom,id=rng0 \
  -device virtio-rng-pci,rng=rng0 \
  -drive if=virtio,file="$state/run/disk.qcow2",format=qcow2 \
  -drive if=virtio,file="$state/run/seed.iso",format=raw,readonly=on \
  -netdev user,id=net0,hostfwd=tcp:0.0.0.0:"$ssh_port"-:22 \
  -device virtio-net-pci,netdev=net0 \
  -display none \
  -monitor none \
  -serial file:"$state/run/serial.log" &
qemu_pid=$!

relay_pid=""
cleanup() {
  [[ -n $relay_pid ]] && kill "$relay_pid" 2>/dev/null || true
  kill "$qemu_pid" 2>/dev/null || true
  wait "$qemu_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "entrypoint: waiting for $image_name guest to answer SSH"
ready=""
for _ in $(seq 1 180); do
  if guest_ssh true 2>/dev/null; then ready=1; break; fi
  if ! kill -0 "$qemu_pid" 2>/dev/null; then
    echo "entrypoint: QEMU exited before the guest booted; serial log follows" >&2
    cat "$state/run/serial.log" >&2
    exit 1
  fi
  sleep 1
done
if [[ -z $ready ]]; then
  echo "entrypoint: guest did not answer SSH within 180s; serial log tail follows" >&2
  tail -40 "$state/run/serial.log" >&2
  exit 1
fi

facts=$state/run/facts
printf 'image=%s\n' "$image_name" > "$facts"
printf 'boot_seconds=%s\n' "$((SECONDS - started))" >> "$facts"
printf 'kernel=%s\n' "$(guest_ssh uname -r)" >> "$facts"
if guest_ssh 'test -s /sys/kernel/btf/vmlinux'; then
  printf 'btf=present\n' >> "$facts"
else
  printf 'btf=missing\n' >> "$facts"
fi
# runcmd runs after sshd starts, so the marker can lag SSH readiness by seconds.
cloud_init=no-user-data
for _ in $(seq 1 30); do
  if guest_ssh 'test -f /etc/kosmos-provisioned'; then cloud_init=ok; break; fi
  sleep 1
done
printf 'cloud_init=%s\n' "$cloud_init" >> "$facts"
printf 'qemu_rss_kib=%s\n' \
  "$(awk '/^VmRSS/{print $2}' "/proc/$qemu_pid/status")" >> "$facts"
printf 'guest_mem_total_mib=%s\n' \
  "$(guest_ssh "awk '/MemTotal/{print int(\$2/1024)}' /proc/meminfo")" >> "$facts"

cat "$facts"
if grep -qx 'btf=missing' "$facts"; then
  echo "entrypoint: warning: guest kernel exposes no BTF; CO-RE probes will not load" >&2
fi
if grep -qx 'cloud_init=no-user-data' "$facts"; then
  echo "entrypoint: warning: cloud-init did not apply the seed; the kosmos user may not exist" >&2
fi

echo "entrypoint: relaying guest :8080 (QEMU host alias 10.0.2.2) to $server_addr"
socat TCP-LISTEN:8080,fork,reuseaddr,bind=127.0.0.1 TCP:"$server_addr" &
relay_pid=$!

wait "$qemu_pid"
