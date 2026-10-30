package cpu

import (
	"testing"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
)

func TestAnalyzeHoldsOnTicks(t *testing.T) {
	a := New()
	a.now = func() time.Time { return time.Unix(50, 0).UTC() }
	w := diagnose.Window{
		ID:    "web-1/0/10",
		Host:  "web-1",
		Start: time.Unix(0, 0).UTC(),
		End:   time.Unix(10, 0).UTC(),
	}
	recs := []bus.Record{{
		Host: "web-1", Domain: cpucontracts.Domain, At: time.Unix(5, 0).UTC(),
		Payload: cpucontracts.Tick{Seq: 1, Pid: 42, Tid: 42, Starttime: 7},
	}}
	got := a.Analyze(w, recs)
	if len(got) != 1 {
		t.Fatalf("reports %d", len(got))
	}
	r := got[0]
	if r.Analyzer != "cpu" || r.Host != "web-1" || r.ID != "web-1/0/10/cpu" {
		t.Fatalf("report %+v", r)
	}
	if r.Conclusion.Kind != "cpu.tick" || r.Conclusion.Stance != diagnose.StanceHolds {
		t.Fatalf("conclusion %+v", r.Conclusion)
	}
	if r.Conclusion.Subject != (diagnose.Subject{Host: "web-1", Pid: 42, Tid: 42, Starttime: 7}) {
		t.Fatalf("subject %+v", r.Conclusion.Subject)
	}
	if len(r.Evidence) != 1 || r.Evidence[0].Kind != "cpu.tick-count" {
		t.Fatalf("evidence %+v", r.Evidence)
	}
	p := r.Evidence[0].Predicate
	if p == nil || p.Observed != 1 || p.Operator != diagnose.OpGT || p.Bound != 0 {
		t.Fatalf("predicate %+v", p)
	}
	if len(r.Evidence[0].Citations) != 2 {
		t.Fatalf("citations %d", len(r.Evidence[0].Citations))
	}
}

func TestAnalyzeDecodesJSONPayload(t *testing.T) {
	a := New()
	w := diagnose.Window{ID: "h/0/10", Host: "h", Start: time.Unix(0, 0).UTC(), End: time.Unix(10, 0).UTC()}
	got := a.Analyze(w, []bus.Record{{
		Host: "h", Domain: cpucontracts.Domain, At: time.Unix(1, 0).UTC(),
		Payload: map[string]any{"Seq": float64(3), "Pid": float64(9), "Tid": float64(9), "Starttime": float64(1)},
	}})
	if len(got) != 1 || got[0].Conclusion.Subject.Pid != 9 {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyzeSkipsEmptyAndPidZero(t *testing.T) {
	a := New()
	w := diagnose.Window{ID: "h/0/10", Host: "h", Start: time.Unix(0, 0).UTC(), End: time.Unix(10, 0).UTC()}
	if got := a.Analyze(w, nil); got != nil {
		t.Fatalf("empty: %+v", got)
	}
	if got := a.Analyze(w, []bus.Record{{Payload: cpucontracts.Tick{Seq: 1}}}); got != nil {
		t.Fatalf("pid 0: %+v", got)
	}
}
