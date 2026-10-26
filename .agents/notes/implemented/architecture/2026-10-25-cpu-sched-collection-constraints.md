# Agent Note: CPU-scheduling collection constraints

Status: implemented

English | [中文](2026-10-25-cpu-sched-collection-constraints.zh.md)

## Problem

CPU-scheduling diagnosis needs in-kernel observables on real hosts.
`sched_switch` fires per context switch — about 5×10⁵/s on a busy 30+-core
host, more with run-queue reconstruction — so a per-event userspace path cannot
meet a small CPU and memory footprint. Target kernels also vary; below some
version the probe set and BTF story fall apart.

## Decision

- **Baseline kernel: 5.8** (ring buffer + fentry/BTF tracing + sysfs BTF
  together). **Floor: 4.9** (perf-event attach + perf event array, no fentry,
  may need BTFHub). Below 4.9 is unsupported.
- The agent aggregates in-kernel (histograms / threshold filters). The protocol
  budgets 10⁵ steady / 10⁶ burst events/s/host.
- Steal time has no tracepoint: read procfs `/proc/stat` or kprobe
  `steal_account_process_time` (guest-side semantics only).
- Off-CPU via `sched_switch` pairing may sample (1/1000 has precedent) to hold
  overhead.

## Alternatives considered

- **Forward every event to userspace**: rejected — volume is about 5×10⁵/s plus
  1.5–2× for run-queue reconstruction; it blows the footprint.
- **Floor below 4.9**: rejected — no viable attach path we will maintain.
- **Perf buffer as the baseline**: rejected — 5.8 is the first kernel where
  ring buffer, fentry/BTF tracing, and sysfs BTF coincide; older kernels use
  the floor path, they are not the default.

## Consequences

Probes and the protocol must assume reduced records, not raw tracepoints.
Hosts on 4.9–5.7 run a degraded attach path. Steal time is not a
tracepoint-shaped record.
