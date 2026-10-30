# Agent Note: CPU-scheduling collection constraints

Status: implemented

[English](2026-10-25-cpu-sched-collection-constraints.md) | 中文

## 问题

CPU 调度诊断需要在真实主机上观察内核侧信号。`sched_switch` 在每次上下文切换时触发——繁忙的 30+ 核主机上约 5×10⁵/s，重建运行队列还会再乘 1.5–2 倍——因此把每条事件送进用户态无法满足很小的 CPU 与内存占用。目标内核版本也参差不齐；低于某一版本后，探针集合与 BTF 路径会散掉。

## 决策

- **基线内核：5.8**（ring buffer、fentry/BTF tracing、sysfs BTF 同时具备）。**地板：4.9**（perf-event 附着 + perf event array，无 fentry，可能需要 BTFHub）。低于 4.9 不支持。这次附着 `tracepoint/sched/sched_switch`，跑在 6.6 的测试客户机上。
- agent 在内核内聚合：tid → 在 CPU 上的纳秒的哈希表，每个间隔排空成 `cpu.TaskRuntime`。用户态看不到每条 `sched_switch`。协议预算为每主机每秒 10⁵ 稳态 / 10⁶ 突发事件。
- steal time 没有 tracepoint：读 procfs `/proc/stat`，或 kprobe `steal_account_process_time`（仅客户机侧语义）。尚未采集。
- 经 `sched_switch` 配对得到的 off-CPU 可以采样（已有 1/1000 的先例）以控制开销。尚未采集。

**本次不用 CO-RE。** 程序只读 tracepoint 参数——由 tracepoint 格式定下的原始上下文——所以不带任何 CO-RE 重定位，也不依赖内核 BTF。要读 `task_struct` 字段（`prev_state`、`comm`、`prio`）或换架构，才切到 `tp_btf/sched_switch` 加 `vmlinux.h`；那时 CO-RE 才开始付钱。

字节码由 [`sched.bpf.c`](../../../../internal/agent/domain/cpu/bpf/sched.bpf.c) 编出并入库：`cpu_bpfel.o` 与 `cpu_bpfeb.o` 用 `go:embed` 嵌进去，所以没有 clang 也能编出 agent。在装有 clang 与 libbpf 头的地方跑 `go generate ./internal/agent/domain/cpu` 重新生成；保留 `-no-strip`，因为去符号那一步要 `llvm-strip`。

在排空时读身份带来两条局限，而且都是静默的：tid 在一个排空间隔内被复用，前一个任务的纳秒会记到后一个任务的 `starttime` 上；任务在排空前退出则被丢弃，不产生 Record。两者都无法从 Report 上看出来。

Record 上的 `Pid` / `Tid` / `Starttime` 在排空时从 `/proc/<tid>` 读。4.9–5.7 的地板附着路径还没做。附着需要 root 或 `CAP_BPF`。

## 考虑过的替代

- **把每条事件转发到用户态**：否决——体量约 5×10⁵/s，重建运行队列再加 1.5–2 倍，打爆占用预算。
- **地板低于 4.9**：否决——没有我们愿意维护的附着路径。
- **用 perf buffer 当基线**：否决——5.8 是 ring buffer、fentry/BTF tracing 与 sysfs BTF 第一次同时出现的内核；更老的内核走地板路径，不是默认。

## 后果

`internal/agent/domain/cpu` 加载 cilium/ebpf 字节码，发出 `TaskRuntime`。CPU Analyzer 按 `{pid,tid,starttime}` 累加 `OnCPUNs`，超过一秒就出一份 Report。`Stop` 在关探针前再排空一次，最后一个间隔不会丢。4.9 地板路径和 steal time 仍是以后的附着工作。
