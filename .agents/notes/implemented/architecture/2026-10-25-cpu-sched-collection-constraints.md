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
  may need BTFHub). Below 4.9 is unsupported. This cut attaches the stable
  `tracepoint/sched/sched_switch` ABI and runs on the 6.6 test guest.
- The agent aggregates in-kernel: a hash of tid → on-CPU nanoseconds, drained
  once per interval into `cpu.TaskRuntime`. Userspace never sees one event per
  `sched_switch`. The protocol budgets 10⁵ steady / 10⁶ burst events/s/host.
- Steal time has no tracepoint: read procfs `/proc/stat` or kprobe
  `steal_account_process_time` (guest-side semantics only). Not collected yet.
- Off-CPU via `sched_switch` pairing may sample (1/1000 has precedent) to hold
  overhead. Not collected yet.

**No CO-RE in this cut.** The program reads tracepoint arguments only — the
raw context laid out by the tracepoint format — so it carries no CO-RE
relocation and does not depend on kernel BTF. Reading `task_struct` fields
(`prev_state`, `comm`, `prio`) or another architecture means switching to
`tp_btf/sched_switch` with `vmlinux.h`; that is where CO-RE starts to pay.

The bytecode is built from [`sched.bpf.c`](../../../../internal/agent/domain/cpu/bpf/sched.bpf.c)
and committed: `cpu_bpfel.o` and `cpu_bpfeb.o` are embedded with `go:embed`,
so a checkout builds the agent without clang. Regenerate with
`go generate ./internal/agent/domain/cpu` wherever clang and libbpf headers
are installed; keep `-no-strip` because the strip step needs `llvm-strip`.

Two limits follow from reading identity at drain time, and both are silent:
a tid reused inside one drain interval charges the predecessor's nanoseconds
to the successor's `starttime`, and a task that exits before the drain is
dropped rather than reported. Neither can be observed from the Report.

`Pid`/`Tid`/`Starttime` on the Record come from `/proc/<tid>` at drain. The
4.9–5.7 floor attach path is not built. Attaching needs root or `CAP_BPF`.

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
emits a Report when the sum exceeds one second. `Stop` drains once more before
closing the probe, so the last interval is not lost. The 4.9 floor path and
steal time remain future attach work.
