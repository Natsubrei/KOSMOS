package store

import (
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

func TestMemoryAppendListCopy(t *testing.T) {
	m := NewMemory()
	rec := bus.Record{Domain: "example", At: time.Unix(1, 0).UTC(), Payload: map[string]any{"value": float64(1)}}
	m.Append(rec)

	got := m.List()
	if len(got) != 1 || got[0].Domain != "example" {
		t.Fatalf("list %+v", got)
	}
	got[0].Domain = "mutated"
	if m.List()[0].Domain != "example" {
		t.Fatal("List must return a copy")
	}
}

func TestDeliverAppends(t *testing.T) {
	m := NewMemory()
	m.Deliver(bus.Record{Domain: "example", Payload: "p"})
	if n := len(m.List()); n != 1 {
		t.Fatalf("len %d, want 1", n)
	}
}
