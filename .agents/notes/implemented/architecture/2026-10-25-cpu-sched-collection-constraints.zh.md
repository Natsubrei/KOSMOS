# Agent Note: CPU-scheduling collection constraints

Status: implemented

[English](2026-10-25-cpu-sched-collection-constraints.md) | 中文

## 问题

CPU 调度诊断需要在真实主机上观察内核侧信号。`sched_switch` 在每次上下文切换时触发——繁忙的 30+ 核主机上约 5×10⁵/s，重建运行队列还会再乘 1.5–2 倍——因此把每条事件送进用户态无法满足很小的 CPU 与内存占用。目标内核版本也参差不齐；低于某一版本后，探针集合与 BTF 路径会散掉。

## 决策

- **基线内核：5.8**（ring buffer、fentry/BTF tracing、sysfs BTF 同时具备）。**地板：4.9**（perf-event 附着 + perf event array，无 fentry，可能需要 BTFHub）。低于 4.9 不支持。
- agent 在内核内聚合（直方图 / 阈值过滤）。协议预算为每主机每秒 10⁵ 稳态 / 10⁶ 突发事件。
- steal time 没有 tracepoint：读 procfs `/proc/stat`，或 kprobe `steal_account_process_time`（仅客户机侧语义）。
- 经 `sched_switch` 配对得到的 off-CPU 可以采样（已有 1/1000 的先例）以控制开销。

## 考虑过的替代

- **把每条事件转发到用户态**：否决——体量约 5×10⁵/s，重建运行队列再加 1.5–2 倍，打爆占用预算。
- **地板低于 4.9**：否决——没有我们愿意维护的附着路径。
- **用 perf buffer 当基线**：否决——5.8 是 ring buffer、fentry/BTF tracing 与 sysfs BTF 第一次同时出现的内核；更老的内核走地板路径，不是默认。

## 后果

探针与协议必须按压缩后的 Record 来设计，而不是原始 tracepoint。4.9–5.7 的主机走降级附着路径。steal time 不是 tracepoint 形态的 Record。
