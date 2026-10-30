# KOSMOS architecture

English | [中文](architecture.zh.md)

What the code is, now. Decisions and their reasons live in
[Agent Notes](../.agents/notes/README.md); change one after reading the note that
owns it.

## Two processes

`cmd/agent` runs on the diagnosed machine and collects. `cmd/server` receives and
diagnoses. They share one Go module, `github.com/natsubrei/kosmos`, and nothing
else: no linked packages across the process boundary, only the wire.

## Agent

`internal/bus` is the in-process record bus. It fans a Record out to every
subscriber of that Record's `Domain`, synchronously, on the publisher's goroutine.

`internal/agent/runtime` is the runtime kit. It owns the collector lifecycle and
hands each bundle a `Handle`, the bundle's only way to publish. `Register` records
a collector; `Apply(domains)` is the snapshot that starts and stops them; `Run`
holds the process lifetime.

`contracts/<domain>` holds the record types and the domain key, and nothing else.
`internal/agent/domain/<domain>` is the bundle that produces them.

`internal/agent/hostid` is the one place the machine identity is chosen. Every
Record, and the control session, carries that string.

`internal/agent/upload` ships Records with `POST /records`.
`internal/agent/control` dials `POST /control`, receives domain snapshots, and
calls `Apply`; a dead session stops collection after a grace period.

## Server

`internal/server/httpapi` turns request bytes into Records and publishes them;
`GET /records` lists what the store kept; `GET /reports` lists what diagnosis
produced. `internal/server/control` holds one session per `host` and sends the
domain snapshot that host should run. `internal/server/store` keeps Records in
process memory.

`internal/server/diagnose` is the diagnosis kit. It cuts 10-second Unix-UTC
Windows and calls every registered `Analyzer`; an Analyzer is one Window of one
domain's Records in, zero or more complete Reports out.
`internal/server/analyzer/<domain>` is the package that implements that.

`cmd/server` is the only place that wires a domain key to the record bus, the
control snapshot, and an Analyzer, the way `cmd/agent` is the only place that
registers a bundle.

## Seams

Each of these is the single place a variable concern is allowed to appear.

| Concern | Seam | Package |
|---|---|---|
| Wire format and transport | subscribers and receivers | `internal/agent/upload`, `internal/agent/control`, `internal/server/httpapi`, `internal/server/control` |
| Record retention | `store.Store` | `internal/server/store` |
| Collection lifecycle | `runtime.Collector`, `Agent.Apply` | `internal/agent/runtime` |
| Diagnosis | `diagnose.Analyzer`, `Loop` | `internal/server/diagnose` |
| Which record kinds and analyzers exist | the domain key plus a registration line | `contracts/`, `internal/server/analyzer/<domain>`, `cmd/` |

The record bus, the runtime kit, the diagnosis kit, and `contracts/` import no
transport and no domain. `scripts/check-layering.sh` enforces the domain half of
that.

## Adding a collection domain

Follow [new-collection-domain](../.agents/skills/new-collection-domain/SKILL.md).
It lists the five places a domain touches and the gate that keeps the rest
untouched.

## Where decisions live

[Agent Notes](../.agents/notes/README.md) hold the rationale: what was chosen,
what it cost, and what it beat. Read the owning note before changing a seam; a
note under `proposed/` is decided but not yet in this code.
