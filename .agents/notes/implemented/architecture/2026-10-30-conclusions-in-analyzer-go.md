# Agent Note: Conclusions derived in Analyzer Go

Status: implemented

English | [中文](2026-10-30-conclusions-in-analyzer-go.zh.md)

## Problem

The report model says what a Report is, not how an Analyzer decides to produce
one. Without that, a rule engine, a config store, or a required-sources method
on the diagnosis kit all look like the next seam.

## Decision

An Analyzer package derives Conclusions in Go. There is no scenario document,
no rule language, and no interpreter. Two claims remain two Reports; nothing
merges Analyzer output.

A numeric bound is a field on the Analyzer value. The package's constructor
fills it from a constant (`cpu.New` fills `defaultNs`, one second). `Analyze`
reads the field and copies that number into `Predicate.Bound` on the Evidence
it writes. The diagnosis kit, the record bus, and `cmd/server` do not know
bounds. `diagnose.Analyzer` does not grow a bounds or required-sources method.
One Conclusion kind, one field — not a `map[kind]bound`. Where the number
comes from is the constructor's job; `Analyze` and `Loop` do not change when
that source changes.

When `Analyze` is called with no records, or the comparison does not fire, it
returns no Report. `does_not_hold` is for an Analyzer that writes a negative
Conclusion, not a census of tasks under the bound. A Missing source is written
by that Analyzer's Go when it produces a Conclusion it cannot finish. The
current `cpu.on-cpu` path needs only `TaskRuntime`, so it writes none.

## Alternatives considered

- **Runtime scenario / rule engine:** rejected — a second diagnosis language,
  and it fights compile-time domain bundles.
- **Config file, database, or HTTP API for bounds:** rejected — no such
  subsystem exists, and there is no auth model for a write API. Predicate
  already records the bound that was used.
- **`RequiredSources()` on `diagnose.Analyzer`:** rejected — the kit would be
  diagnosing emptiness; liveness is not a Record, and an empty Window is idle,
  not-collecting, and a dead agent alike.
- **Insufficient-evidence Report on every empty Window:** rejected —
  `GET /reports` would become a census of quiet hosts.
- **`does_not_hold` for every task under the bound:** rejected — the same
  census, per thread.
- **Bounds interface on the diagnosis kit:** rejected — one implementation per
  Analyzer; an empty adapter is debt. The next step is the constructor filling
  the field from somewhere else.
- **`map[kind]bound`:** rejected — a second kind adds a field when that
  Conclusion exists.

## Consequences

`internal/server/analyzer/cpu` matches: `bound` is a field, `New` fills
`defaultNs`, `Analyze` copies it into Predicate. Tests in the same package set
`a.bound`. The kit in `internal/server/diagnose` is unchanged.

Steal time and off-CPU are still not collected; a Conclusion that needs them
is a later Analyzer change, which is when Missing source would first appear
on this domain.

[Report and evidence-chain data model](2026-10-26-report-evidence-model.md)
