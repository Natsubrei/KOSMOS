# Agent Note: Agent–server protocol: data path and control session

Status: implemented

English | [中文](2026-10-27-agent-server-protocol.zh.md)

## Problem

`POST /records` already carries Records as HTTP JSON. The agent sits on a
possibly-NATed host, so the server cannot dial it. Compiling a domain in is not
the same as collecting it: running every bundle wastes CPU. Without a liveness
signal, a quiet host and a dead agent look the same. The data transport will
change (gRPC is a candidate); the record bus and domain bundles must not.

## Decision

Two channels, both agent-outbound, both HTTP in this cut, no TLS.

The **data channel** remains `POST /records` (JSON array of Records, each with
`host`). HTTP failure drops the batch; the 64-slot upload buffer is the
backpressure. Records are not retried and not written to disk.

The **control session** is a long-lived connection the agent dials. The server
holds one handle per `host`. Commands are an idempotent snapshot of domains
that should run (known names `Start`, unknown names ignored; a missing name is
`Stop`). The session plus application-level keepalive is liveness: connected
means alive; no Records on the bus. Keepalive failure starts a 60-second timer;
within 60 seconds collection keeps the last snapshot; at 60 seconds the agent
`Stop`s every bundle. After that it reconnects and waits for a new snapshot.
A new session for the same `host` replaces the old handle and closes the old
connection.

Until a snapshot arrives, no bundle runs. In this cut the server sends `{cpu}`
when a session appears.

The same `host` string identifies the session and every Record. How the agent
chooses that string is not this note.

The record bus, runtime kit, and `contracts/` do not import HTTP, JSON, or
gRPC. A later transport replaces `internal/agent/upload`,
`internal/server/httpapi`, and the control-session ends, not Publish/Deliver.

## Alternatives considered

- **No command channel; process lifetime or local enable list:** rejected —
  cannot cap overhead from the server, and cannot tell a live quiet agent from
  a dead one.
- **Server dials the agent:** rejected — NAT.
- **Periodic pull instead of a session:** rejected — the server needs a
  per-agent handle to push a new snapshot without waiting for the next poll.
- **Heartbeat as a Record:** rejected — liveness is not evidence; it must not
  enter the bus, a Window, or an Analyzer.
- **Records on the control session:** rejected — evidence framing stays off
  the session; the two receivers move together if the transport changes.
- **Stop collection on any disconnect:** rejected — short flaps would tear
  down eBPF. Sixty seconds is the waste bound.
- **TLS or enrollment in this cut:** rejected — development auth remains
  `X-User-Id`; hardening is a later auth model.
- **gRPC in this cut:** rejected — the human GET is JSON; the seam is the
  subscribers and receivers, not a second stack now.

## Consequences

`internal/agent/control` dials `POST /control`; `internal/server/control`
holds one handle per `host` and sends `{cpu}` on connect. Keepalive is 10s;
collection stops 60s after the session dies. Start/Stop stays the Runtime
handle. TLS and enrollment are a later auth model.

[Record-path seams](2026-10-26-record-path-seams.md)

[In-memory store and the ingest/diagnosis boundary](../../proposed/architecture/2026-10-27-storage-async-boundary.md)

[Compile-time domain bundles](2026-10-25-compile-time-domain-bundles.md)
