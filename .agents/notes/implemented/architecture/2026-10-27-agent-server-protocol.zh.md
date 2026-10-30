# Agent Note: Agent–server protocol: data path and control session

Status: implemented

[English](2026-10-27-agent-server-protocol.md) | 中文

## 问题

`POST /records` 已经用 HTTP JSON 运 Record。agent（采集端）多半在 NAT 后面，server 拨不回去。域包编译进来不等于正在采集：全域都跑会空耗 CPU。没有存活信号时，安静的主机和死掉的 agent 看起来一样。数据传输以后会换（gRPC 是候选）；记录总线和域包不能绑死在某一种传输上。

## 决策

两条通道，都由 agent 出站，这次都是 HTTP，不上 TLS。

**数据通道**仍是 `POST /records`（Record 的 JSON 数组，每条带 `host`）。HTTP 失败则丢掉这一批；背压是 `upload` 的 64 槽缓冲。Record 不重试、不落盘。

**控制会话**是 agent 打出去的长连接。server 对每个 `host` 一个句柄。命令是「现在该跑的域」的幂等快照（已知名 `Start`，未知名忽略；快照里没有的就是 `Stop`）。会话加上应用层保活就是存活：连着就是活着；存活不进记录总线。保活失败开始计 60 秒；60 秒内采集保持上一份快照；满 60 秒 agent 对全部域包 `Stop`。之后重连，再等新快照。同一 `host` 的新会话顶掉旧句柄，并关掉旧连接。

尚未收到快照时，任何域包都不跑。这次 server 在会话出现时下发 `{cpu}`。

会话身份和每条 Record 上的 `host` 是同一个字符串。agent 怎么选定该字符串不是本笔记。

记录总线、运行时工具箱、`contracts/` 不 import HTTP、JSON 或 gRPC。以后换传输，换 `internal/agent/upload`、`internal/server/httpapi` 和控制会话的两端，不换 Publish/Deliver。

## 考虑过的替代

- **没有命令通道；跟进程走或本地启用列表**：否决——server 封不住开销，也分不出安静但活着的 agent 和死掉的 agent。
- **server 去拨 agent**：否决——NAT。
- **定时拉取，不建会话**：否决——server 需要对每个 agent 的句柄，才能不等下一轮拉取就推新快照。
- **心跳做成 Record**：否决——存活不是证据；不能进总线、Window 或 Analyzer。
- **Record 走控制会话**：否决——证据的帧不能焊在会话上；换传输时两个接收端一起换。
- **会话一断就停采**：否决——短抖动会拆掉 eBPF。60 秒是空耗上限。
- **这次上 TLS 或 enrollment**：否决——开发鉴权仍是 `X-User-Id`；加固是以后的鉴权模型。
- **这次上 gRPC**：否决——给人看的 GET 已是 JSON；缝在订阅者和接收端，不是现在再加一套栈。

## 后果

`internal/agent/control` 拨 `POST /control`；`internal/server/control` 对每个 `host` 一个句柄，连上时下发 `{cpu}`。保活 10 秒；会话死后 60 秒停采。Start/Stop 仍是 Runtime handle。TLS 和 enrollment 是以后的鉴权模型。

[记录路径缝](2026-10-26-record-path-seams.zh.md)

[内存存储与接收/诊断边界](2026-10-27-storage-async-boundary.zh.md)

[编译期域包](2026-10-25-compile-time-domain-bundles.zh.md)
