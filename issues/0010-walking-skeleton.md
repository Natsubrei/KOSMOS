---
labels: [wayfinder:task]
assignee:
---

# 0010 · Walking skeleton: CPU domain end-to-end

## Question

Implement and accept the destination: agent collects CPU-scheduling evidence via eBPF on the QEMU guest, reports to the containerized server, which aggregates windows and emits an auditable evidence-chain report (JSON over API) under development auth. Acceptance is human-reviewed: the report for a synthetic CPU-stress incident must localize to a PID/TID with evidence a person can follow. This ticket executes; every earlier ticket decides.

Blocked by: 0003, 0004, 0005, 0006, 0009.
