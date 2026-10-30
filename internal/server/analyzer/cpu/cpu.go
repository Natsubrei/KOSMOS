// Package cpu is the CPU-scheduling Analyzer: one Window of cpu Records in,
// zero or more complete Reports out. Bounds are constants until the
// scenario/threshold work replaces them.
package cpu

import (
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
)

const (
	name      = "cpu"
	kind      = "cpu.tick"
	evKind    = "cpu.tick-count"
	tickBound = 0.0
)

// Analyzer emits a Report when a Window contains at least one cpu.Tick that
// names a task. The comparison is count > 0.
type Analyzer struct {
	now func() time.Time
}

// New returns the CPU analyzer.
func New() *Analyzer {
	return &Analyzer{now: time.Now}
}

// Name implements diagnose.Analyzer.
func (a *Analyzer) Name() string { return name }

// Domain implements diagnose.Analyzer.
func (a *Analyzer) Domain() string { return cpucontracts.Domain }

// Analyze implements diagnose.Analyzer.
func (a *Analyzer) Analyze(w diagnose.Window, records []bus.Record) []diagnose.Report {
	var ticks []bus.Record
	var sub diagnose.Subject
	for _, r := range records {
		t, ok := decodeTick(r.Payload)
		if !ok || t.Pid == 0 {
			continue
		}
		if sub.Pid == 0 {
			sub = diagnose.Subject{Host: w.Host, Pid: t.Pid, Tid: t.Tid, Starttime: t.Starttime}
		}
		ticks = append(ticks, r)
	}
	if len(ticks) == 0 {
		return nil
	}
	n := float64(len(ticks))
	cites := []diagnose.Citation{{
		Window: &diagnose.WindowCite{ID: w.ID, Start: w.Start, End: w.End},
	}}
	for _, r := range ticks {
		cites = append(cites, diagnose.Citation{Record: &diagnose.RecordSnap{
			Domain: r.Domain, At: r.At, Payload: r.Payload,
		}})
	}
	return []diagnose.Report{{
		ID:         w.ID + "/" + name,
		ProducedAt: a.now(),
		Host:       w.Host,
		Analyzer:   name,
		Window:     w,
		Conclusion: diagnose.Conclusion{
			Kind:    kind,
			Subject: sub,
			Stance:  diagnose.StanceHolds,
		},
		Evidence: []diagnose.Evidence{{
			ID:      "e1",
			Kind:    evKind,
			Payload: map[string]any{"count": n},
			Predicate: &diagnose.Predicate{
				Observed: n,
				Operator: diagnose.OpGT,
				Bound:    tickBound,
			},
			Citations: cites,
		}},
	}}
}

func decodeTick(payload any) (cpucontracts.Tick, bool) {
	switch p := payload.(type) {
	case cpucontracts.Tick:
		return p, true
	case map[string]any:
		var t cpucontracts.Tick
		if v, ok := asInt64(p["Seq"]); ok {
			t.Seq = v
		}
		if v, ok := asInt64(p["Pid"]); ok {
			t.Pid = int32(v)
		}
		if v, ok := asInt64(p["Tid"]); ok {
			t.Tid = int32(v)
		}
		if v, ok := asUint64(p["Starttime"]); ok {
			t.Starttime = v
		}
		return t, true
	default:
		return cpucontracts.Tick{}, false
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
