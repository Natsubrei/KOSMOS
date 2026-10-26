---
labels: [wayfinder:grilling]
assignee:
---

# 0005 · Central storage and the async boundary

## Question

What stores Records, Windows, and Reports on the central server, and where is the asynchronous boundary between ingestion and diagnosis? Windows are cut at diagnosis as in the [time-window note](../.agents/notes/implemented/architecture/2026-10-26-time-window-semantics.md); this ticket decides persistence, late Records after a closed interval, store choice (PostgreSQL is the leading candidate; justify against SQLite/TimescaleDB-class alternatives), schema shape, retention/rollover, and the single-writer processing model (one diagnosis consumer over a queue/table). Ground the decision in the scale target: one server, tens of agents.
