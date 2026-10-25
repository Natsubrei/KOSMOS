# KOSMOS

English | [中文](GLOSSARY.zh.md)

KOSMOS diagnoses anomalies on Linux virtual machines and hosts: agents collect
in-kernel evidence via eBPF; a central server runs the diagnosis. This glossary fixes
the project's own vocabulary.

## Language

**Collection domain**:
One diagnosable area of a machine, e.g. CPU scheduling. The unit of extension on both
agent and server; everything else in this glossary exists to make adding one cheap.
_Avoid_: target, probe, module

**Domain bundle**:
The self-contained implementation of a collection domain on the agent: eBPF C program,
map schema, decoder, record types, optional enricher — plus its own maps and ring
buffer. One directory, one registration line.
_Avoid_: plugin (we compile in; nothing loads at runtime)

**Collector**:
The role a domain bundle plays toward the core: the one interface the core knows —
start and stop collecting for its domain.
_Avoid_: driver, provider

**Analyzer**:
The server-side counterpart of a domain bundle: subscribes to that domain's records on
the server bus and interprets them into findings and evidence.
_Avoid_: rule engine, processor

**Record bus**:
The in-process typed-record pipeline. Each side has one (agent, server); both carry
zero domain branches — new record kinds arrive by registration, never by editing the
core.
_Avoid_: event loop, message queue

**Zero domain branches**:
The core rule (ADR 0001): no core code branches on a specific collection domain — the
core knows records carry a domain, never any particular one. Domain behavior reaches a
process only by registration, so adding a domain leaves the core directories untouched.
_Avoid_: microkernel

**Registry**:
The process-local table where bundles and analyzers register; the only wiring point
between cores and domains.
_Avoid_: plugin host, container

**Runtime handle**:
The service face the runtime kit injects into every bundle: lifecycle, clock
conversion, stats and drop accounting, eBPF loading, drain scheduling. A bundle's only
way to touch the host process.
_Avoid_: framework, context

**Contracts**:
Per-domain pure-type packages shared by agent and server (`contracts/<domain>`): record
types and constants, zero implementation.
_Avoid_: API, schema (those name the wire protocol, ticket 0006)

**Record**:
One observation a bundle publishes: payload, domain, and wall-clock time stamped at
emit. A raw kernel occurrence (one sched_switch) is a kernel event; records are what
the in-kernel aggregation produces from them.
_Avoid_: event, message

**Runtime kit**:
The host-process machinery every bundle shares through its Runtime handle: lifecycle,
clock stamping, stats, eBPF loading, drain scheduling. Carries no domain knowledge.
_Avoid_: framework, kernel

**Core**:
The domain-free part of each process: on the agent, the record bus and the runtime
kit; on the server, the shared ingest/window/report pipeline. Domains enter only by
registration.
_Avoid_: kernel, host
