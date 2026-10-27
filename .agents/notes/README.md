# Agent Notes

English | [中文](README.zh.md)

Durable decision rationale: what was chosen, what it cost, what was given up,
and the verification the decision requires.

## When to write one

Write a note when a change carries rationale that no code comment, commit
message, or existing note already owns — non-obvious causes, process rules,
rejected alternatives worth remembering. Mechanical changes are exempt. A
change that invalidates a note updates it in the same change.

When the decision exists and the code does not, the note starts in
`proposed/`. Move it to `implemented/` in the same change that ships the
code. Do not write a note before the decision exists.

## Layout and status

`{proposed,implemented,rejected}/{architecture,feature,bug-fix,process,simplification,testing}/YYYY-MM-DD-slug.md`

The date is when the topic was first proposed. Moving a note between lifecycle
folders does not rename it.

- **proposed** — decided, not yet in the tree or only partly built. Future
  tense is allowed. `## Proposal` is the intended shape; `## Acceptance criteria`
  is what observable state means done; `## Risks` is what can go wrong and what
  the change knowingly gives up.
- **implemented** — shipped reality, present tense. Kept current with paths,
  symbols, and defaults in the same change that alters them. No spec-speak, no
  migration plans, no acceptance checklists. Reversing the decision is a new
  note plus a cross-link, not an edit of this file.
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

Implemented notes use `## Problem`, `## Decision`, `## Alternatives
considered`, `## Consequences`. Proposed notes use `## Proposal` instead of
`## Decision`, plus `## Acceptance criteria` and `## Risks`. Moving
`proposed/` → `implemented/` rewrites Proposal into present-tense Decision and
folds Acceptance and Risks into Consequences, in that same change. Every note
records each genuine alternative and why it lost. Notes are bilingual pairs
under `scripts/verify-translation-pair.sh`.
