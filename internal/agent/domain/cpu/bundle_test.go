//go:build linux

package cpu

import (
	"context"
	"sync"
	"testing"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
	"github.com/natsubrei/kosmos/internal/bus"
)

type fakeProbe struct {
	samples []cpucontracts.TaskRuntime
	n       int
	closed  bool
}

func (f *fakeProbe) open() error  { return nil }
func (f *fakeProbe) close() error { f.closed = true; return nil }
func (f *fakeProbe) snapshot() ([]cpucontracts.TaskRuntime, error) {
	if len(f.samples) == 0 {
		return nil, nil
	}
	s := f.samples[f.n%len(f.samples)]
	f.n++
	return []cpucontracts.TaskRuntime{s}, nil
}

func TestNewRejectsNonPositiveInterval(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Second} {
		if _, err := New(d); err == nil {
			t.Fatalf("New(%v): expected error", d)
		}
	}
}

func TestStopBeforeStart(t *testing.T) {
	bundle, err := newBundle(time.Millisecond, &fakeProbe{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.Stop(); err == nil {
		t.Fatal("expected Stop-before-Start error")
	}
}

func TestBundleEmitsTaskRuntimeAndStops(t *testing.T) {
	b := bus.New()
	ch := make(chan bus.Record, 8)
	b.Subscribe(cpucontracts.Domain, bus.Func(func(rec bus.Record) { ch <- rec }))

	sample := cpucontracts.TaskRuntime{Pid: 7, Tid: 7, Starttime: 11, OnCPUNs: 1000}
	bundle, err := newBundle(2*time.Millisecond, &fakeProbe{samples: []cpucontracts.TaskRuntime{sample}})
	if err != nil {
		t.Fatal(err)
	}
	a := runtime.New(b, "web-1")
	if err := a.Register(bundle); err != nil {
		t.Fatal(err)
	}
	if err := a.Apply([]string{cpucontracts.Domain}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	deadline := time.After(2 * time.Second)
	n := 0
	for {
		select {
		case rec := <-ch:
			got, ok := rec.Payload.(cpucontracts.TaskRuntime)
			if !ok {
				t.Fatalf("payload = %#v, want TaskRuntime", rec.Payload)
			}
			if got != sample {
				t.Fatalf("got %+v", got)
			}
			n++
			if n >= 3 {
				cancel()
				if err := <-done; err != nil {
					t.Fatalf("Run: %v", err)
				}
				return
			}
		case <-deadline:
			t.Fatal("no TaskRuntime reached the subscriber within 2s")
		}
	}
}

// Stop must drain once more: the interval between the last tick and the stop
// would otherwise lose its on-CPU time.
func TestStopDrainsLastInterval(t *testing.T) {
	b := bus.New()
	var mu sync.Mutex
	var got int
	b.Subscribe(cpucontracts.Domain, bus.Func(func(rec bus.Record) {
		mu.Lock()
		got++
		mu.Unlock()
	}))

	sample := cpucontracts.TaskRuntime{Pid: 7, Tid: 7, Starttime: 11, OnCPUNs: 1000}
	tp := &fakeProbe{samples: []cpucontracts.TaskRuntime{sample}}
	bundle, err := newBundle(time.Hour, tp) // no tick fires before Stop
	if err != nil {
		t.Fatal(err)
	}
	a := runtime.New(b, "web-1")
	if err := a.Register(bundle); err != nil {
		t.Fatal(err)
	}
	if err := a.Apply([]string{cpucontracts.Domain}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if got != 1 {
		t.Fatalf("emitted %d records on Stop, want 1 (the final drain)", got)
	}
	if !tp.closed {
		t.Fatal("probe not closed")
	}
}
