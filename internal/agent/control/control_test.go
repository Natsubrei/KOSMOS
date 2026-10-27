package control

import (
	"context"
	"reflect"
	"testing"
	"time"

	servercontrol "github.com/natsubrei/kosmos/internal/server/control"
	"net/http/httptest"
)

func TestRunAppliesSnapshot(t *testing.T) {
	hub := servercontrol.NewHub(func(string) []string { return []string{"cpu"} })
	hub.Keepalive = time.Hour
	srv := httptest.NewServer(hub.Handler())
	t.Cleanup(srv.Close)

	got := make(chan []string, 1)
	c := Client{
		BaseURL:   srv.URL,
		User:      "dev",
		Host:      "web-1",
		Keepalive: time.Hour,
		StopAfter: time.Hour,
		Apply: func(domains []string) error {
			got <- append([]string(nil), domains...)
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = c.Run(ctx) }()

	select {
	case domains := <-got:
		if !reflect.DeepEqual(domains, []string{"cpu"}) {
			t.Fatalf("applied %#v", domains)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no snapshot")
	}
}

func TestDisconnectStopsAfterGrace(t *testing.T) {
	hub := servercontrol.NewHub(func(string) []string { return []string{"cpu"} })
	hub.Keepalive = 15 * time.Millisecond
	srv := httptest.NewServer(hub.Handler())

	got := make(chan []string, 4)
	c := Client{
		BaseURL:   srv.URL,
		User:      "dev",
		Host:      "web-1",
		Keepalive: 30 * time.Millisecond,
		StopAfter: 80 * time.Millisecond,
		Apply: func(domains []string) error {
			got <- append([]string(nil), domains...)
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = c.Run(ctx) }()

	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("no snapshot")
	}

	hub.Close()
	srv.Close()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		select {
		case domains := <-got:
			if len(domains) == 0 {
				return
			}
		case <-time.After(20 * time.Millisecond):
		}
	}
	t.Fatal("want Apply(nil) after grace")
}
