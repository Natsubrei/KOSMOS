# Agent Notes（代理笔记）

[English](README.md) | 中文

耐久的决策依据：选了什么、代价是什么、放弃了什么、该决策需要什么验证。

## 何时写

当一个变更的理由没有任何现有归宿——代码注释、提交信息、既有笔记都不承载它（非显然的成因、流程规则、值得记住的放弃项）——就写一篇笔记。机械变更豁免。使既有笔记失效的变更，在同一批变更里更新那篇笔记。

进行中的工作留在 `issues/`。不要为未关的票预写 `proposed/` 笔记：有了决策再写。

## 布局与状态

`{proposed,implemented,rejected}/{architecture,feature,bug-fix,process,simplification,testing}/YYYY-MM-DD-slug.md`

日期是该主题首次提出的时间。笔记在生命周期目录之间移动时不改文件名。

- **proposed（提议）**——讨论中；可以包含计划与"应该"。
- **implemented（已落地）**——已发布的现实，用现在时；不写 spec 腔、迁移计划、验收清单。
- **rejected（已否决）**——考虑过并否决，连同理由。

| 类 | 覆盖 |
|---|---|
| `architecture` | 针对已发布源码的结构性决策。 |
| `feature` | 用户可感知的新能力。 |
| `bug-fix` | 纠正缺陷，或补上事后分析发现的缺口。 |
| `process` | 代码周围的工具、策略或工作流。 |
| `simplification` | 删代码、行为或表面，不加能力。 |
| `testing` | 测试基础设施与策略。 |

## 文件格式

```markdown
# Agent Note: <title>

Status: <proposed|implemented|rejected — <why>>
```

已落地笔记随后使用 `## Problem`、`## Decision`、`## Alternatives considered`、`## Consequences`。提议笔记用 `## Proposal` 代替 `## Decision`，并加 `## Acceptance criteria` 与 `## Risks`。每篇笔记记录每个真正考虑过的替代以及它为何落选。笔记是双语对，受 `scripts/verify-translation-pair.sh` 门禁。
