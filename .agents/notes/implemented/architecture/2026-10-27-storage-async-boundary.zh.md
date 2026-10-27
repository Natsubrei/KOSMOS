# Agent Note: In-memory store and the ingest/diagnosis boundary

Status: implemented

[English](2026-10-27-storage-async-boundary.md) | 中文

## 问题

诊断需要按 host 把 Record 切进已关闭的 Window，并且 Report 不必去接收存储做 join 仍可读。接收路径已经把 Record 放在进程内存里。在一台测试机还没跑出量之前，上数据库、第二套消费者模型、或多段处理 pipeline，都是多出来的面。

## 决策

Record、Report、以及已诊断区间的集合都活在 server 进程里。重启即丢。这次不上 PostgreSQL、SQLite、Timescale，也不留空的存储适配器。没有滚动删除：留存等于进程寿命。

**Record** 带 `host`：被采集的那台机器，由 agent（采集端）在发出时盖上，随 `POST /records` 发送。server 不从 TCP 对端推断 host。切窗和已诊断集合都按这个 `host` 分组。

`POST /records` 只做追加。诊断是进程内另一条 ticker 循环、单个消费者：切最早一扇已关闭且尚未诊断的 Window（`end` 不晚于 server 墙钟），跑 Analyzer，即使 0 份 Report 也把该 host、start、end 标成已诊断。不在每次接收时跑，也不提供「请诊断」的 HTTP。

没有 Window 表。Window 在切窗时算出，嵌进 Report。迟到的 Record（`at` 落在已诊断区间）仍追加；已有 Report 不变；循环不再切该区间。

接收是总线扇出进内存存储。诊断是后面那个单消费者。不是多段 pipeline。

## 考虑过的替代

- **现在就上 PostgreSQL**：否决——一台 server、几十 agent 不够再开一个进程；空适配器是债。
- **现在就上 SQLite**：这次否决——撑过重启等骨架跑出量再说；那时候首选 SQLite 单文件，不是 PostgreSQL。
- **TimescaleDB 一类**：否决——还没有时序负载就上时序引擎。
- **在 `POST /records` 里诊断**：否决——`Deliver` 必须快返回；诊断不能堵住接收。
- **多个诊断消费者**：否决——同一 host、start、end 切两次是两份没有顺序的 Report。
- **只切最新已关闭的窗**：否决——合成事故可能落在更早的一格。
- **单存一份 Window 集合**：否决——身份就是 host、start、end；Report 里已有快照。
- **迟到 Record 再跑同一扇窗**：否决——再跑是另一份 Report；本循环每区间一次。
- **从 TCP 对端推断 host**：否决——NAT 会指向错误的机器。
- **把这条路径叫 pipeline**：否决——总线是扇出；诊断是后面一个消费者，不是一串阶段。

## 后果

ticker 间隔和列出 Report 的 HTTP 路径，随第一个 Analyzer 落地。`GET /records` 仍是接收和列出。Record 的 JSON 数组带 `host`。agent 怎么选定该字符串是 `internal/agent/hostid`；总线只运结果。Conclusion 怎么推导仍不是本笔记。
