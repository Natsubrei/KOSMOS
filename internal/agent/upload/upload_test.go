package upload

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

func TestDeliverPostsJSONBatch(t *testing.T) {
	got := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/records" || r.Method != http.MethodPost {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-User-Id") != "dev" {
			t.Errorf("user %q", r.Header.Get("X-User-Id"))
		}
		body, _ := io.ReadAll(r.Body)
		got <- body
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	u := New(srv.URL, "dev")
	u.Start(ctx)
	u.Deliver(bus.Record{
		Host:    "web-1",
		Domain:  "cpu",
		At:      time.Unix(1700000000, 0).UTC(),
		Payload: map[string]any{"Seq": float64(7)},
	})

	select {
	case body := <-got:
		var batch []map[string]any
		if err := json.Unmarshal(body, &batch); err != nil {
			t.Fatal(err)
		}
		if len(batch) != 1 || batch[0]["host"] != "web-1" || batch[0]["domain"] != "cpu" {
			t.Fatalf("batch %#v", batch)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no POST")
	}
}

func TestDeliverDropsWhenBufferFull(t *testing.T) {
	u := New("http://127.0.0.1:1", "dev")
	// Fill the buffer without starting the worker so nothing drains.
	for i := 0; i < cap(u.ch); i++ {
		u.Deliver(bus.Record{Domain: "cpu"})
	}
	u.Deliver(bus.Record{Domain: "cpu"})
	if u.Dropped() != 1 {
		t.Fatalf("dropped %d, want 1", u.Dropped())
	}
}
