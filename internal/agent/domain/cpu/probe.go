//go:build linux

package cpu

import (
	"fmt"

	"github.com/cilium/ebpf/link"
	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
)

// probe is the in-kernel collector. Tests swap in a fake; production uses eBPF.
type probe interface {
	open() error
	snapshot() ([]cpucontracts.TaskRuntime, error)
	close() error
}

type ebpfProbe struct {
	objs cpuObjects
	tp   link.Link
}

func newEBPFProbe() probe { return &ebpfProbe{} }

func (p *ebpfProbe) open() error {
	if err := loadCpuObjects(&p.objs, nil); err != nil {
		return fmt.Errorf("cpu: load bpf: %w", err)
	}
	tp, err := link.AttachTracing(link.TracingOptions{Program: p.objs.OnSchedSwitch})
	if err != nil {
		_ = p.objs.Close()
		return fmt.Errorf("cpu: attach tp_btf/sched_switch: %w", err)
	}
	p.tp = tp
	return nil
}

func (p *ebpfProbe) close() error {
	var err error
	if p.tp != nil {
		err = p.tp.Close()
		p.tp = nil
	}
	if e := p.objs.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

func (p *ebpfProbe) snapshot() ([]cpucontracts.TaskRuntime, error) {
	var keys []cpuTaskKey
	var vals []cpuTaskVal
	var k cpuTaskKey
	var v cpuTaskVal
	it := p.objs.RuntimeNs.Iterate()
	for it.Next(&k, &v) {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("cpu: iterate runtime_ns: %w", err)
	}
	out := make([]cpucontracts.TaskRuntime, 0, len(keys))
	for i, k := range keys {
		if err := p.objs.RuntimeNs.Delete(&k); err != nil {
			return nil, fmt.Errorf("cpu: delete runtime_ns: %w", err)
		}
		out = append(out, taskRuntime(k, vals[i]))
	}
	return out, nil
}

// taskRuntime maps one kernel map entry to the record contract. The kernel key
// carries the thread-group id as tgid; the contract calls it Pid.
func taskRuntime(k cpuTaskKey, v cpuTaskVal) cpucontracts.TaskRuntime {
	return cpucontracts.TaskRuntime{
		Pid:       int32(k.Tgid),
		Tid:       int32(k.Tid),
		Starttime: k.StartTime,
		OnCPUNs:   v.OnCpuNs,
	}
}
