---
labels: [wayfinder:grilling]
assignee:
---

# 0004 · Time-window semantics

## Question

How is the timeline cut for diagnosis: fixed-length windows, incident-anchored windows opened by a trigger and closed by quiescence, or both? Who defines window boundaries (agent, server, or diagnosis rules), how do windows reference each other (parent/adjacent), and what must every window carry so evidence can cite it? Decide with the [report model](../.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md) in view; keep both consistent. A Report already requires a Window `{id, start, end}` (end exclusive); this ticket fills how the timeline is cut and what else a Window carries.
