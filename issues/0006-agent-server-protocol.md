---
labels: [wayfinder:grilling]
assignee:
---

# 0006 · Agent–server protocol and collection activation

## Question

How do agent and server talk: transport (long-poll commands + event upload vs gRPC streaming vs WebSocket), serialization (protobuf vs JSON), backpressure, and reconnection semantics — given a credential-free agent on a possibly-NATed host and a containerized single server? Include the collection-activation stance: always-on probes vs server-instructed start/stop, and how the overhead budget in the [CPU-scheduling collection constraints](../.agents/notes/implemented/architecture/2026-10-25-cpu-sched-collection-constraints.md) note shapes it.
