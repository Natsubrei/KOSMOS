//go:build linux

// Package cpu is the CPU-scheduling collection domain bundle: it aggregates
// on-CPU time in-kernel on sched_switch and emits one TaskRuntime per task
// per drain interval.
package cpu

import (
	"errors"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
)

// Bundle drains the in-kernel map at interval and emits TaskRuntime records.
type Bundle struct {
	interval time.Duration
	probe    probe

	h       *runtime.Handle
	stop    chan struct{}
	stopped chan struct{}
}

// New returns a CPU bundle that loads the eBPF program on Start.
func New(interval time.Duration) (*Bundle, error) {
	return newBundle(interval, newEBPFProbe())
}

func newBundle(interval time.Duration, p probe) (*Bundle, error) {
	if interval <= 0 {
		return nil, errors.New("cpu bundle: interval must be positive")
	}
	return &Bundle{interval: interval, probe: p}, nil
}

// Domain implements runtime.Collector.
func (b *Bundle) Domain() string { return cpucontracts.Domain }

// Start attaches the probe and launches the drain loop.
func (b *Bundle) Start(h *runtime.Handle) error {
	if err := b.probe.open(); err != nil {
		return err
	}
	b.h = h
	b.stop = make(chan struct{})
	b.stopped = make(chan struct{})
	go b.loop()
	return nil
}

// Stop ends the drain loop, drains what is left, then closes the probe.
func (b *Bundle) Stop() error {
	if b.stop == nil {
		return errors.New("cpu bundle: Stop before Start")
	}
	close(b.stop)
	<-b.stopped
	b.drain()
	b.stop = nil
	return b.probe.close()
}

func (b *Bundle) loop() {
	defer close(b.stopped)
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			b.drain()
		}
	}
}

// drain emits one Record per task seen since the previous drain.
func (b *Bundle) drain() {
	recs, err := b.probe.snapshot()
	if err != nil {
		return
	}
	for _, r := range recs {
		if r.Pid == 0 {
			continue
		}
		b.h.Emit(r)
	}
}

var _ runtime.Collector = (*Bundle)(nil)
