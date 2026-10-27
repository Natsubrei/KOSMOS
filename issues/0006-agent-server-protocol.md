---
labels: [wayfinder:grilling]
assignee:
---

# 0006 · Agent–server protocol and collection activation

## Question

Outbound upload is already HTTP JSON (`POST /records`) as in the [record-path](../.agents/notes/implemented/architecture/2026-10-26-record-path-seams.md) note; a Record carries `host` as in the [storage](../.agents/notes/implemented/architecture/2026-10-27-storage-async-boundary.md) note. Remaining: command channel (if any), reconnection, and collection start/stop (always-on vs server-instructed), given a credential-free agent on a possibly-NATed host and the overhead budget in the [CPU-scheduling collection constraints](../.agents/notes/implemented/architecture/2026-10-25-cpu-sched-collection-constraints.md) note.
