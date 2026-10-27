# Agent Note: Time-window cut on one host

Status: proposed

English | [中文](2026-10-26-time-window-semantics.zh.md)

## Problem

A Report must cover a Window, and Evidence may rest on that Window as a
snapshot `{id, start, end}`. Without a single cut, each Analyzer would pick
its own interval, and two Reports could not mean the same span.

## Proposal

A **Window** is `{id, host, start, end}` on one host. `start` and `end` use
the same wall clock as `Record.at` (stamped at emit); `end` is exclusive.
Domain-free server code cuts Windows. The agent emits Records only. The
Analyzer receives one already-cut Window and does not change `start` or
`end`.

Records do not carry a window id. The cut happens when diagnosis runs: the
cutter takes a host and a closed `[start, end)`, assigns the id, and hands
that interval's Records to every Analyzer for that host. The same
`(host, start, end)` always yields the same id; a later Report is a
different Report, not a different Window.

The timeline is adjacent 10-second intervals, Unix UTC aligned (`start` is
a multiple of 10 seconds, `end = start + 10s`). The length is one server
constant for every host and collection domain. No sliding intervals, no
incident-triggered open or close. Windows do not store parent, prev, or
next; adjacency is `end == next.start`.

Only a closed interval is cut: `end` is not after the server's wall clock
at cut time. Evidence may cite only this Report's Window, as `{id, start,
end}` embedded in the Report (not a live lookup). It does not cite
neighbouring intervals. The Analyzer still sees only this Window's Records.

## Alternatives considered

- **Agent cuts Windows:** rejected — the agent does not diagnose.
- **Analyzer chooses `start`/`end`:** rejected — that picks the interval
  that favours the claim; domain-free packages would still have to agree
  on one span for audit.
- **Incident-anchored windows (open on trigger, close on quiescence):**
  rejected for this cut — that needs a collection-domain signal; it is
  how conclusions are derived, not how the timeline is cut.
- **Sliding windows:** rejected — overlapping spans duplicate work and
  make "the same interval" undefined.
- **Stamp window id at ingest:** rejected — the agent does not know
  Windows; the receive path must not rewrite Records to attach a window.
- **Random id per diagnosis run:** rejected — two Reports on the same
  span could not name the same Window.
- **Stored parent/prev/next:** rejected — adjacent fixed intervals are
  recoverable from `start`/`end`; no nesting in this cut.
- **Per-domain or per-host length:** rejected — the cutter carries no
  domain knowledge.
- **Cite a neighbouring Window from Evidence:** rejected — that is a
  second Window of Records, against one Window in.

## Acceptance criteria

Domain-free server code cuts 10-second Unix-UTC Windows per host at diagnosis.
An Analyzer receives one already-cut Window and does not change `start` or
`end`. Late Records do not rewrite Reports. `GET /records` remains
receive-and-list.

## Risks

The cutter runs as in the storage note. Window types land with the first
Analyzer; this note does not add them. The 10-second constant can change
without changing the Window shape.

[In-memory store and the ingest/diagnosis boundary](2026-10-27-storage-async-boundary.md)
