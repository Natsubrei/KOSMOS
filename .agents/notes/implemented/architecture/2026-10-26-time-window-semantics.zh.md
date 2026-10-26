# Agent Note: Time-window cut on one host

Status: implemented

[English](2026-10-26-time-window-semantics.md) | 中文

## 问题

Report 必须覆盖一扇 Window，Evidence 也可以把该窗当作 id、start、end 快照来依据。若没有统一的切法，每个 Analyzer 会自选区间，两份 Report 就不能表示同一段时间。

## 决策

**Window** 是一台 host 上的 id、host、start、end。`start` 与 `end` 和 `Record.at` 同一墙钟（发出时盖上）；`end` 不含。由 server 上不含域知识的代码切窗。agent（采集端）只发 Record。Analyzer 吃已经切好的一扇窗，不改 `start` 或 `end`。

Record 不带 window id。诊断要跑时才切：切窗代码接受一台 host 和一截已关闭的半开区间（start 起、end 不含），赋 id，把该区间内的 Record 交给该 host 上每个 Analyzer。同一 host、start、end 永远得到同一 id；再跑一次是另一份 Report，不是另一扇 Window。

时间线切成相邻的 10 秒半开区间，按 Unix UTC 对齐（`start` 是 10 秒的整数倍，`end = start + 10s`）。所有 host、所有采集域（Collection domain）共用 server 上这一个常量长度。不滑动，不按事故开窗或平息关窗。Window 不存 parent、prev、next；相邻就是 `end == 下一扇的 start`。

只切已关闭的区间：切窗时 `end` 不晚于 server 墙钟。Evidence 里对窗的依据只能是本 Report 的那一扇，以 id、start、end 嵌在 Report 里（不是去存储做实时查找），不能指向相邻区间。Analyzer 仍然只看见这一扇窗里的 Record。

## 考虑过的替代

- **由 agent 切窗**：否决——agent 不诊断。
- **由 Analyzer 选 `start`/`end`**：否决——等于挑选对断言有利的区间；不含域知识的包也无法为审计约定同一段时间。
- **按事故开窗、按平息关窗**：这次切法否决——那需要采集域信号，是结论怎么推导，不是时间线怎么切。
- **滑动窗**：否决——重叠区间重复劳动，且「同一区间」没有定义。
- **接收时盖上 window id**：否决——agent 不知道 Window；接收路径不该改写 Record 来挂窗。
- **每次诊断发随机 id**：否决——两份 Report 覆盖同一段时间时无法命名同一扇 Window。
- **存 parent/prev/next**：否决——相邻定长区间可以从 `start`/`end` 推出来；这次切法没有嵌套。
- **按域或按 host 各设长度**：否决——切窗代码不含域知识。
- **Evidence 引用相邻 Window**：否决——那是第二扇窗的 Record，和「一窗进」冲突。

## 后果

何时调用切窗（定时器、请求、水位）不是本笔记。区间关闭之后才到的 Record 是存储问题。10 秒这个常量可以改，Window 形状不用改。Window 类型随第一个 Analyzer 在 server 上落地；本笔记不引入它们。`GET /records` 仍是接收和列出。
