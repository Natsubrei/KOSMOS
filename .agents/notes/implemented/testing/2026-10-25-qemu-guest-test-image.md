# Agent Note: QEMU guest image for in-kernel agent tests

Status: implemented

English | [中文](2026-10-25-qemu-guest-test-image.zh.md)

## Problem

eBPF probes must run against a real guest kernel without giving the agent
host-level privileges. The test harness is QEMU inside a container.

## Decision

The test guest is **openEuler 24.03 LTS** (kernel 6.6), selected 2026-10-25.
**Ubuntu 24.04 LTS** (kernel 6.8) is the documented fallback. QEMU runs
unprivileged: `devices: [/dev/kvm]` plus `group_add: [<host kvm GID>]`;
user-mode networking; hostfwd for SSH. Boot budget is 10–30 s to SSH-ready,
healthcheck-gated. BTF is checked on every boot and reported in the boot facts.

## The provisioned stack

`test/qemu/` holds the environment; its README is the usage surface.

- `compose.yaml` runs two services: `qemu-guest` (QEMU inside a Debian
  container, `/dev/kvm` and nothing else) and `server` (this repository's
  `cmd/server` binary, published on the host loopback for inspection).
- `images.env` pins each image URL and the publisher's checksum; `entrypoint.sh`
  verifies the checksum before the first boot, caches the qcow2, and boots a
  fresh copy-on-write overlay on every start.
- A NoCloud seed built with `cloud-localds` injects the runner's SSH key and
  creates the `kosmos` user; `up.sh` generates that key on first run.
- The agent is delivered by scp over hostfwd (`push-agent.sh`); the image
  carries no agent payload.
- Boot facts land in `.runbin/qemu/run/facts`: image, seconds to SSH, kernel,
  BTF, cloud-init marker, QEMU RSS, guest memory.

## Boot, cost, and access path

Measured 2026-10-26 on a WSL2 host with `/dev/kvm`:

- openEuler: 13 s to SSH-ready, kernel `6.6.0-28.0.0.34.oe2403.x86_64`,
  `/sys/kernel/btf/vmlinux` present (about 5 MB).
- Ubuntu fallback: 12 s to SSH-ready, kernel `6.8.0-147-generic`, BTF present.
- Idle at the default 2 vCPU / 2048 MiB: guest MemTotal 1713 MiB, guest used about
  220 MiB, QEMU RSS about 600–640 MiB, container CPU under 1 % after boot. openEuler's
  kernel cmdline reserves `crashkernel=256M`, so the guest sees about 340 MiB less
  than configured; at 1024 MiB configured only 707 MiB is visible, which is why the
  default is 2048.
- NoCloud seeds are honored on both images: the injected key authenticates and
  the `runcmd` marker exists.
- Host to guest: `ssh -p 2222 kosmos@127.0.0.1` with
  `.runbin/qemu/ssh/id_ed25519`.
- Guest to server: the agent posts to `http://10.0.2.2:8080`; the guest-host
  container relays that to `server:8080`. The guest has no route to the Compose
  network, and nothing on its path holds host privileges.
- `test/qemu/smoke.sh` proves the chain end to end in about 20 s from a cold
  start.

## Alternatives considered

- **Ubuntu 24.04 as the primary image** (kernel 6.8, BTF present, about 600 MB):
  rejected as primary — the project selected openEuler 24.03 LTS; Ubuntu stays
  the fallback.
- **Privileged containers**: rejected — unprivileged QEMU with `/dev/kvm` and
  the host kvm GID is viable; GID mismatch between host and image is the known
  pitfall.
- **Host-level bpf without a guest**: rejected — probes run against a real
  guest kernel with no host-level privileges.
- **Baking the agent into the seed**: rejected — every agent rebuild would
  rebuild the seed and reboot the guest; scp over hostfwd keeps delivery out of
  the boot path.
- **A shared 9p or virtiofs directory for delivery**: rejected — it adds a
  guest kernel-module dependency and a mount to debug, while sshd is already
  required for readiness.
- **Guest attached to the Compose network** (host networking or a bridge):
  rejected — user-mode networking plus the `socat` relay keeps `/dev/kvm` as
  the only privilege the container holds.
- **QEMU `guestfwd` instead of the `socat` relay**: rejected — it needs a
  synthetic guest address (`10.0.2.100`) instead of the host alias `10.0.2.2`
  that the agent default expects.

## Consequences

- The idle floor is known: budget 2 vCPU, 2048 MiB configured (1713 MiB
  guest-visible), and about 620 MiB of host RSS per idle guest; measure again
  before running a pool in parallel.
- Agent tests assume a 6.6 CO-RE guest with BTF and that the container can open
  `/dev/kvm`. Without KVM the stack fails at boot; `KOSMOS_QEMU_ACCEL=tcg` is an
  untested escape hatch, not the supported path.
- Compose delivers a published port to the container interface, so hostfwd must
  bind `0.0.0.0` inside the container; a loopback-only hostfwd resets host
  connections.
- A proxy bound to the host loopback is unreachable from the default Docker
  build network; those hosts set `KOSMOS_BUILD_NETWORK=host`.
