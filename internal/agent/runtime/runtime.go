// Package runtime is the agent-side runtime kit: it owns the collector
// lifecycle and hands each bundle a Handle as its only way to reach the host
// process. The kit knows no domains.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

// Collector is the domain-bundle seam: everything a collection domain plugs
// into the agent process. Implementations are self-contained and reach host
// services only through the Handle Start receives.
type Collector interface {
	// Domain keys the records this bundle publishes, e.g. contracts/cpu.Domain.
	Domain() string
	// Start prepares resources and begins collection. The handle is valid
	// until Stop returns.
	Start(h *Handle) error
	// Stop ends collection, drains in-flight work, and releases resources,
	// leaving no goroutine behind.
	Stop() error
}

// Handle is the runtime kit's service face toward one running bundle: the
// only way a bundle touches the host process.
type Handle struct {
	domain string
	bus    *bus.Bus
	now    func() time.Time
}

// Emit stamps payload with the bundle's domain and wall-clock time, then
// publishes it on the record bus.
func (h *Handle) Emit(payload any) {
	h.bus.Publish(bus.Record{Domain: h.domain, At: h.now(), Payload: payload})
}

// Agent runs a fixed set of collectors under one lifecycle; it is single-use.
type Agent struct {
	bus *bus.Bus

	started  bool
	cols     []Collector
	startedC []Collector
}

// New returns an agent publishing on b.
func New(b *bus.Bus) *Agent {
	return &Agent{bus: b}
}

// Register adds a collector. A second collector for the same domain, or a
// Register after Run, is an error.
func (a *Agent) Register(c Collector) error {
	if a.started {
		return errors.New("runtime: Register after Run")
	}
	for _, have := range a.cols {
		if have.Domain() == c.Domain() {
			return fmt.Errorf("runtime: duplicate collector for domain %q", c.Domain())
		}
	}
	a.cols = append(a.cols, c)
	return nil
}

// Run starts every registered collector in order, blocks until ctx is done,
// then stops them in reverse start order. A Start failure stops what already
// started (reverse order) and returns wrapped; a Stop failure does not stop
// the remaining Stops — all failures are joined into the return value.
func (a *Agent) Run(ctx context.Context) error {
	a.started = true
	for _, c := range a.cols {
		if err := c.Start(&Handle{domain: c.Domain(), bus: a.bus, now: time.Now}); err != nil {
			stopErr := a.stopStarted()
			return fmt.Errorf("runtime: start %s: %w (stop: %v)", c.Domain(), err, stopErr)
		}
		a.startedC = append(a.startedC, c)
	}
	<-ctx.Done()
	return a.stopStarted()
}

func (a *Agent) stopStarted() error {
	var errs []error
	for i := len(a.startedC) - 1; i >= 0; i-- {
		if err := a.startedC[i].Stop(); err != nil {
			errs = append(errs, fmt.Errorf("runtime: stop %s: %w", a.startedC[i].Domain(), err))
		}
	}
	a.startedC = nil
	return errors.Join(errs...)
}
