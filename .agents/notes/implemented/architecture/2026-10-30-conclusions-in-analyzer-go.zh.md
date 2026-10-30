# Agent Note: Conclusions derived in Analyzer Go

Status: implemented

[English](2026-10-30-conclusions-in-analyzer-go.md) | 中文

## 问题

报告模型说的是 Report 是什么，不是 Analyzer 凭什么产出一份。没有这条决定，规则引擎、配置存储、或诊断工具箱上的「所需源」方法，都会看起来像下一处 seam。

## 决策

Analyzer 包用 Go 推导 Conclusion。没有 scenario 文档，没有规则语言，也没有解释器。两个 Conclusion 仍是两份 Report；没有东西合并 Analyzer 的产出。

数值界限是 Analyzer 实例上的字段。该包的构造函数从常量填入（`cpu.New` 填 `defaultNs`，即 1 秒）。`Analyze` 读这个字段，把当时用的数抄进它所写的 Evidence 上的 `Predicate.Bound`。诊断工具箱、记录总线、`cmd/server` 都不知道界限。`diagnose.Analyzer` 不加界限方法，也不加所需源方法。一种 Conclusion kind 对应一个字段，不做 `map[kind]bound`。数字从哪来是构造函数的事；那个来源变了，`Analyze` 和 `Loop` 不变。

`Analyze` 收到空 records，或比较未触发时，返回零份 Report。`does_not_hold` 留给写出否定 Conclusion 的 Analyzer，不是对未超界任务做普查。Missing source 由该 Analyzer 的 Go 在写出一条完成不了的 Conclusion 时自己写入。当前 `cpu.on-cpu` 路径只需要 `TaskRuntime`，所以不写 Missing source。

## 考虑过的替代

- **运行时 scenario / 规则引擎：** 否决——等于第二种诊断语言，也和编译期域包冲突。
- **用配置文件、数据库或 HTTP API 存放界限：** 否决——没有这些子系统，写接口也没有鉴权模型。Predicate 已经记下当时用的 bound。
- **在 `diagnose.Analyzer` 上加 `RequiredSources()`：** 否决——那是让工具箱去诊断「空」；活着与否不是 Record，空 Window 也分不出空闲、没在采、和死掉的 agent。
- **每个空 Window 都出证据不足的 Report：** 否决——`GET /reports` 会变成安静主机的普查。
- **未超界的每个任务都出 `does_not_hold`：** 否决——同样是普查，而且按线程。
- **在诊断工具箱上加 Bounds 接口：** 否决——每个 Analyzer 一个实现；空适配器是债。下一步是构造函数从别处填这个字段。
- **`map[kind]bound`：** 否决——第二种 kind 在那条 Conclusion 出现时再加字段。

## 后果

`internal/server/analyzer/cpu` 已经是这个形状：`bound` 是字段，`New` 填 `defaultNs`，`Analyze` 把它抄进 Predicate。同包测试直接设 `a.bound`。`internal/server/diagnose` 里的工具箱不变。

steal time 和 off-CPU 仍未采集；需要它们的 Conclusion 是以后改这个 Analyzer 的事，也是这个域第一次出现 Missing source 的时候。

[报告与证据链数据模型](2026-10-26-report-evidence-model.zh.md)
