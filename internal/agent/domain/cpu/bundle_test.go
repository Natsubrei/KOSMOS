package cpu

import (
	"context"
	"testing"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
	"github.com/natsubrei/kosmos/internal/bus"
)

func TestNewRejectsNonPositiveInterval(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Second} {
		if _, err := New(d); err == nil {
			t.Fatalf("New(%v): expected error", d)
		}
	}
}

func TestStopBeforeStart(t *testing.T) {
	bundle, err := New(time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.Stop(); err == nil {
		t.Fatal("expected Stop-before-Start error")
	}
}

func TestBundleEmitsTicksAndStops(t *testing.T) {
	b := bus.New()
	ch := make(chan bus.Record, 8)
	b.Subscribe(cpucontracts.Domain, bus.Func(func(rec bus.Record) { ch <- rec }))

	bundle, err := New(2 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	a := runtime.New(b, "web-1")
	if err := a.Register(bundle); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case rec := <-ch:
			tick, ok := rec.Payload.(cpucontracts.Tick)
			if !ok {
				t.Fatalf("payload = %#v, want cpucontracts.Tick", rec.Payload)
			}
			if tick.Seq >= 3 {
				cancel()
				if err := <-done; err != nil {
					t.Fatalf("Run: %v", err)
				}
				return
			}
		case <-deadline:
			t.Fatal("no ticks reached the subscriber within 2s")
		}
	}
}
