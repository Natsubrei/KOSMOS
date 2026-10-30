// Package cpu holds the shared record types of the CPU-scheduling collection
// domain. It is pure types and constants: zero implementation, no module
// dependencies.
package cpu

// Domain keys every CPU-domain record on the record bus.
const Domain = "cpu"

// Tick is a placeholder record proving the collection chain end to end; the
// scheduling records replace it once the eBPF probes land.
type Tick struct {
	// Seq increments once per emitted tick within the agent process.
	Seq int64
	// Pid, Tid, and Starttime name the emitting task so a Report can localize.
	Pid       int32
	Tid       int32
	Starttime uint64
}
