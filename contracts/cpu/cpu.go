// Package cpu holds the shared record types of the CPU-scheduling collection
// domain. It is pure types and constants: zero implementation, no module
// dependencies.
package cpu

// Domain keys every CPU-domain record on the record bus.
const Domain = "cpu"

// TaskRuntime is one task's on-CPU time after in-kernel aggregation over
// one drain interval. Pid is the thread-group id; Tid is the thread id;
// Starttime is /proc/<tid>/stat field 22 so PID reuse does not alias.
type TaskRuntime struct {
	Pid       int32
	Tid       int32
	Starttime uint64
	OnCPUNs   uint64
}
