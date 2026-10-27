# Terminology

本表约定本仓库的中英术语统一译法。含义以 [glossary.md](../glossary.md) 为准；本表只管怎么写。本文件自身是双语对照，不是 `foo.md` + `foo.zh.md` 对。

**通用规则：**
- 「中文」列为中文译文的正文默认用词。若该列为英文，则中文译文的正文中保留英文不翻译。
- 首次出现列有内容时：该中文文件里第一次按该列写（带括号）；后文只写括号前的部分。列为空则始终用「中文」列，不括注。英文类要不要括一次：读者可能读不懂时括（教一次，后文仍是英文）；已经当作规范英文、或「不要译作」在防中文别名时留空。缩写类的括注是展开全称，不是翻译。
- 「不要译作」列为严格禁止的译法。
- 如果某术语已经作为另一个术语的组成部分被括注过，则该术语后续单独出现时无需再次括注。
- 未收录的技术术语：有通行中文译法（可引用）才翻译；否则留英文。禁止当场发明译法；定下来的词在同一次变更里加入本表。
- 本表由作者在改中文时遵守。`scripts/verify-translation-pair.sh` 不检查译法、标点或空格；不要为「不要译作」、首次括注、标点或中西文空格加扫描器。怎么翻译一对文档见 [translation-rules.md](translation-rules.md)。

## 缩写类（中英文文本中均使用缩写）

| English | 中文 | 首次出现 | 不要译作 | 备注 |
|---|---|---|---|---|
| API | API | | | |
| CPU | CPU | | | |
| eBPF | eBPF | | | |
| HTTP | HTTP | | | |
| JSON | JSON | | | |
| PID | PID | | | |
| TID | TID | | | |

## 英文类（中英文文本中均使用英文）

| English | 中文 | 首次出现 | 不要译作 | 备注 |
|---|---|---|---|---|
| agent | agent | agent（采集端） | 代理、智能体 | 目标机上的采集进程；不是 LLM agent |
| Agent Note | Agent Note | | 代理笔记、智能体笔记、智能体注记 | 仓库定义的文档类型；中文对侧 H1 保持 `# Agent Note: `，标题中不加术语括注 |
| Analyzer | Analyzer | | | |
| Collector | Collector | | | |
| Conclusion | Conclusion | | | |
| Evidence | Evidence | | | |
| Missing source | Missing source | Missing source（缺失的源） | gap、缺口 | |
| Predicate | Predicate | Predicate（谓词） | | |
| QEMU | QEMU | | | |
| Record | Record | | 事件、event | 观测；内核事件仍称 kernel event |
| Report | Report | | | |
| Runtime handle | Runtime handle | | | |
| Subject | Subject | | 资源、对象 | |
| Window | Window | | | |
| server | server | | | |

## 双语类（中英文文本各自使用中英文）

| English | 中文 | 首次出现 | 不要译作 | 备注 |
|---|---|---|---|---|
| Collection domain | 采集域 | 采集域（Collection domain） | target、probe | |
| Contracts | 契约包 | 契约包（Contracts） | | |
| Domain bundle | 域包 | 域包（Domain bundle） | plugin | |
| Record bus | 记录总线 | 记录总线（Record bus） | 消息队列 | |
| Registry | 注册表 | 注册表（Registry） | | |
| Runtime kit | 运行时工具箱 | 运行时工具箱（Runtime kit） | | |
| Zero domain branches | 零域分支 | 零域分支（Zero domain branches） | | |
