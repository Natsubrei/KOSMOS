// Package bus is the agent's in-process record bus: domain bundles publish
// typed records and subscribers receive them per domain. The bus carries no
// domain knowledge — new record kinds arrive by subscription, never by
// editing this package.
package bus

import (
	"sync"
	"time"
)

// Record is one collected observation. Host is the machine it was collected
// on; At is wall-clock time. Both are stamped by the runtime at emit;
// producers and consumers never touch clocks or identity.
type Record struct {
	Host    string
	Domain  string
	At      time.Time
	Payload any
}

// Subscriber receives records for the domains it subscribed to. Deliver runs
// on the publishing goroutine and must return quickly.
type Subscriber interface {
	Deliver(rec Record)
}

// Func adapts a function to Subscriber.
type Func func(rec Record)

// Deliver implements Subscriber.
func (f Func) Deliver(rec Record) { f(rec) }

// Bus fans each record out to the subscribers registered for its domain.
// A record whose domain has no subscriber is dropped.
type Bus struct {
	mu   sync.RWMutex
	subs map[string][]Subscriber
}

// New returns an empty bus.
func New() *Bus {
	return &Bus{subs: make(map[string][]Subscriber)}
}

// Subscribe registers s for domain.
func (b *Bus) Subscribe(domain string, s Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[domain] = append(b.subs[domain], s)
}

// Publish delivers rec to every current subscriber of rec.Domain.
func (b *Bus) Publish(rec Record) {
	b.mu.RLock()
	subs := b.subs[rec.Domain]
	b.mu.RUnlock()
	for _, s := range subs {
		s.Deliver(rec)
	}
}
