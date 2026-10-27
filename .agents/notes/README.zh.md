# Agent Notes

[English](README.md) | 中文

耐久的决策依据：选了什么、代价是什么、放弃了什么、该决策需要什么验证。

## 何时写

当一个变更的理由没有任何现有归宿——代码注释、提交信息、既有笔记都不承载它（非显然的成因、流程规则、值得记住的放弃项）——就写一篇笔记。机械变更豁免。使既有笔记失效的变更，在同一批变更里更新那篇笔记。

决策已经有了、代码还没有时，笔记从 `proposed/` 起。代码落地的那次变更里搬进 `implemented/`。决策还没有时不要写笔记。

## 布局与状态

`{proposed,implemented,rejected}/{architecture,feature,bug-fix,process,simplification,testing}/YYYY-MM-DD-slug.md`

日期是该主题首次提出的时间。笔记在生命周期目录之间移动时不改文件名。

- **proposed（提议）**——已决定，尚未进树或只建了一部分。可以用将来时。`## Proposal` 是打算长成的样子；`## Acceptance criteria` 是什么可观察状态算完成；`## Risks` 是可能出错的地方以及这次明知放弃的东西。
- **implemented（已落地）**——已发布的现实，用现在时。路径、符号、默认值随代码同一次变更更新。不写 spec 腔、迁移计划、验收清单。推翻该决定是新开一篇并交叉链接，不是改这篇。
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

已落地笔记使用 `## Problem`、`## Decision`、`## Alternatives considered`、`## Consequences`。提议笔记用 `## Proposal` 代替 `## Decision`，并加 `## Acceptance criteria` 与 `## Risks`。从 `proposed/` 搬到 `implemented/` 时，把 Proposal 改写成现在时的 Decision，把 Acceptance 与 Risks 折进 Consequences，就在那一次变更里做。每篇笔记记录每个真正考虑过的替代以及它为何落选。笔记是双语对，受 `scripts/verify-translation-pair.sh` 门禁。
