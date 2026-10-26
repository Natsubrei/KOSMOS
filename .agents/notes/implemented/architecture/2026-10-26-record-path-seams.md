# Agent Note: Record-path seams for the first vertical slice

Status: implemented

English | [中文](2026-10-26-record-path-seams.zh.md)

## Problem

The agent already emitted Records on an in-process bus. The next slice had to
carry those Records to a server and let a human GET them, without welding
HTTP, JSON, or a database into domain bundles, the bus, or analyzers — those
will change. Performance cannot come from the bus: a busy host produces about
5×10⁵ `sched_switch`/s, which userspace cannot forward event-by-event.

## Decision

Each process has a Record bus (`internal/bus`) that fans out by `Domain`
(the collection-domain key, e.g. `"cpu"`), synchronously, on the publisher's
goroutine. `Deliver` must return quickly. The bus does not retain data and is
not a message queue.

The agent does not diagnose. Userspace decodes, may lightly enrich, and ships
within the protocol budget (10⁵ steady / 10⁶ burst Records/s/host).
`internal/agent/upload` is a bus subscriber: `Deliver` never blocks on the
network; a full buffer of 64 drops and counts. In-kernel aggregation remains
the volume control.

Agent to server is outbound HTTP JSON (`POST /records`, JSON array of
`{domain, at, payload}`), not an external broker and not gRPC.
`cmd/agent -server` enables upload; domain bundles do not import
`internal/agent/upload`. Contracts have no JSON tags.

`cmd/server` listens on HTTP. The handler (`internal/server/httpapi`) turns
bytes into Records and `Publish`es; `internal/server/store` Memory implements
append/list and is the bus subscriber that retains them. `GET /records`
returns that list as JSON. Development auth is a non-empty `X-User-Id`.
Restarting the server loses the list. Wiring `Subscribe(cpu)` lives in
`cmd/server`, like bundle registration in `cmd/agent`.

## Alternatives considered

- **Treat the Record bus as a message queue (Kafka/NATS/RabbitMQ):** rejected
  — the bus is in-process fan-out; an external broker is another ops surface
  at one server / tens of agents.
- **Free-form tags on Records, core switches on tags:** rejected — `Domain`
  is the collection-domain key from compiled-in bundles; subscribers register.
- **Diagnose on the agent:** rejected — footprint lives on the target
  machine; diagnosis belongs to server analyzers.
- **Unbounded userspace queues so nothing is lost:** rejected — memory
  occupancy explodes; drop and count instead.
- **gRPC + protobuf in this slice:** rejected — Tick does not need it; the
  human GET is already JSON. A later transport is another subscriber/receiver.
- **JSON tags on `contracts/` or analyzers reading HTTP bodies:** rejected —
  that welds the wire to the vocabulary.
- **PostgreSQL (or SQLite) in this slice:** rejected — Memory behind
  append/list is enough; a vacant adapter is debt.
- **GET returns an evidence-chain report:** rejected for this slice — the
  model is the [report and evidence-chain](2026-10-26-report-evidence-model.md)
  note; `GET /records` remains receive-and-list.

## Consequences

curl against `GET /records` accepts the slice. JSON will not carry 10⁵
Records/s/host: change `upload` and `httpapi`, not the cores or contracts.
Memory store does not survive restart. Reconnection is "drop on HTTP
failure"; backpressure is the 64-slot buffer. Server-instructed collection
activation is not this path.
