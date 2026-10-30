// Package cpu is the CPU-scheduling Analyzer: one Window of cpu Records in,
// zero or more complete Reports out. Bounds are constants until the
// scenario/threshold work replaces them.
package cpu

import (
	"fmt"
	"sort"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
)

const (
	name      = "cpu"
	kind      = "cpu.on-cpu"
	evKind    = "cpu.on-cpu-ns"
	defaultNs = uint64(1e9) // 1s of on-CPU time in the Window
)

// Analyzer emits one Report per task whose summed OnCPUNs in the Window
// exceeds the bound. Evidence cites the Window and keeps the sum in payload.
type Analyzer struct {
	now   func() time.Time
	bound uint64
}

// New returns the CPU analyzer.
func New() *Analyzer {
	return &Analyzer{now: time.Now, bound: defaultNs}
}

// Name implements diagnose.Analyzer.
func (a *Analyzer) Name() string { return name }

// Domain implements diagnose.Analyzer.
func (a *Analyzer) Domain() string { return cpucontracts.Domain }

// Analyze implements diagnose.Analyzer.
func (a *Analyzer) Analyze(w diagnose.Window, records []bus.Record) []diagnose.Report {
	// diagnose.Subject is the identity contract; grouping by any other shape
	// would let the key drift from the Subject the Report claims.
	sum := make(map[diagnose.Subject]uint64)
	for _, r := range records {
		t, ok := decodeRuntime(r.Payload)
		if !ok || t.Pid == 0 {
			continue
		}
		k := diagnose.Subject{Host: w.Host, Pid: t.Pid, Tid: t.Tid, Starttime: t.Starttime}
		sum[k] += t.OnCPUNs
	}
	keys := make([]diagnose.Subject, 0, len(sum))
	for k, ns := range sum {
		if ns > a.bound {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Pid != keys[j].Pid {
			return keys[i].Pid < keys[j].Pid
		}
		if keys[i].Tid != keys[j].Tid {
			return keys[i].Tid < keys[j].Tid
		}
		return keys[i].Starttime < keys[j].Starttime
	})
	out := make([]diagnose.Report, 0, len(keys))
	for _, k := range keys {
		ns := float64(sum[k])
		out = append(out, diagnose.Report{
			ID:         fmt.Sprintf("%s/%s/%d/%d/%d", w.ID, name, k.Pid, k.Tid, k.Starttime),
			ProducedAt: a.now(),
			Host:       w.Host,
			Analyzer:   name,
			Window:     w,
			Conclusion: diagnose.Conclusion{
				Kind:    kind,
				Subject: k,
				Stance:  diagnose.StanceHolds,
			},
			Evidence: []diagnose.Evidence{{
				ID:      "e1",
				Kind:    evKind,
				Payload: map[string]any{"on_cpu_ns": ns},
				Predicate: &diagnose.Predicate{
					Observed: ns,
					Operator: diagnose.OpGT,
					Bound:    float64(a.bound),
				},
				Citations: []diagnose.Citation{{
					Window: &diagnose.WindowCite{ID: w.ID, Start: w.Start, End: w.End},
				}},
			}},
		})
	}
	return out
}

func decodeRuntime(payload any) (cpucontracts.TaskRuntime, bool) {
	switch p := payload.(type) {
	case cpucontracts.TaskRuntime:
		return p, true
	case map[string]any:
		var t cpucontracts.TaskRuntime
		if v, ok := asInt64(p["Pid"]); ok {
			t.Pid = int32(v)
		}
		if v, ok := asInt64(p["Tid"]); ok {
			t.Tid = int32(v)
		}
		if v, ok := asUint64(p["Starttime"]); ok {
			t.Starttime = v
		}
		if v, ok := asUint64(p["OnCPUNs"]); ok {
			t.OnCPUNs = v
		}
		return t, true
	default:
		return cpucontracts.TaskRuntime{}, false
	}
}

func asInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int32:
		return int64(n), true
	case int:
		return int64(n), true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

func asUint64(v any) (uint64, bool) {
	i, ok := asInt64(v)
	if !ok || i < 0 {
		return 0, false
	}
	return uint64(i), true
}

var _ diagnose.Analyzer = (*Analyzer)(nil)
