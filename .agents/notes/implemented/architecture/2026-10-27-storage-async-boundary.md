# Agent Note: In-memory store and the ingest/diagnosis boundary

Status: implemented

English | [中文](2026-10-27-storage-async-boundary.zh.md)

## Problem

Diagnosis needs Records grouped by host into closed Windows, then Reports
that remain readable without joining live ingest. The receive path already
keeps Records in process memory. A database, a second consumer model, or a
staged processing pipeline would add surface before a single test guest has
produced volume.

## Decision

Records, Reports, and the set of diagnosed intervals live in the server
process. Restart loses them. There is no PostgreSQL, SQLite, or Timescale
in this cut, and no vacant store adapter. There is no rollover: retention
is the process lifetime.

A **Record** carries `host`, the machine it was collected on, stamped by
the agent at emit and sent with `POST /records`. The server does not infer
host from the TCP peer. Windows and the diagnosed set are keyed by this
`host`.

`POST /records` only appends. Diagnosis is a separate in-process ticker
loop with one consumer: each Tick drains every closed, not-yet-diagnosed
Window (`end` not after the server's wall clock), runs Analyzers, and
marks `(host, start, end)` diagnosed even when no Report is produced.
It does not run on each ingest, and it does not expose a diagnose HTTP
path.

There is no Window table. A Window is computed at cut time and embedded
in the Report. Late Records (whose `at` falls in an already-diagnosed
interval) are still appended; existing Reports do not change; the loop
does not cut that interval again.

Ingest is bus fan-out into the memory store. Diagnosis is that later
single consumer. It is not a multi-stage pipeline.

## Alternatives considered

- **PostgreSQL now:** rejected — one server and tens of agents do not
  justify another process; a vacant adapter is debt.
- **SQLite now:** rejected for this cut — restart survival waits on
  volume from the walking skeleton; SQLite is the first candidate then,
  not PostgreSQL.
- **TimescaleDB-class store:** rejected — time-series machinery before
  there is a time-series workload.
- **Diagnose inside `POST /records`:** rejected — `Deliver` must return
  quickly; diagnosis must not block ingest.
- **Several diagnosis consumers:** rejected — two cuts of the same
  `(host, start, end)` are two Reports with no order.
- **Only the latest closed Window:** rejected — a synthetic incident can
  sit in an older tile.
- **A stored Window collection:** rejected — identity is
  `(host, start, end)`; the Report already snapshots it.
- **Re-run a Window when a late Record arrives:** rejected — a later run
  is a different Report; this loop diagnoses each interval once.
- **Infer host from the TCP peer:** rejected — NAT makes the peer the
  wrong machine.
- **Call the path a pipeline:** rejected — the bus fans out; diagnosis
  is one later consumer, not a chain of stages.

## Consequences

`internal/server/store` keeps Records. `internal/server/diagnose.Loop`
keeps Reports and the diagnosed set. `GET /reports` lists Reports.
`GET /records` remains receive-and-list. How Conclusions are derived is
still outside this note. SQLite waits on volume.
