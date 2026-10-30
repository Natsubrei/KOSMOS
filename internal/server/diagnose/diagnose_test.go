package diagnose

import (
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/store"
)

type fake struct {
	name, domain string
	calls        int
	last         Window
	nrec         int
	out          []Report
}

func (f *fake) Name() string   { return f.name }
func (f *fake) Domain() string { return f.domain }
func (f *fake) Analyze(w Window, records []bus.Record) []Report {
	f.calls++
	f.last = w
	f.nrec = len(records)
	if len(records) == 0 {
		return nil
	}
	return f.out
}

func TestTickCutsOldestClosedAndMarksEmpty(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(5, 0).UTC(), Payload: "a"})
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(25, 0).UTC(), Payload: "b"})

	az := &fake{name: "cpu", domain: "cpu"}
	l := New(st, []Analyzer{az})
	l.now = func() time.Time { return time.Unix(40, 0).UTC() }

	l.Tick()
	if az.calls != 1 {
		t.Fatalf("calls %d, want 1", az.calls)
	}
	if az.last.Host != "web-1" || az.last.Start.Unix() != 0 || az.last.End.Unix() != 10 {
		t.Fatalf("window %+v", az.last)
	}
	if az.nrec != 1 {
		t.Fatalf("records %d, want 1 in [0,10)", az.nrec)
	}

	l.Tick() // [10,20) empty
	if az.calls != 2 || az.nrec != 0 || az.last.Start.Unix() != 10 {
		t.Fatalf("empty tile: calls=%d nrec=%d start=%d", az.calls, az.nrec, az.last.Start.Unix())
	}

	l.Tick() // [20,30) has the second record
	if az.nrec != 1 || az.last.Start.Unix() != 20 {
		t.Fatalf("third tile %+v nrec=%d", az.last, az.nrec)
	}

	l.Tick() // [30,40) empty, last closed
	if az.last.Start.Unix() != 30 {
		t.Fatalf("fourth tile %+v", az.last)
	}
	l.Tick() // nothing left
	if az.calls != 4 {
		t.Fatalf("calls %d after drain, want 4", az.calls)
	}
}

func TestTickDoesNotCutOpenWindow(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	az := &fake{name: "cpu", domain: "cpu"}
	l := New(st, []Analyzer{az})
	l.now = func() time.Time { return time.Unix(9, 0).UTC() }
	l.Tick()
	if az.calls != 0 {
		t.Fatalf("cut open window, calls %d", az.calls)
	}
}

func TestTickDoesNotRecut(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	az := &fake{name: "cpu", domain: "cpu", out: []Report{{ID: "r1"}}}
	l := New(st, []Analyzer{az})
	l.now = func() time.Time { return time.Unix(20, 0).UTC() }
	l.Tick()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(6, 0).UTC()}) // late
	l.Tick()
	if az.calls != 2 { // [0,10) then [10,20)
		t.Fatalf("calls %d", az.calls)
	}
	if n := len(l.ListReports()); n != 1 {
		t.Fatalf("reports %d, late record must not add a second report for [0,10)", n)
	}
}

func TestTickFiltersByDomainAndRunsEveryAnalyzer(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	st.Append(bus.Record{Host: "web-1", Domain: "io", At: time.Unix(6, 0).UTC()})
	cpuAz := &fake{name: "cpu", domain: "cpu"}
	ioAz := &fake{name: "io", domain: "io"}
	l := New(st, []Analyzer{cpuAz, ioAz})
	l.now = func() time.Time { return time.Unix(20, 0).UTC() }
	l.Tick()
	if cpuAz.nrec != 1 || ioAz.nrec != 1 || cpuAz.calls != 1 || ioAz.calls != 1 {
		t.Fatalf("cpu nrec=%d calls=%d io nrec=%d calls=%d", cpuAz.nrec, cpuAz.calls, ioAz.nrec, ioAz.calls)
	}
}

func TestListReportsCopy(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "web-1", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	az := &fake{name: "cpu", domain: "cpu", out: []Report{{ID: "r1"}}}
	l := New(st, []Analyzer{az})
	l.now = func() time.Time { return time.Unix(20, 0).UTC() }
	l.Tick()
	got := l.ListReports()
	got[0].ID = "mutated"
	if l.ListReports()[0].ID != "r1" {
		t.Fatal("ListReports must return a copy")
	}
}

func TestWindowIDStable(t *testing.T) {
	a := windowID("web-1", time.Unix(10, 0).UTC(), time.Unix(20, 0).UTC())
	b := windowID("web-1", time.Unix(10, 0).UTC(), time.Unix(20, 0).UTC())
	if a != b || a != "web-1/10/20" {
		t.Fatalf("id %q", a)
	}
}
