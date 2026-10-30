// Package diagnose is the server-side diagnosis kit: it cuts Windows and
// runs registered Analyzers. The kit knows records carry a domain, never
// any particular one; new diagnosis output is a package plus one
// registration line in cmd/server.
package diagnose

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/store"
)

// Tile is the one Window length: adjacent 10-second Unix-UTC intervals.
const Tile = 10 * time.Second

// Analyzer is the collection-domain seam on the server. One already-cut
// Window of that domain's Records in, zero or more complete Reports out.
// Implementations are compile-time packages. Bounds and required sources are
// that package's data and code, not this interface.
type Analyzer interface {
	// Name identifies this analyzer on Report.Analyzer, e.g. "cpu".
	Name() string
	// Domain is the record domain this analyzer reads.
	Domain() string
	// Analyze must not change w. Empty records may yield no Report.
	Analyze(w Window, records []bus.Record) []Report
}

// Loop is the single diagnosis consumer. Each Tick drains every closed
// undiagnosed Window, runs every Analyzer, and marks the interval diagnosed
// even when no Report is produced.
type Loop struct {
	st  store.Store
	az  []Analyzer
	now func() time.Time

	mu        sync.Mutex
	diagnosed map[string]struct{}
	reports   []Report
}

// New returns a loop reading records from st. now defaults to time.Now.
func New(st store.Store, analyzers []Analyzer) *Loop {
	return &Loop{
		st:        st,
		az:        analyzers,
		now:       time.Now,
		diagnosed: make(map[string]struct{}),
	}
}

// Run ticks once a second until ctx is done.
func (l *Loop) Run(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	l.Tick()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.Tick()
		}
	}
}

// Tick drains every closed undiagnosed Window. Tests call it with a frozen now.
func (l *Loop) Tick() {
	now := l.now()
	recs := l.st.List()
	for _, w := range l.dueWindows(recs, now) {
		var produced []Report
		for _, az := range l.az {
			produced = append(produced, az.Analyze(w, filter(recs, w, az.Domain()))...)
		}
		l.mu.Lock()
		l.diagnosed[w.ID] = struct{}{}
		l.reports = append(l.reports, produced...)
		l.mu.Unlock()
	}
}

// ListReports returns a copy of the Reports produced so far.
func (l *Loop) ListReports() []Report {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Report, len(l.reports))
	copy(out, l.reports)
	return out
}

func (l *Loop) dueWindows(recs []bus.Record, now time.Time) []Window {
	lastEnd := tileStart(now)
	hosts := hostFirst(recs)
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []Window
	for host, first := range hosts {
		for start := tileStart(first); start.Before(lastEnd); start = start.Add(Tile) {
			end := start.Add(Tile)
			id := windowID(host, start, end)
			if _, done := l.diagnosed[id]; done {
				continue
			}
			out = append(out, Window{ID: id, Host: host, Start: start, End: end})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Start.Equal(out[j].Start) {
			return out[i].Start.Before(out[j].Start)
		}
		return out[i].Host < out[j].Host
	})
	return out
}

func filter(recs []bus.Record, w Window, domain string) []bus.Record {
	var out []bus.Record
	for _, r := range recs {
		if r.Host != w.Host || r.Domain != domain {
			continue
		}
		if r.At.Before(w.Start) || !r.At.Before(w.End) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func hostFirst(recs []bus.Record) map[string]time.Time {
	first := make(map[string]time.Time)
	for _, r := range recs {
		if r.Host == "" {
			continue
		}
		if t, ok := first[r.Host]; !ok || r.At.Before(t) {
			first[r.Host] = r.At
		}
	}
	return first
}

func tileStart(t time.Time) time.Time {
	u := t.UTC().Unix()
	if u < 0 {
		u -= 9
	}
	return time.Unix(u/10*10, 0).UTC()
}

func windowID(host string, start, end time.Time) string {
	return host + "/" + strconv.FormatInt(start.Unix(), 10) + "/" + strconv.FormatInt(end.Unix(), 10)
}
