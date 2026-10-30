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
  may need BTFHub). Below 4.9 is unsupported. The proving image is the 6.6
  test guest.
- The agent aggregates in-kernel, drained once per interval into
  `cpu.TaskRuntime`. Userspace never sees one event per `sched_switch`. The
  protocol budgets 10⁵ steady / 10⁶ burst events/s/host.
- Steal time has no tracepoint: read procfs `/proc/stat` or kprobe
  `steal_account_process_time` (guest-side semantics only). Not collected yet.
- Off-CPU via `sched_switch` pairing may sample (1/1000 has precedent) to hold
  overhead. Not collected yet.

This note owns the kernel floor, the in-kernel aggregation budget, and what
is not collected yet. The attach path and the identity contract belong to
[In-kernel task identity via CO-RE](2026-10-30-in-kernel-task-identity.md).

The bytecode is built from [`sched.bpf.c`](../../../../internal/agent/domain/cpu/bpf/sched.bpf.c)
and committed: `cpu_bpfel.o` and `cpu_bpfeb.o` are embedded with `go:embed`,
so a checkout builds the agent without clang or `bpftool`. How to regenerate
is [linux.md](../../../../internal/agent/domain/cpu/bpf/README.md).

The 4.9–5.7 floor attach path is not built. Attaching needs root or `CAP_BPF`.

## Alternatives considered

- **Forward every event to userspace**: rejected — volume is about 5×10⁵/s plus
  1.5–2× for run-queue reconstruction; it blows the footprint.
- **Floor below 4.9**: rejected — no viable attach path we will maintain.
- **Perf buffer as the baseline**: rejected — 5.8 is the first kernel where
  ring buffer, fentry/BTF tracing, and sysfs BTF coincide; older kernels use
  the floor path, they are not the default.

## Consequences

`internal/agent/domain/cpu` loads cilium/ebpf bytecode and emits
`TaskRuntime`. The CPU Analyzer sums `OnCPUNs` per `{pid,tid,starttime}` and
produces a Report when the sum exceeds one second — the bound is a field on
that Analyzer
([Conclusions derived in Analyzer Go](2026-10-30-conclusions-in-analyzer-go.md)).
`Stop` drains once more before closing the probe, so the last interval is not
lost. The 4.9 floor path and steal time remain future attach work.
