# Agent Note: Record-path seams for the first vertical slice

Status: implemented

[English](2026-10-26-record-path-seams.md) | 中文

## 问题

agent 已经在进程内总线上发出 Record。现在要把这些 Record 送到 server，并让人能 GET，但不能把 HTTP、JSON 或某个数据库写进域包、总线或 Analyzer——这些都会变。性能也不能靠总线：繁忙主机约 5×10⁵ 次 `sched_switch`/s，用户态不能按事件转发。

## 决策

每个进程各有一条记录总线（`internal/bus`），按 `Domain`（采集域键，如 `"cpu"`）同步扇出，跑在发布者的 goroutine 上。`Deliver` 必须很快返回。总线不负责留下数据，也不是消息队列。

agent 不做诊断。用户态解码、可做轻量富集、按协议预算上送（稳态 10⁵ / 突发 10⁶ 条 Record/秒/主机）。`internal/agent/upload` 是总线订阅者：`Deliver` 不在网络上阻塞；64 槽缓冲满则丢并计数。内核内聚合仍是体量控制。

agent 到 server 是主动外连的 HTTP JSON（`POST /records`，JSON 数组，元素为 `host`、`domain`、`at`、`payload`），不是外部 broker，也不是 gRPC。`cmd/agent -server` 打开上送；域包不 import `internal/agent/upload`。契约包上不打 JSON 标签。

`cmd/server` 听 HTTP。处理函数（`internal/server/httpapi`）把字节变成 Record 并 `Publish`；`internal/server/store` 的 Memory 实现追加/列出，作为总线订阅者把 Record 留下来。`GET /records` 以 JSON 返回该列表。开发认证是非空的 `X-User-Id`。重启 server 列表消失。`Subscribe(cpu)` 的接线在 `cmd/server`，和 `cmd/agent` 里注册域包一样。

## 考虑过的替代

- **把记录总线当成消息队列（Kafka/NATS/RabbitMQ）**：否决——总线是进程内扇出；在单 server、几十 agent 的规模上再加外部 broker 是多出来的运维面。
- **Record 上自由标签、不含域知识的包按标签分支**：否决——`Domain` 是编译进来的域包的采集域键；订阅者注册。
- **在 agent 上诊断**：否决——占用发生在目标机；诊断属于 server 的 Analyzer。
- **用户态无界排队以免丢数据**：否决——内存占用会爆；改为丢并计数。
- **现在就上 gRPC + protobuf**：否决——Tick 不需要；给人的 GET 已经是 JSON。以后换传输是再写订阅者/接收端。
- **在 `contracts/` 上打 JSON 标签，或 Analyzer 直接读 HTTP body**：否决——那是把线协议焊进词汇。
- **现在就上 PostgreSQL（或 SQLite）**：否决——追加/列出后面的 Memory 就够；空适配器是债。
- **GET 返回证据链 Report**：否决（仅这次）——模型见下。`GET /records` 仍是接收和列出。

[Report 与 Evidence 链](2026-10-26-report-evidence-model.zh.md)

## 后果

对 `GET /records` 做 curl 即可验收。JSON 扛不住 10⁵ 条 Record/秒/主机：改 `upload` 和 `httpapi`，不改记录总线、运行时工具箱或契约包。内存存储不能熬过重启。Record 还带 host（见下）；JSON 数组随诊断循环加上该字段。数据通道重连就是 HTTP 失败则丢；背压是 64 槽缓冲。采集激活是控制会话，不是这条路径。

[数据通道与控制会话](2026-10-27-agent-server-protocol.zh.md)

[内存存储与接收/诊断边界](2026-10-27-storage-async-boundary.zh.md)
