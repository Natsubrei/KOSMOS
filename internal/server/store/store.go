// Package store retains Records for listing. The first implementation is
// in-process memory; callers depend only on append and list.
package store

import (
	"sync"

	"github.com/natsubrei/kosmos/internal/bus"
)

// Store retains Records. Implementations must be safe for concurrent use.
type Store interface {
	Append(rec bus.Record)
	List() []bus.Record
}

// Memory keeps Records in process. Restart loses them.
type Memory struct {
	mu   sync.Mutex
	recs []bus.Record
}

// NewMemory returns an empty memory store.
func NewMemory() *Memory {
	return &Memory{}
}

// Append implements Store.
func (m *Memory) Append(rec bus.Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs = append(m.recs, rec)
}

// List implements Store. The returned slice is a copy.
func (m *Memory) List() []bus.Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]bus.Record, len(m.recs))
	copy(out, m.recs)
	return out
}

// Deliver implements bus.Subscriber.
func (m *Memory) Deliver(rec bus.Record) {
	m.Append(rec)
}

var _ bus.Subscriber = (*Memory)(nil)
var _ Store = (*Memory)(nil)
