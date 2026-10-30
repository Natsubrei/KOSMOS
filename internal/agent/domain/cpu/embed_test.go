//go:build linux

package cpu

import (
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/btf"
)

// The embedded bytecode is the only copy the agent ever loads, and the C source
// is not compiled at build time. This test parses the embed the way the loader
// will, so a stale or corrupt object fails here instead of on a target host.
// It does not prove the object matches bpf/sched.bpf.c; regenerating is still
// the author's job after editing the C.
func TestEmbeddedSpecHasProgramAndMaps(t *testing.T) {
	spec, err := loadCpu()
	if err != nil {
		t.Fatalf("load embed: %v", err)
	}
	p, ok := spec.Programs[cpuProgOnSchedSwitch]
	if !ok {
		t.Fatalf("missing program %q", cpuProgOnSchedSwitch)
	}
	if p.Type != ebpf.Tracing {
		t.Fatalf("program type %s, want Tracing", p.Type)
	}
	if p.AttachTo != "sched_switch" {
		t.Fatalf("attach to %q, want sched_switch", p.AttachTo)
	}
	relocs := 0
	for i := range p.Instructions {
		if btf.CORERelocationMetadata(&p.Instructions[i]) != nil {
			relocs++
		}
	}
	if relocs != 3 {
		t.Fatalf("CO-RE relocs %d, want 3 (pid, tgid, start_time)", relocs)
	}
	m, ok := spec.Maps[cpuMapRuntimeNs]
	if !ok {
		t.Fatalf("missing map %q", cpuMapRuntimeNs)
	}
	if m.KeySize != 16 || m.ValueSize != 8 {
		t.Fatalf("runtime_ns key=%d val=%d, want 16 and 8", m.KeySize, m.ValueSize)
	}
	if _, ok := spec.Maps[cpuMapLastTs]; !ok {
		t.Fatalf("missing map %q", cpuMapLastTs)
	}
}
