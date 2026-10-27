# Agent Note: Compile-time domain bundles over a zero-branch record bus

Status: implemented

English | [中文](2026-10-25-compile-time-domain-bundles.zh.md)

## Problem

KOSMOS must absorb new collection targets continuously. If the record bus
branches on each domain, every new domain edits that bus. Runtime plugin loading
would put a toolchain or a loader security surface on the target machine, where
the agent must remain a single low-overhead static binary, and eBPF C still has
to compile in-tree.

## Decision

The extension unit is the **domain bundle**: a self-contained compile-time
package for one collection domain — eBPF C program, map schema, decoder, record
types, optional enricher — with its own maps and ring buffer, registered in a
process-local registry. The agent's record bus and the server's isomorphic
receive bus carry **zero domain branches**: adding a domain changes no line in
those packages. Cross-domain sharing goes only through
`contracts/<domain>` pure-type packages; sibling-domain imports are banned and
enforced by `scripts/check-layering.sh` over `go list -deps`.

One Go module (`github.com/natsubrei/kosmos`): `cmd/agent`, `cmd/server`,
`internal/{agent,server}`, `contracts/<domain>`. Which domains run is
server-instructed over the control session; Start/Stop stays the Runtime
handle. That swaps the driver, not this architecture.

[Agent–server protocol: data path and control session](../../proposed/architecture/2026-10-27-agent-server-protocol.md)

Non-goals: no runtime plugin loading; no config DSL or dynamic schema registry
(records are Go types); no per-domain concurrency framework (the Runtime handle
owns lifecycle, drain scheduling, and stats); no cross-domain aggregation
inside one host; no plugin marketplace, signing, or version negotiation.

## Alternatives considered

- **Runtime plugins** (hashicorp go-plugin, Go `plugin` package, WASM):
  rejected — kernel-side C still needs a compile step, target machines must not
  grow a toolchain or a loader security surface, and a single static binary is
  part of the product promise.
- **Config-activated built-ins** (no interface, everything compiled in and
  toggled): rejected — the scenario/threshold work would fork the skeleton
  instead of plugging in.
- **Compile-time registry with an interface seam** (chosen): new domain = new
  package plus one registration line; the registry seam survives if runtime
  loading is ever wanted.

## Consequences

Adding a domain is a new package plus one registration line. The layering gate
fails a sibling-domain import or a domain-free package that learns a domain.
Those packages stay domain-free. A new domain requires a rebuild of the agent binary.
