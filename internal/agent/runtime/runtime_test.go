package runtime

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

type fakeCollector struct {
	domain   string
	startErr error
	mu       *sync.Mutex
	events   *[]string
}

func (f fakeCollector) Domain() string { return f.domain }

func (f fakeCollector) Start(*Handle) error {
	if f.startErr != nil {
		return f.startErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	*f.events = append(*f.events, "start:"+f.domain)
	return nil
}

func (f fakeCollector) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	*f.events = append(*f.events, "stop:"+f.domain)
	return nil
}

func joinEvents(mu *sync.Mutex, events *[]string) string {
	mu.Lock()
	defer mu.Unlock()
	return strings.Join(*events, ",")
}

func TestRunStartsNothingUntilApply(t *testing.T) {
	var mu sync.Mutex
	var events []string
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	if got := joinEvents(&mu, &events); got != "" {
		t.Fatalf("events = %q, want none until Apply", got)
	}
	cancel()
	waitRun(t, done)
	if got := joinEvents(&mu, &events); got != "" {
		t.Fatalf("events = %q, want none", got)
	}
}

func waitRun(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return")
	}
}

func TestApplyThenRunStartsAndStopsInOrder(t *testing.T) {
	var mu sync.Mutex
	var events []string
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(fakeCollector{domain: "two", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Apply([]string{"one", "two"}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	cancel()
	waitRun(t, done)

	want := "start:one,start:two,stop:two,stop:one"
	if got := joinEvents(&mu, &events); got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func waitEvents(t *testing.T, mu *sync.Mutex, events *[]string, want string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if joinEvents(mu, events) == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("events = %q, want %q", joinEvents(mu, events), want)
}

func TestApplyWhileRunningReconciles(t *testing.T) {
	var mu sync.Mutex
	var events []string
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(fakeCollector{domain: "two", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	deadline := time.Now().Add(time.Second)
	for {
		if err := a.Apply([]string{"one"}); err != nil {
			t.Fatal(err)
		}
		if joinEvents(&mu, &events) == "start:one" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("events = %q, want start:one", joinEvents(&mu, &events))
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := a.Apply([]string{"one", "two", "nope"}); err != nil {
		t.Fatal(err)
	}
	waitEvents(t, &mu, &events, "start:one,start:two")
	if err := a.Apply(nil); err != nil {
		t.Fatal(err)
	}
	waitEvents(t, &mu, &events, "start:one,start:two,stop:two,stop:one")
	cancel()
	waitRun(t, done)
}

func TestRunStartFailureFailsFast(t *testing.T) {
	var mu sync.Mutex
	var events []string
	boom := errors.New("boom")
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", mu: &mu, events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(fakeCollector{domain: "two", mu: &mu, events: &events, startErr: boom}); err != nil {
		t.Fatal(err)
	}
	if err := a.Apply([]string{"one", "two"}); err != nil {
		t.Fatal(err)
	}

	err := a.Run(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
	if got, want := joinEvents(&mu, &events), "start:one,stop:one"; got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestApplyAfterRunErrors(t *testing.T) {
	a := New(bus.New(), "web-1")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	cancel()
	waitRun(t, done)
	if err := a.Apply([]string{"cpu"}); err == nil {
		t.Fatal("expected Apply after Run error")
	}
}

func TestRegisterDuplicateDomain(t *testing.T) {
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "cpu", events: &[]string{}}); err != nil {
		t.Fatal(err)
	}
	err := a.Register(fakeCollector{domain: "cpu", events: &[]string{}})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err = %v, want duplicate-domain error", err)
	}
}

func TestHandleEmitDropsEmptyHost(t *testing.T) {
	b := bus.New()
	ch := make(chan bus.Record, 1)
	b.Subscribe("cpu", bus.Func(func(rec bus.Record) { ch <- rec }))

	h := &Handle{domain: "cpu", bus: b, now: time.Now}
	h.Emit("payload")

	select {
	case rec := <-ch:
		t.Fatalf("published %+v, want drop", rec)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHandleEmitStampsDomainClockAndHost(t *testing.T) {
	b := bus.New()
	ch := make(chan bus.Record, 4)
	b.Subscribe("cpu", bus.Func(func(rec bus.Record) { ch <- rec }))

	stamp := time.Unix(1700000000, 0)
	h := &Handle{domain: "cpu", host: "web-1", bus: b, now: func() time.Time { return stamp }}
	h.Emit("payload")

	select {
	case rec := <-ch:
		if rec.Domain != "cpu" || rec.Host != "web-1" || !rec.At.Equal(stamp) || rec.Payload != "payload" {
			t.Fatalf("record = %+v, want host web-1 domain cpu at %v payload \"payload\"", rec, stamp)
		}
	case <-time.After(time.Second):
		t.Fatal("record not delivered")
	}
}
