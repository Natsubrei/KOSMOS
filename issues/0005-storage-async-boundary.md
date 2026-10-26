---
labels: [wayfinder:grilling]
assignee:
---

# 0005 · Central storage and the async boundary

## Question

What stores events, windows, and reports on the central server, and where is the asynchronous boundary between ingestion and diagnosis? Decide: store choice (PostgreSQL is the leading candidate; justify against SQLite/TimescaleDB-class alternatives), schema shape for events vs windows vs reports, retention/rollover stance, and the single-writer processing model (one diagnosis consumer over a queue/table). Ground the decision in the scale target: one server, tens of agents.
