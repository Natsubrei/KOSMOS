# KOSMOS

English | [中文](glossary.zh.md)

KOSMOS diagnoses anomalies on Linux virtual machines and hosts: agents collect
in-kernel evidence via eBPF; a central server runs the diagnosis. This glossary fixes
the project's own vocabulary. How those words render in each language is
[terminology.md](i18n/terminology.md).

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
The role a domain bundle plays toward the runtime kit: start and stop collecting
for its domain. That is the only interface the runtime kit knows.
_Avoid_: driver, provider

**Analyzer**:
The server-side counterpart of a domain bundle: subscribes to that domain's Records on
the server bus and produces Reports.
_Avoid_: rule engine, processor, finding

**Record bus**:
The in-process typed-record pipeline. Each side has one (agent, server); both carry
zero domain branches — new record kinds arrive by registration, never by adding a
domain branch to the bus.
_Avoid_: event loop, message queue

**Zero domain branches**:
The rule ([compile-time domain bundles](../.agents/notes/implemented/architecture/2026-10-25-compile-time-domain-bundles.md)): domain-free packages do not branch on a specific collection domain — they know records carry a domain, never any particular one. Domain behavior reaches a process only by registration, so adding a domain leaves those packages untouched.
_Avoid_: microkernel

**Registry**:
The process-local table where bundles and analyzers register; the only wiring point
between domain-free packages and domains.
_Avoid_: plugin host, container

**Runtime handle**:
The service face the runtime kit injects into every bundle: lifecycle, clock
conversion, stats and drop accounting, eBPF loading, drain scheduling. A bundle's only
way to touch the host process.
_Avoid_: framework, context

**Contracts**:
Per-domain pure-type packages shared by agent and server (`contracts/<domain>`): record
types and constants, zero implementation.
_Avoid_: API, schema (those name the wire protocol)

**Record**:
One observation a bundle publishes: payload, domain, and wall-clock time stamped at
emit. A raw kernel occurrence (one sched_switch) is a kernel event; records are what
the in-kernel aggregation produces from them.
_Avoid_: event, message

**Runtime kit**:
The host-process machinery every bundle shares through its Runtime handle: lifecycle,
clock stamping, stats, eBPF loading, drain scheduling. Carries no domain knowledge.
_Avoid_: framework, kernel

## Diagnosis

**Report**:
The auditable document a human reads: one Conclusion, the Evidence it rests on, and
the Window it covers. Immutable; a later run is a different Report.
_Avoid_: finding, problem card, diagnosis result

**Conclusion**:
A typed claim about one Subject: a kind, the Subject, and a stance (holds, does
not hold, or insufficient evidence). Prose is generated from these, not stored as
the claim.
_Avoid_: finding, verdict, root cause

**Evidence**:
One item a Conclusion rests on: an opaque kind, a payload, and a list of what it
rests on (a Record snapshot, a Window, another Evidence in the same Report, or a
Missing source).
_Avoid_: citation, attachment, justification

**Subject**:
The typed identity a Conclusion is about: host, pid, tid, and the task start time
(so PID reuse does not alias). A process-level claim leaves tid empty. comm and
cgroup are annotations, not identity.
_Avoid_: resource, entity, target, object

**Window**:
The time interval a Report covers on one host, cited by id with a half-open start
and end on the same wall clock as Record.at.
_Avoid_: period, bucket, range

**Predicate**:
An optional comparison recorded on Evidence: the observed value, an operator, and
the bound that was used.
_Avoid_: threshold (the configured bound; the Predicate is the comparison that fired)

**Missing source**:
A required source that was absent, with a reason. An insufficient-evidence
Conclusion rests on at least one Missing source.
_Avoid_: gap, missing data, error
