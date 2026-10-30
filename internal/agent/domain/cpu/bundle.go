// Package cpu is the CPU-scheduling collection domain bundle. It is a stub
// for now: one placeholder record per interval proves the chain end to end;
// the eBPF probes replace the emitter once the kernel-side design lands.
package cpu

import (
	"errors"
	"os"
	"time"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/agent/runtime"
)

// Bundle emits one cpucontracts.Tick per interval.
type Bundle struct {
	interval time.Duration

	h       *runtime.Handle
	stop    chan struct{}
	stopped chan struct{}
}

// New returns a CPU bundle ticking at interval; interval must be positive.
func New(interval time.Duration) (*Bundle, error) {
	if interval <= 0 {
		return nil, errors.New("cpu bundle: interval must be positive")
	}
	return &Bundle{interval: interval}, nil
}

// Domain implements runtime.Collector.
func (b *Bundle) Domain() string { return cpucontracts.Domain }

// Start launches the tick loop.
func (b *Bundle) Start(h *runtime.Handle) error {
	b.h = h
	b.stop = make(chan struct{})
	b.stopped = make(chan struct{})
	go b.loop()
	return nil
}

// Stop ends the tick loop and waits for it to exit.
func (b *Bundle) Stop() error {
	if b.stop == nil {
		return errors.New("cpu bundle: Stop before Start")
	}
	close(b.stop)
	<-b.stopped
	b.stop = nil
	return nil
}

func (b *Bundle) loop() {
	defer close(b.stopped)
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	var seq int64
	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			seq++
			pid := int32(os.Getpid())
			b.h.Emit(cpucontracts.Tick{Seq: seq, Pid: pid, Tid: pid})
		}
	}
}

var _ runtime.Collector = (*Bundle)(nil)
