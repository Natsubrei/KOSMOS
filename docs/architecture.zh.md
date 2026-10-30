# KOSMOS 代码结构

[English](architecture.md) | 中文

这是代码现在的样子。决定和理由在 [Agent Notes](../.agents/notes/README.zh.md) 里；改一处之前先读拥有它的那篇笔记。

## 两个进程

`cmd/agent` 跑在被诊断的机器上采集。`cmd/server` 负责接收和诊断。两者共用一个 Go module，`github.com/natsubrei/kosmos`，除此之外不共享：跨进程只走线上协议，不互相 import 包。

## Agent

`internal/bus` 是进程内的记录总线。它把一条 Record 同步扇出给订阅了该 Record 的 `Domain` 的所有订阅者，跑在发布者的 goroutine 上。

`internal/agent/runtime` 是运行时工具箱。它拥有采集器生命周期，并把一个 `Handle` 交给每个域包，那是域包唯一的发布途径。`Register` 登记采集器；`Apply(domains)` 是启停它们的快照；`Run` 撑住进程寿命。

`contracts/<domain>` 只放 Record 类型和域键，别的都不放。`internal/agent/domain/<domain>` 是产出它们的域包。

`internal/agent/hostid` 是选定机器身份的唯一地方。每条 Record 和控制会话都带这个字符串。

`internal/agent/upload` 用 `POST /records` 上送 Record。`internal/agent/control` 拨 `POST /control`，收到域快照就调 `Apply`；会话死了，宽限期过后停采。

## Server

`internal/server/httpapi` 把请求字节变成 Record 并发布；`GET /records` 列出存储留下的东西；`GET /reports` 列出诊断产出。`internal/server/control` 对每个 `host` 持有一条会话，下发该 host 该跑的域快照。`internal/server/store` 把 Record 留在进程内存里。

`internal/server/diagnose` 是诊断工具箱。它按 10 秒的 Unix-UTC 切开 Window，并调用每一个已注册的 `Analyzer`；Analyzer 是一窗某一域的 Record 进、零或多份完整 Report 出。每个 Analyzer 在自己的 Go 里推导 Conclusion；数值界限是该实例上的字段，不在诊断工具箱上。`internal/server/analyzer/<domain>` 是实现那一处的包。

`cmd/server` 是把域键接到记录总线、控制快照和 Analyzer 的唯一地方，就像 `cmd/agent` 是注册域包的唯一地方。

## seam

下面每一处，都是某个会变的东西唯一被允许出现的地方。

| 关注点 | seam | 包 |
|---|---|---|
| 线上格式与传输 | 订阅者与接收端 | `internal/agent/upload`、`internal/agent/control`、`internal/server/httpapi`、`internal/server/control` |
| Record 留存 | `store.Store` | `internal/server/store` |
| 采集生命周期 | `runtime.Collector`、`Agent.Apply` | `internal/agent/runtime` |
| 诊断 | `diagnose.Analyzer`、`Loop` | `internal/server/diagnose` |
| 有哪些 Record 种类和 Analyzer | 域键加一行注册 | `contracts/`、`internal/server/analyzer/<domain>`、`cmd/` |

记录总线、运行时工具箱、诊断工具箱和 `contracts/` 不 import 任何传输，也不 import 任何域。域那一半由 `scripts/check-layering.sh` 强制。

## 加一个采集域

照 [new-collection-domain](../.agents/skills/new-collection-domain/SKILL.md) 做。它列出域会碰到的五处，以及保证其余部分不动的门禁。

## 决定在哪里

[Agent Notes](../.agents/notes/README.zh.md) 承载理由：选了什么、代价是什么、打败了什么。改一条 seam 之前先读拥有它的笔记；放在 `proposed/` 的笔记表示已经决定，但还没进这份代码。
