# Agent Note: CPU-scheduling collection constraints

Status: implemented

[English](2026-10-25-cpu-sched-collection-constraints.md) | 中文

## 问题

CPU 调度诊断需要在真实主机上观察内核侧信号。`sched_switch` 在每次上下文切换时触发——繁忙的 30+ 核主机上约 5×10⁵/s，重建运行队列还会再乘 1.5–2 倍——因此把每条事件送进用户态无法满足很小的 CPU 与内存占用。目标内核版本也参差不齐；低于某一版本后，探针集合与 BTF 路径会散掉。

## 决策

- **基线内核：5.8**（ring buffer、fentry/BTF tracing、sysfs BTF 同时具备）。**地板：4.9**（perf-event 附着 + perf event array，无 fentry，可能需要 BTFHub）。低于 4.9 不支持。验证镜像是 6.6 的测试客户机。
- agent 在内核内聚合，每个间隔 drain 成 `cpu.TaskRuntime`。用户态看不到每条 `sched_switch`。协议预算为每主机每秒 10⁵ 稳态 / 10⁶ 突发事件。
- steal time 没有 tracepoint：读 procfs `/proc/stat`，或 kprobe `steal_account_process_time`（仅客户机侧语义）。尚未采集。
- 经 `sched_switch` 配对得到的 off-CPU 可以采样（已有 1/1000 的先例）以控制开销。尚未采集。

本笔记拥有内核地板、内核内聚合预算、以及尚未采集的项。附着路径与身份契约归 [内核内的任务身份（CO-RE）](2026-10-30-in-kernel-task-identity.zh.md)。

字节码由 [`sched.bpf.c`](../../../../internal/agent/domain/cpu/bpf/sched.bpf.c) 编出并入库：`cpu_bpfel.o` 与 `cpu_bpfeb.o` 用 `go:embed` 嵌进去，所以没有 clang 或 `bpftool` 也能编出 agent。怎么重新生成见 [README.md](../../../../internal/agent/domain/cpu/bpf/README.md)。

4.9–5.7 的地板附着路径还没做。附着需要 root 或 `CAP_BPF`。

## 考虑过的替代

- **把每条事件转发到用户态**：否决——体量约 5×10⁵/s，重建运行队列再加 1.5–2 倍，打爆占用预算。
- **地板低于 4.9**：否决——没有我们愿意维护的附着路径。
- **用 perf buffer 当基线**：否决——5.8 是 ring buffer、fentry/BTF tracing 与 sysfs BTF 第一次同时出现的内核；更老的内核走地板路径，不是默认。

## 后果

`internal/agent/domain/cpu` 加载 cilium/ebpf 字节码，发出 `TaskRuntime`。CPU Analyzer 按 `{pid,tid,starttime}` 累加 `OnCPUNs`，超过 1 秒就出一份 Report——界限是该 Analyzer 上的字段（[Analyzer 的 Go 推导 Conclusion](2026-10-30-conclusions-in-analyzer-go.zh.md)）。`Stop` 在关探针前再排空一次，最后一个间隔不会丢。4.9 地板路径和 steal time 仍是以后的附着工作。
