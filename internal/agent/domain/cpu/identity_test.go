//go:build linux

package cpu

import "testing"

func TestStatStarttime(t *testing.T) {
	// pid, comm, then fields 3..22; starttime is the 20th token after comm.
	stat := []byte("42 (yes) S 1 1 1 0 -1 0 0 0 0 0 0 0 0 0 0 0 1 0 12345 0")
	got, ok := statStarttime(stat)
	if !ok || got != 12345 {
		t.Fatalf("got %d ok=%v", got, ok)
	}
	if _, ok := statStarttime([]byte("no-paren")); ok {
		t.Fatal("expected false")
	}
}

func TestStatusTgid(t *testing.T) {
	st := []byte("Name:\tyes\nTgid:\t42\nPid:\t42\n")
	got, ok := statusTgid(st)
	if !ok || got != 42 {
		t.Fatalf("got %d ok=%v", got, ok)
	}
	if _, ok := statusTgid([]byte("Name:\tx\n")); ok {
		t.Fatal("expected false")
	}
}
