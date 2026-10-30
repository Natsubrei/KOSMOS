//go:build linux

package cpu

import "testing"

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
	if _, ok := spec.Programs[cpuProgOnSchedSwitch]; !ok {
		t.Fatalf("missing program %q", cpuProgOnSchedSwitch)
	}
	for _, m := range []string{cpuMapRuntimeNs, cpuMapLastTs} {
		if _, ok := spec.Maps[m]; !ok {
			t.Fatalf("missing map %q", m)
		}
	}
}
