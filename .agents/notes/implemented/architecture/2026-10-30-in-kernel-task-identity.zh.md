# Agent Note: In-kernel task identity via CO-RE

Status: implemented

[English](2026-10-30-in-kernel-task-identity.md) | 中文

## 问题

on-CPU 纳秒按 tid 在内核里聚合，身份（`Pid`、`Starttime`）却在 drain 时从 `/proc/<tid>` 读。一个间隔内 tid 被复用，前一个任务的纳秒会记到后一个上；drain 前退出的任务会消失。两者都是静默的。程序还手写了带 x86_64 pad 的 `sched_switch` 布局，所以不跨架构。

## 决策

用 `tp_btf/sched_switch` 经 BTF 附着。内核哈希的键是切换时从 `prev` 取的 `{tid, tgid, start_time}`：`pid` 是 tid，`tgid` 是线程组 id，`start_time` 是 `task_struct.start_time`（自开机起的纳秒）。用户态 drain 把这个键抄进 `cpu.TaskRuntime`，不打开 `/proc`。

CO-RE 的 C 就是 libbpf 的通常写法：`vmlinux.h`、`bpf_helpers.h`、`BPF_PROG`、`BPF_CORE_READ`。附着是 `link.AttachTracing`。`comm` 仍是注解，不是身份，也不采集；以后可以在 `task_val` 上加字段带上它，不必改键，也不必改 Analyzer。

本笔记拥有附着路径与身份契约。内核地板、内核内聚合预算、以及尚未采集的项仍归 [CPU 调度采集约束](2026-10-25-cpu-sched-collection-constraints.zh.md)。

## 考虑过的替代

- **drain 时继续读 `/proc`：** 否决——那就是竞态路径，忙主机上还是按 tid 付系统调用。
- **键仍是 tid，身份放在 value 里：** 否决——复用时纳秒仍加到同一槽。
- **把倾倒的 `vmlinux.h` 入库：** 否决——体积以兆计，而且文件里的布局是编译机的，不是 agent 加载时用的。
- **经典 `tracepoint/sched/sched_switch` 再加一条回退：** 否决——一个 probe 两条附着路径。4.9 地板仍是以后的路径，不是静默双路径。

## 后果

用户态的附着和 drain 只在 `internal/agent/domain/cpu/probe.go`。`contracts/cpu.TaskRuntime` 仍是四个字段；`Starttime` 是自开机起的纳秒。Analyzer 仍按 `{pid,tid,starttime}` 分组。

表容纳 16384 项、每项 24 字节——常驻约 390 KiB——而键现在包含一个 drain 间隔内出现过的每个不同任务启动，不再只是活着的 tid。某台主机在这个间隔内周转的任务超过这个数时，插入失败，该任务的纳秒被丢弃，而且是静默的：返回值没有检查，也没有任何地方统计丢弃。这与问题一节点名的属同一类缺陷，只是从普通的 tid 复用挪到了极端的周转量上。丢弃计数器是以后的变更。

drain 前退出的任务仍会被丢弃，因为它的表项没有消费方；键现在带的身份已经足够，留下它是选择而不是意外。

未在低于 6.6 客户机的内核上验证。程序固定了三参数的 `sched_switch` 原型，所以 5.8 基线是目标，不是已验证的地板。
