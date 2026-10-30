//go:build linux

package cpu

import (
	"testing"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
)

// The kernel key names the thread-group id tgid; the record contract calls it
// Pid and pairs it with Tid and Starttime. Swapping those silently mislabels
// every report, so the mapping is pinned here rather than inferred from the
// struct field order.
func TestTaskRuntimeMapsKeyToRecord(t *testing.T) {
	cases := []struct {
		name string
		key  cpuTaskKey
		val  cpuTaskVal
		want cpucontracts.TaskRuntime
	}{
		{
			name: "thread of a process",
			key:  cpuTaskKey{Tid: 42, Tgid: 7, StartTime: 1234},
			val:  cpuTaskVal{OnCpuNs: 99},
			want: cpucontracts.TaskRuntime{Pid: 7, Tid: 42, Starttime: 1234, OnCPUNs: 99},
		},
		{
			name: "single-threaded task",
			key:  cpuTaskKey{Tid: 7, Tgid: 7, StartTime: 1234},
			val:  cpuTaskVal{OnCpuNs: 0},
			want: cpucontracts.TaskRuntime{Pid: 7, Tid: 7, Starttime: 1234},
		},
		{
			name: "kernel thread keeps tgid equal to tid",
			key:  cpuTaskKey{Tid: 3, Tgid: 3, StartTime: 1},
			val:  cpuTaskVal{OnCpuNs: 5},
			want: cpucontracts.TaskRuntime{Pid: 3, Tid: 3, Starttime: 1, OnCPUNs: 5},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := taskRuntime(tc.key, tc.val); got != tc.want {
				t.Fatalf("taskRuntime = %+v, want %+v", got, tc.want)
			}
		})
	}
}
