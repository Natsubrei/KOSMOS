# KOSMOS

[English](glossary.md) | 中文

KOSMOS 诊断 Linux 虚拟机与物理主机上的异常：agent 经 eBPF 在内核内取证，中心服务器执行诊断。本术语表约定项目自身的词汇。

## 词汇

**采集域（Collection domain）**：
机器上一个可诊断的领域，例如 CPU 调度。是 agent 与 server 两侧共同的扩展单元；本表其余词条的存在，都是为了让新增一个采集域的成本足够低。
_避免使用_：target、probe、module

**域包（Domain bundle）**：
采集域在 agent 侧的自包含实现：eBPF C 程序、map schema、解码器、记录类型、可选的富集器——连同自己的 maps 与 ring buffer。一个目录，一行注册代码。
_避免使用_：plugin（我们编译期编入，没有运行时加载）

**采集器（Collector）**：
域包在核心面前扮演的角色：核心认识的唯一接口——为自己的采集域启动与停止采集。
_避免使用_：driver、provider

**分析器（Analyzer）**：
域包在 server 侧的对应实现：在 server 总线上订阅该域的 Record，产出 Report。
_避免使用_：rule engine、processor、finding

**记录总线（Record bus）**：
进程内的类型化记录管线。agent 与 server 各有一条，都保持零域分支——新的记录种类靠注册进入，从不靠修改核心。
_避免使用_：event loop、message queue

**零域分支（Zero domain branches）**：
核心的规则（编译期域包笔记）：核心代码不为任何具体采集域写分支——核心知道记录带着域标签，但从不知道任何一个具体的域。域行为只能经注册进入进程，因此新增一个域时，核心目录零改动。
_避免使用_：microkernel

**注册表（Registry）**：
域包与分析器注册的进程内表；核心与域之间的唯一接线点。
_避免使用_：plugin host、container

**Runtime 句柄（Runtime handle）**：
运行时工具箱注入每个域包的一组服务：生命周期、时钟换算、统计与丢弃计数、eBPF 加载、排空调度。域包访问宿主进程的唯一途径。
_避免使用_：framework、context

**契约包（Contracts）**：
agent 与 server 共享的、按域划分的纯类型包（`contracts/<domain>`）：记录类型与常量，零实现。
_避免使用_：API、schema（这两个词留给线上协议）

**记录（Record）**：
域包发布的一条观察：载荷、所属域、发出时由 runtime 盖上的墙钟时间。一次原始的内核发生（如一条 sched_switch）是内核事件；记录是内核内聚合从事件中产出的东西。
_避免使用_：event（留给内核事件）、message

**运行时工具箱（Runtime kit）**：
每个域包经 Runtime 句柄共享的宿主进程机制：生命周期、时钟盖戳、统计、eBPF 加载、排空调度。不含任何域知识。
_避免使用_：framework、kernel

**核心（Core）**：
每个进程里不含域知识的部分：agent 侧是记录总线与运行时工具箱；server 侧是共享的摄取/窗口/报告管线。域只能经注册进入核心。
_避免使用_：kernel、host

## 诊断

**报告（Report）**：
人审计的文档：一条 Conclusion、它所依据的 Evidence、以及它所覆盖的 Window。不可变；再跑一次是另一份 Report。
_避免使用_：finding、problem card、diagnosis result

**结论（Conclusion）**：
对一个 Subject 的带类型断言：kind、该 Subject、以及 stance（成立、不成立或证据不足）。给人看的句子由这三项生成，不作为主数据存储。
_避免使用_：finding、verdict、root cause

**证据（Evidence）**：
Conclusion 所依据的一条：不透明的 kind、payload、以及依据列表（一条 Record 快照、一个 Window、本 Report 内另一条 Evidence、或一个 Gap）。
_避免使用_：citation、attachment、justification

**主体（Subject）**：
Conclusion 所断言的类型化身份：host、pid、tid、以及任务 starttime（避免 PID 复用把两个任务当成同一个）。进程级断言的 tid 为空。comm 与 cgroup 只做注解，不当身份。
_避免使用_：resource、entity、target、object

**窗口（Window）**：
一份 Report 所覆盖的时间区间，用 id 加上半开的起止引用。
_避免使用_：period、bucket、range

**谓词（Predicate）**：
记录在 Evidence 上的可选数值比较：观测值、闭集中的算子、以及这次用过的边界。
_避免使用_：threshold（配置中的边界；Predicate 是这次打出来的比较）

**缺口（Gap）**：
该有却没有的源，带原因。证据不足的 Conclusion 至少依据一条 Gap。
_避免使用_：missing data、error
