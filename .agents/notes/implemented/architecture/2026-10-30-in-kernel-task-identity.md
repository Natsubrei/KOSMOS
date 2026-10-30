# Agent Note: In-kernel task identity via CO-RE

Status: implemented

English | [中文](2026-10-30-in-kernel-task-identity.zh.md)

## Problem

On-CPU nanoseconds were aggregated in-kernel by tid, then identity
(`Pid`, `Starttime`) was read from `/proc/<tid>` at drain. A tid reused
inside one interval charged the predecessor's nanoseconds to the successor;
a task that exited before drain vanished. Both were silent. The program also
parsed a hand-copied `sched_switch` layout with an x86_64 pad, so it did not
cross architecture.

## Decision

`tp_btf/sched_switch` attaches through BTF. The in-kernel hash is keyed by
`{tid, tgid, start_time}` taken from `prev` at the switch: `pid` is the tid,
`tgid` the thread-group id, `start_time` is `task_struct.start_time` in nsec
since boot. Userspace drain copies that key into `cpu.TaskRuntime` and does
not open `/proc`.

CO-RE C is the libbpf usual form: `vmlinux.h`, `bpf_helpers.h`, `BPF_PROG`,
`BPF_CORE_READ`. Attach is `link.AttachTracing`. `comm` stays an annotation,
not identity, and is not collected; a later field on `task_val` can carry it
without changing the map key or the Analyzer.

This note owns the attach path and the identity contract. The kernel floor,
the in-kernel aggregation budget, and what is still uncollected stay with
[CPU-scheduling collection constraints](2026-10-25-cpu-sched-collection-constraints.md).

## Alternatives considered

- **Keep `/proc` at drain:** rejected — it is the racy path and the per-tid
  syscall cost on a busy host.
- **Key by tid, store identity in the value:** rejected — a reuse still
  adds nanoseconds onto the same slot.
- **Commit the dumped `vmlinux.h`:** rejected — megabytes, and the layout in
  that file is the build machine's, not what the agent loads.
- **Classic `tracepoint/sched/sched_switch` plus a fallback:** rejected —
  two attach paths in one probe. The 4.9 floor remains a later path, not a
  silent dual.

## Consequences

`internal/agent/domain/cpu/probe.go` is the only userspace attach and drain
site. `contracts/cpu.TaskRuntime` keeps the same four fields; `Starttime` is
nsec since boot. The Analyzer still groups by `{pid,tid,starttime}`.

The map holds 16384 entries of 24 bytes — about 390 KiB resident — and keys
now include every distinct task start seen in a drain interval, not just live
tids. On a host churning through more tasks than that in one interval the
insert fails and that task's nanoseconds are dropped, silently: the return
value is not checked and nothing counts drops. This is the same class of
defect the Problem names, bounded to absurd churn rather than to ordinary tid
reuse. A drop counter is a later change.

A task that exits before drain is still dropped, because its slot has no
record consumer; the key now carries enough identity that keeping it would be
a choice, not an accident.

Untested on kernels below the 6.6 guest. The program fixes the three-argument
`sched_switch` prototype, so the 5.8 baseline is a target, not a verified
floor.
