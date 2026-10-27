package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

type fakeCollector struct {
	domain   string
	startErr error
	events   *[]string
}

func (f fakeCollector) Domain() string { return f.domain }

func (f fakeCollector) Start(*Handle) error {
	if f.startErr != nil {
		return f.startErr
	}
	*f.events = append(*f.events, "start:"+f.domain)
	return nil
}

func (f fakeCollector) Stop() error {
	*f.events = append(*f.events, "stop:"+f.domain)
	return nil
}

func TestRunStopsCollectorsInReverseOrder(t *testing.T) {
	var events []string
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(fakeCollector{domain: "two", events: &events}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after cancel")
	}

	want := "start:one,start:two,stop:two,stop:one"
	if got := strings.Join(events, ","); got != want {
		t.Fatalf("events = %q, want %q", got, want)
	}
}

func TestRunStartFailureFailsFast(t *testing.T) {
	var events []string
	boom := errors.New("boom")
	a := New(bus.New(), "web-1")
	if err := a.Register(fakeCollector{domain: "one", events: &events}); err != nil {
		t.Fatal(err)
	}
	if err := a.Register(fakeCollector{domain: "two", events: &events, startErr: boom}); err != nil {
		t.Fatal(err)
	}

	err := a.Run(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
	if got, want := strings.Join(events, ","), "start:one,stop:one"; got != want {
		t.Fatalf("events = %q, want %q", got, want)
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
