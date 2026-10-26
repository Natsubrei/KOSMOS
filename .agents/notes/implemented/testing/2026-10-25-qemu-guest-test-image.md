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
healthcheck-gated. BTF is verified per image build.

## Alternatives considered

- **Ubuntu 24.04 as the primary image** (kernel 6.8, BTF present, about 600 MB):
  rejected as primary — the project selected openEuler 24.03 LTS; Ubuntu stays
  the fallback.
- **Privileged containers**: rejected — unprivileged QEMU with `/dev/kvm` and
  the host kvm GID is viable; GID mismatch between host and image is the known
  pitfall.
- **Host-level bpf without a guest**: rejected — probes run against a real
  guest kernel with no host-level privileges.

## Consequences

Provisioning measures the idle resource floor on this image. Agent tests
assume a 6.6 CO-RE guest and that `/dev/kvm` is available to the container.
