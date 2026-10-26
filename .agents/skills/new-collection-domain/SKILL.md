---
name: new-collection-domain
description: Add a KOSMOS collection domain (a diagnosable area like CPU scheduling, locks, reclaim, IO) — domain bundle, contracts, registration, analyzer, gates. Use when adding or extending a diagnosis domain or wiring a new one into the agent.
---

# New collection domain

Five places, then the gates; domain-free packages stay untouched
([compile-time domain bundles](../../notes/implemented/architecture/2026-10-25-compile-time-domain-bundles.md)).

1. **Contracts** — `contracts/<domain>/`: the domain key constant and record
   types. Pure types, zero module imports; layering rule 1 rejects any.
2. **Bundle** — `internal/agent/domain/<name>/`: the self-contained agent
   side. Kernel-side collector (eBPF C; a timer stub while proving the
   chain), map schema, decoder to the contract records, optional enricher.
   Implements `runtime.Collector`; Stop drains and leaves no goroutine. Own
   maps and ring buffer. The layering gate fails any sibling-domain import.
3. **Registration** — `cmd/agent/main.go`: construct, `Register`, and
   `Subscribe` the domain key to the reporter.
4. **Analyzer** — `internal/server/analyzer/<name>/`: subscribes to the
   domain's records on the server bus (the Analyzer path lands with the walking
   skeleton).
5. **Gates green**: `bash scripts/check.sh`. The layering gate auto-covers the
   new packages; a red line means the bundle reached across the seam — fix the
   code, and the gate stays.

Records carry what survived in-kernel reduction; the wire budget is 10^5
steady / 10^6 burst events/s/host (issue 0001). Placeholder records
(`cpu.Tick`) prove the chain and are replaced when the probes land.
