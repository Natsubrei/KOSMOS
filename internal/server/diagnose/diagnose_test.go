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
	totalRec     int
	out          []Report
}

func (f *fake) Name() string   { return f.name }
func (f *fake) Domain() string { return f.domain }
func (f *fake) Analyze(w Window, records []bus.Record) []Report {
	f.calls++
	f.last = w
	f.nrec = len(records)
	f.totalRec += len(records)
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
	if az.calls != 4 {
		t.Fatalf("calls %d, want 4 windows in one Tick", az.calls)
	}
	if az.last.Host != "web-1" || az.last.Start.Unix() != 30 || az.last.End.Unix() != 40 {
		t.Fatalf("last window %+v", az.last)
	}
	l.Tick()
	if az.calls != 4 {
		t.Fatalf("second Tick recut, calls %d", az.calls)
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
	if az.calls != 2 { // [0,10) and [10,20) in the first Tick
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
	if cpuAz.totalRec != 1 || ioAz.totalRec != 1 || cpuAz.calls != 2 || ioAz.calls != 2 {
		t.Fatalf("cpu totalRec=%d calls=%d io totalRec=%d calls=%d", cpuAz.totalRec, cpuAz.calls, ioAz.totalRec, ioAz.calls)
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

func TestTickDrainsAllDueWindows(t *testing.T) {
	st := store.NewMemory()
	st.Append(bus.Record{Host: "a", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	st.Append(bus.Record{Host: "b", Domain: "cpu", At: time.Unix(5, 0).UTC()})
	az := &fake{name: "cpu", domain: "cpu"}
	l := New(st, []Analyzer{az})
	l.now = func() time.Time { return time.Unix(3600, 0).UTC() }
	l.Tick()
	const want = 2 * 360 // two hosts, [0,3600) in 10s tiles
	if az.calls != want {
		t.Fatalf("calls %d, want %d", az.calls, want)
	}
	l.Tick()
	if az.calls != want {
		t.Fatalf("recut, calls %d", az.calls)
	}
}

func TestWindowIDStable(t *testing.T) {
	a := windowID("web-1", time.Unix(10, 0).UTC(), time.Unix(20, 0).UTC())
	b := windowID("web-1", time.Unix(10, 0).UTC(), time.Unix(20, 0).UTC())
	if a != b || a != "web-1/10/20" {
		t.Fatalf("id %q", a)
	}
}
