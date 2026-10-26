# Agent Notes

English | [中文](README.zh.md)

Durable decision rationale: what was chosen, what it cost, what was given up,
and the verification the decision requires.

## When to write one

Write a note when a change carries rationale that no code comment, commit
message, or existing note already owns — non-obvious causes, process rules,
rejected alternatives worth remembering. Mechanical changes are exempt. A
change that invalidates a note updates it in the same change.

In-progress work stays in `issues/`. Do not pre-write a `proposed/` note for an
open ticket: a note starts when the decision exists.

## Layout and status

`{proposed,implemented,rejected}/{architecture,feature,bug-fix,process,simplification,testing}/YYYY-MM-DD-slug.md`

The date is when the topic was first proposed. Moving a note between lifecycle
folders does not rename it.

- **proposed** — under discussion; may hold plans and "should".
- **implemented** — shipped reality, present tense; no spec-speak, no
  migration plans, no acceptance checklists.
- **rejected** — considered and declined, with the reason.

| Class | What it covers |
|---|---|
| `architecture` | A structural decision about the shipped source. |
| `feature` | A new user-facing capability. |
| `bug-fix` | Corrects a defect or closes a gap a postmortem surfaced. |
| `process` | Tooling, policy, or workflow around the code. |
| `simplification` | Removes code, behavior, or surface area without adding a capability. |
| `testing` | Test infrastructure and strategy. |

## The file format

```markdown
# Agent Note: <title>

Status: <proposed|implemented|rejected — <why>>
```

Implemented notes then use `## Problem`, `## Decision`, `## Alternatives
considered`, `## Consequences`. Proposed notes use `## Proposal` instead of
`## Decision`, plus `## Acceptance criteria` and `## Risks`. Every note records
each genuine alternative and why it lost. Notes are bilingual pairs under
`scripts/verify-translation-pair.sh`.
