# Agent Note: Compile-time domain bundles over a zero-branch record bus

Status: implemented

[English](2026-10-25-compile-time-domain-bundles.md) | 中文

## 问题

KOSMOS 必须持续吸收新的采集目标。若记录总线为每个域写分支，每加一个域都要改总线。运行时插件加载会在目标机上长出工具链或装载器安全面，而 agent 必须保持单个低开销静态二进制，eBPF 的 C 代码反正要在树内编译。

## 决策

扩展单元是**域包**：一个自包含的编译期包，实现一个采集域——eBPF C 程序、map schema、解码器、Record 类型、可选的富集器——拥有自己的 maps 与 ring buffer，注册进进程内注册表。agent 的记录总线与 server 侧同构的接收总线保持**零域分支**：新增一个域不改这些包的任何一行。跨域共享只经 `contracts/<domain>` 纯类型包；禁止 import 兄弟域，由 `scripts/check-layering.sh` 基于 `go list -deps` 强制。

单个 Go module（`github.com/natsubrei/kosmos`）：`cmd/agent`、`cmd/server`、`internal/{agent,server}`、`contracts/<domain>`。跑哪些域由 server 经控制会话下令；Start/Stop 仍是 Runtime handle。只换驱动，不换这套架构。

[数据通道与控制会话](2026-10-27-agent-server-protocol.zh.md)

非目标：不做运行时插件加载；不做配置 DSL 或动态 schema 注册（Record 就是 Go 类型）；不做每域并发框架（Runtime handle 统一管生命周期、排空调度与统计）；不做单机内跨域聚合；不做插件市场、签名与版本协商。

## 考虑过的替代

- **运行时插件**（hashicorp go-plugin、Go `plugin` 包、WASM）：否决——内核侧 C 仍需编译步骤，目标机不应长出工具链或装载器安全面，且单个静态二进制是产品承诺的一部分。
- **配置激活的内建功能**（无接口，全部编入、按开关切换）：否决——场景与阈值的工作会分叉骨架，而不是即插即入。
- **编译期 registry 加接口缝**（选中）：新域 = 新包 + 一行注册；将来真要运行时加载，registry 这道缝还在。

## 后果

新增一个域是新包加一行注册。层级门禁会拦住兄弟域 import 或不含域知识却学了某个域的包。那些包保持零域分支。新域需要重新编译 agent 二进制。
