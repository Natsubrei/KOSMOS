# Agent Note: Report and evidence-chain data model

Status: implemented

English | [中文](2026-10-26-report-evidence-model.zh.md)

## Problem

A human must audit why a diagnosis follows. Surveyed eBPF products stop at
telemetry, LLM prose, or closed causal engines; none ship a deterministic
evidence chain. Without that document, `GET /records` cannot become a report.

## Decision

A **Report** is the auditable document: one **Conclusion**, the **Evidence**
it rests on, and the **Window** it covers. It is an immutable value
`{id, produced_at, host, analyzer, window, conclusion, evidence}` in
`internal/server/diagnose`. A later run is a different Report. The Report
does not record who asked.

The Conclusion is a typed claim `{kind, subject, stance}`, not prose.
Stance is holds, does not hold, or insufficient evidence. **Subject** is
the identity the claim is about: `{host, pid, tid, starttime}`;
`starttime` matches `/proc/<pid>/stat` so PID reuse does not alias. A
process-level claim leaves `tid` empty. `comm` and cgroup are annotations,
not identity. This shape is a Linux task on a host; a second shape is added
when that collection domain exists.

The Conclusion rests on at least one Evidence item. There is no separate
justification field. Evidence is `{id, kind, payload, citations}` with an
optional **Predicate** `{observed, operator, bound}`. `kind` is an opaque
string namespaced by collection domain (`cpu.…`), owned by the Analyzer
package, not by `contracts/`. Domain-free packages do not branch on kind.
Evidence ids are unique inside one Report. Today's Predicate operators are
the closed set `lt | le | gt | ge | eq` and compare numbers; a later
non-numeric comparison still uses Predicate.

Each entry in the rests-on list is a Record snapshot `{domain, at, payload}`
(the observation used, not every Record in the window), a Window
`{id, start, end}` with `end` exclusive, another Evidence id in the same
Report (a DAG, no cycles), or a **Missing source** `{source, reason}`. An
insufficient-evidence Conclusion rests on at least one Missing source.
Aggregates may cite only a Window and keep the histogram in payload.

The Analyzer's interface is one Window's Records for its domain in, zero or
more complete Reports out. Nothing besides the Analyzer assembles fragments
or merges Reports across analyzers; two claims are two Reports. Report types
belong on the server, outside `contracts/`. The agent and `contracts/<domain>`
do not import them; `contracts/` still hold only Records. The Window value
is `{id, host, start, end}`; Evidence cites `{id, start, end}`. How the
timeline is cut is
[Time-window cut on one host](2026-10-26-time-window-semantics.md).
A Report always has a Window.

## Alternatives considered

- **One problem card holding many findings:** rejected — audit is per claim;
  several claims are several Reports.
- **Prose conclusion plus attachments:** rejected — the sentence can
  disagree with the attachments; the typed claim plus Evidence is the proof.
- **Pointer-only Evidence, join live storage to read a Report:** rejected —
  the ingest store is memory and drops on restart; a Report must stay
  readable after Records are gone.
- **Domain-free packages switch on Evidence union types:** rejected — that puts collection
  domains into packages that must stay domain-free.
- **Analyzer emits fragments, something else assembles the chain:** rejected — one
  adapter would be a hypothetical seam; the Analyzer is the producer.
- **Agent-assigned global Record ids:** rejected — the agent does not
  diagnose; snapshots identify what was used.
- **Merge Reports across Analyzers into one card:** rejected — that is a
  later product surface, not this model.
- **Severity or confidence on the Conclusion:** rejected for this model —
  stance and Evidence are enough to audit; ranking is not.

## Consequences

`internal/server/diagnose` holds the types. `GET /reports` encodes stored
Reports; a human follows Conclusion → Evidence → Record snapshot or Missing
source without joining live ingest. How Conclusions are derived, and where
bounds live, is
[Conclusions derived in Analyzer Go](2026-10-30-conclusions-in-analyzer-go.md).
Severity and a problem-card aggregator are not this model.
