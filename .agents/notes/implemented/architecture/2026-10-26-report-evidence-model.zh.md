# Agent Note: Report and evidence-chain data model

Status: implemented

[English](2026-10-26-report-evidence-model.md) | 中文

## 问题

人必须能审计一条诊断为何成立。已调研的 eBPF 产品停在遥测、大模型散文或闭源因果引擎，没有一份可复现的证据链。没有这份文档，`GET /records` 也不能变成 Report。

## 决策

**Report** 是人审计的文档：一条 **Conclusion**、它所依据的 **Evidence**、以及它所覆盖的 **Window**。它是不可变值，含 id、produced_at、host、analyzer、window、conclusion、evidence，类型在 `internal/server/diagnose`。再跑一次是另一份 Report。Report 不记录是谁提出诊断。

Conclusion 是带类型的断言（kind、subject、stance），不是一段话。stance 为成立、不成立或证据不足。**Subject** 是断言所关于的身份：host、pid、tid、starttime；starttime 是 `task_struct.start_time`（自开机起的纳秒），避免 PID 复用把两个任务当成同一个。进程级断言的 tid 为空。comm 与 cgroup 只做注解，不当身份。当前形状是一台主机上的 Linux 任务；第二种形状等那个采集域存在再加。

Conclusion 至少依据一条 Evidence。没有独立的理由字段。Evidence 含 id、kind、payload、依据列表，以及可选的 **Predicate**（observed、operator、bound）。kind 是按采集域加前缀的不透明字符串（`cpu.…`），常量在 Analyzer 包，不进 `contracts/`。不含域知识的包不按 kind 分支。Evidence 的 id 只在本 Report 内唯一。当前 Predicate 的算子是闭集 `lt | le | gt | ge | eq`，只比较数；以后非数值比较仍用 Predicate。

依据列表的每一项是：一条 Record 快照（domain、at、payload，当时用到的那条，不是 Window 内全部）、一个 Window（id、start、end，end 为开区间）、本 Report 内另一条 Evidence 的 id（DAG，无环）、或一个 **Missing source**（source、reason）。证据不足的 Conclusion 至少依据一条 Missing source。聚合可以只引用 Window，把直方图放在 payload。

Analyzer 的接口是：一个 Window 里该域的 Record 进去，零份或多份完整 Report 出来。不另做装配层，也不把多个 Analyzer 的 Report 合成一张；两条断言就是两份 Report。Report 类型在 server 上、在 `contracts/` 之外。agent 和 `contracts/<domain>` 不引用它们；`contracts/` 仍只放 Record。Window 值含 id、host、start、end；Evidence 对窗的依据是 id、start、end。时间线怎么切见：

[一台 host 上的时间窗切法](2026-10-26-time-window-semantics.zh.md)

Report 始终有 Window。

## 考虑过的替代

- **一张问题卡里塞多条 finding**：否决——审计按断言进行；多条断言就是多份 Report。
- **一段散文 Conclusion 加附件**：否决——句子可以和附件打架；带类型的断言加 Evidence 才是证明。
- **Evidence 只留指针，读 Report 要去实时存储 join**：否决——接收侧存储是内存，重启即丢；Record 消失后 Report 仍须可读。
- **不含域知识的包按联合类型分支的 Evidence**：否决——那是把采集域写进必须保持零域分支的包。
- **Analyzer 吐碎片、另做一层装配证据链**：否决——装配层只有一种实现，不必加这一层；Analyzer 才是产出者。
- **由 agent 分配全局 Record id**：否决——agent 不诊断；快照标明用过什么。
- **把多个 Analyzer 的 Report 合成一张卡**：否决——那是以后的产品面，不是这个模型。
- **Conclusion 上加严重程度或置信度**：本模型否决——stance 与 Evidence 已够审计；排序不是本模型的事。

## 后果

类型在 `internal/server/diagnose`。`GET /reports` 编码已存的 Report；人从 Conclusion 跟到 Evidence 再跟到 Record 快照或 Missing source，不必去接收存储做 join。Conclusion 怎么推导、界限在哪，见
[Analyzer 的 Go 推导 Conclusion](2026-10-30-conclusions-in-analyzer-go.zh.md)。
严重程度和问题卡聚合器不是这个模型。
