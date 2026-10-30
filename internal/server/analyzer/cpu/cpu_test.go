package cpu

import (
	"testing"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
)

func window() diagnose.Window {
	return diagnose.Window{
		ID:    "web-1/0/10",
		Host:  "web-1",
		Start: time.Unix(0, 0).UTC(),
		End:   time.Unix(10, 0).UTC(),
	}
}

func TestAnalyzeGroupsBySubjectAndCitesWindow(t *testing.T) {
	a := New()
	a.now = func() time.Time { return time.Unix(50, 0).UTC() }
	a.bound = 100
	w := window()
	recs := []bus.Record{
		{
			Host: "web-1", Domain: cpucontracts.Domain, At: time.Unix(1, 0).UTC(),
			Payload: cpucontracts.TaskRuntime{Pid: 42, Tid: 42, Starttime: 7, OnCPUNs: 80},
		},
		{
			Host: "web-1", Domain: cpucontracts.Domain, At: time.Unix(2, 0).UTC(),
			Payload: cpucontracts.TaskRuntime{Pid: 42, Tid: 42, Starttime: 7, OnCPUNs: 40},
		},
		{
			Host: "web-1", Domain: cpucontracts.Domain, At: time.Unix(3, 0).UTC(),
			Payload: cpucontracts.TaskRuntime{Pid: 9, Tid: 9, Starttime: 1, OnCPUNs: 50},
		},
	}
	got := a.Analyze(w, recs)
	if len(got) != 1 {
		t.Fatalf("reports %d, want 1 (pid 9 is under bound)", len(got))
	}
	r := got[0]
	if r.Conclusion.Kind != "cpu.on-cpu" || r.Conclusion.Stance != diagnose.StanceHolds {
		t.Fatalf("conclusion %+v", r.Conclusion)
	}
	if r.Conclusion.Subject != (diagnose.Subject{Host: "web-1", Pid: 42, Tid: 42, Starttime: 7}) {
		t.Fatalf("subject %+v", r.Conclusion.Subject)
	}
	if r.ID != "web-1/0/10/cpu/42/42/7" {
		t.Fatalf("id %q", r.ID)
	}
	if len(r.Evidence) != 1 || r.Evidence[0].Kind != "cpu.on-cpu-ns" {
		t.Fatalf("evidence %+v", r.Evidence)
	}
	p := r.Evidence[0].Predicate
	if p == nil || p.Observed != 120 || p.Operator != diagnose.OpGT || p.Bound != 100 {
		t.Fatalf("predicate %+v", p)
	}
	if len(r.Evidence[0].Citations) != 1 || r.Evidence[0].Citations[0].Window == nil {
		t.Fatalf("citations %+v", r.Evidence[0].Citations)
	}
}

func TestAnalyzeDecodesJSONPayload(t *testing.T) {
	a := New()
	a.bound = 0
	got := a.Analyze(window(), []bus.Record{{
		Host: "web-1", Domain: cpucontracts.Domain, At: time.Unix(1, 0).UTC(),
		Payload: map[string]any{
			"Pid": float64(9), "Tid": float64(9), "Starttime": float64(1), "OnCPUNs": float64(2),
		},
	}})
	if len(got) != 1 || got[0].Conclusion.Subject.Pid != 9 {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyzeSkipsEmptyAndPidZero(t *testing.T) {
	a := New()
	a.bound = 0
	w := window()
	if got := a.Analyze(w, nil); got != nil && len(got) != 0 {
		t.Fatalf("empty: %+v", got)
	}
	if got := a.Analyze(w, []bus.Record{{Payload: cpucontracts.TaskRuntime{OnCPUNs: 9}}}); len(got) != 0 {
		t.Fatalf("pid 0: %+v", got)
	}
}

func TestAnalyzeTwoSubjectsTwoReports(t *testing.T) {
	a := New()
	a.bound = 0
	got := a.Analyze(window(), []bus.Record{
		{Payload: cpucontracts.TaskRuntime{Pid: 2, Tid: 2, Starttime: 1, OnCPUNs: 1}},
		{Payload: cpucontracts.TaskRuntime{Pid: 1, Tid: 1, Starttime: 1, OnCPUNs: 1}},
	})
	if len(got) != 2 {
		t.Fatalf("reports %d", len(got))
	}
	if got[0].Conclusion.Subject.Pid != 1 || got[1].Conclusion.Subject.Pid != 2 {
		t.Fatalf("order %+v %+v", got[0].Conclusion.Subject, got[1].Conclusion.Subject)
	}
	if got[0].ID == got[1].ID {
		t.Fatalf("duplicate id %q", got[0].ID)
	}
}
