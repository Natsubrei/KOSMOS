// Package runtime is the agent-side runtime kit: it owns the collector
// lifecycle and hands each bundle a Handle as its only way to reach the host
// process. The kit knows no domains.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	host   string
	bus    *bus.Bus
	now    func() time.Time
}

// Emit stamps payload with the bundle's domain, host, and wall-clock time,
// then publishes it on the record bus. An empty host is not published.
func (h *Handle) Emit(payload any) {
	if h.host == "" {
		return
	}
	h.bus.Publish(bus.Record{Host: h.host, Domain: h.domain, At: h.now(), Payload: payload})
}

// Agent runs a fixed set of collectors under one lifecycle; it is single-use.
// Register does not start collection. Apply reconciles running collectors to
// a domain snapshot; unknown names are ignored. Run blocks until ctx is done,
// then stops whatever is running.
type Agent struct {
	mu sync.Mutex

	bus  *bus.Bus
	host string

	cols     []Collector
	startedC []Collector
	desired  []string

	runStarted bool
	runDone    bool
}

// New returns an agent publishing on b, stamping host on every Record.
// host is the machine identity; how it is chosen is the caller's rule.
func New(b *bus.Bus, host string) *Agent {
	return &Agent{bus: b, host: host}
}

// Register adds a collector. A second collector for the same domain, or a
// Register after Run, is an error.
func (a *Agent) Register(c Collector) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.runStarted {
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

// Apply sets the domains that should be collecting. Names with no registered
// collector are ignored. Before Run, the snapshot is kept and applied when
// Run starts. After Run returns, Apply is an error.
func (a *Agent) Apply(domains []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.runDone {
		return errors.New("runtime: Apply after Run")
	}
	a.desired = append([]string(nil), domains...)
	if !a.runStarted {
		return nil
	}
	return a.reconcileLocked()
}

// Run applies the current snapshot (empty means nothing is collecting),
// blocks until ctx is done, then stops whatever is running. A Start failure
// during the first reconcile stops what already started and returns; it does
// not wait for ctx.
func (a *Agent) Run(ctx context.Context) error {
	a.mu.Lock()
	a.runStarted = true
	if err := a.reconcileLocked(); err != nil {
		a.runDone = true
		stopErr := a.stopStarted()
		a.mu.Unlock()
		return fmt.Errorf("%w (stop: %v)", err, stopErr)
	}
	a.mu.Unlock()
	<-ctx.Done()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runDone = true
	return a.stopStarted()
}

func (a *Agent) reconcileLocked() error {
	want := make(map[string]bool, len(a.desired))
	for _, d := range a.desired {
		want[d] = true
	}

	for i := len(a.startedC) - 1; i >= 0; i-- {
		c := a.startedC[i]
		if want[c.Domain()] {
			continue
		}
		if err := c.Stop(); err != nil {
			return fmt.Errorf("runtime: stop %s: %w", c.Domain(), err)
		}
		a.startedC = append(a.startedC[:i], a.startedC[i+1:]...)
	}

	running := make(map[string]bool, len(a.startedC))
	for _, c := range a.startedC {
		running[c.Domain()] = true
	}
	for _, c := range a.cols {
		if !want[c.Domain()] || running[c.Domain()] {
			continue
		}
		if err := c.Start(&Handle{domain: c.Domain(), host: a.host, bus: a.bus, now: time.Now}); err != nil {
			return fmt.Errorf("runtime: start %s: %w", c.Domain(), err)
		}
		a.startedC = append(a.startedC, c)
	}
	return nil
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
