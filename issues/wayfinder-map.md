---
labels: [wayfinder:map]
---

# Wayfinder map: KOSMOS rewrite → CPU-domain walking skeleton

## Destination

Every architectural decision for the rewrite settled, plus a walking skeleton proving the shape end-to-end for one diagnosis domain: an agent collects CPU-scheduling evidence via eBPF on a Linux VM, reports to the central server, which aggregates time windows and emits an auditable evidence-chain report (JSON over API). The skeleton is the map's final ticket; the map decides first, then builds only that.

## Notes

- Clean-room: never consult the pre-rewrite implementation on any git branch or in its docs; derive domain knowledge from public sources and first principles only.
- Stack anchored at charting: Go for agent and server; eBPF via cilium/ebpf (CO-RE); no further stack debate.
- Test-guest image: openEuler 24.03 LTS (kernel 6.6); Ubuntu 24.04 is the documented fallback ([note](../.agents/notes/implemented/testing/2026-10-25-qemu-guest-test-image.md)).
- Scale: single server, tens of agents; no multi-server / HA design in this effort.
- Skeleton auth is development mode (trusted `X-User-Id` header); the real auth model is ticket 0008.
- Tickets are agent-facing working artifacts: English only, no bilingual pairs.
- Tracker (local markdown): map at `issues/wayfinder-map.md`, tickets at `issues/NNNN-slug.md`, blocking via `Blocked by` lines in ticket bodies; the frontier is open tickets whose blockers are all closed and that carry no assignee.
- Research findings are local working material under `docs/research/` (gitignored). Durable choices they produced live in Agent Notes; tickets do not keep a copy.
- Every session working this map: consult the grilling and domain-modeling skills before a grilling ticket; research tickets resolve via the research skill.

## Decisions so far

<!-- one line per closed ticket, newest last -->

- 0001 · CPU-scheduling collection constraints: baseline kernel 5.8, floor 4.9; in-kernel aggregation; protocol budget 10⁵–10⁶ events/s/host ([note](../.agents/notes/implemented/architecture/2026-10-25-cpu-sched-collection-constraints.md)).
- 0002 · QEMU guest image: openEuler 24.03 LTS, Ubuntu 24.04 fallback; unprivileged QEMU ([note](../.agents/notes/implemented/testing/2026-10-25-qemu-guest-test-image.md)).
- 0011 · Compile-time domain bundles over a zero-branch record bus ([note](../.agents/notes/implemented/architecture/2026-10-25-compile-time-domain-bundles.md)).
- 0003 · Report is one Conclusion plus Evidence over one Window; Analyzers produce complete Reports ([note](../.agents/notes/implemented/architecture/2026-10-26-report-evidence-model.md)).

## Not yet specified

- Agent binary distribution / packaging (rpm / deb / raw tarball) — sharpens after the protocol and skeleton shape are known.
- TLS and agent–server channel hardening — depends on the protocol choice and the auth model.
- Retention / rollover policy details beyond what ticket 0005 resolves — revisit with real event volumes from the skeleton.
- Whether and when an AI-assisted diagnosis layer enters — deferred until a deterministic evidence-chain report exists to build on.

## Out of scope

- Web UI / workbench: the destination's report is JSON over API; rendering it is a separate future effort.
- Multi-server / HA / horizontal scaling: single-writer central server by design for this effort.
- Non-CPU diagnosis domains in this map's skeleton (locks, memory reclaim, IO, syscall): future efforts, charted fresh after the CPU domain proves the chain.
