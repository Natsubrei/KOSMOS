package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cpucontracts "github.com/natsubrei/kosmos/contracts/cpu"
	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/store"
)

func TestPostThenGetRoundTrip(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe(cpucontracts.Domain, st)
	srv := httptest.NewServer(Handler(b, st))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/records", strings.NewReader(
		`[{"host":"web-1","domain":"cpu","at":"2023-11-14T22:13:20Z","payload":{"Seq":7}}]`,
	))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-User-Id", "dev")
	req.Header.Set("Content-Type", "application/json")
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("POST status %d body %s", res.StatusCode, body)
	}

	get, err := http.NewRequest(http.MethodGet, srv.URL+"/records", nil)
	if err != nil {
		t.Fatal(err)
	}
	get.Header.Set("X-User-Id", "dev")
	gres, err := srv.Client().Do(get)
	if err != nil {
		t.Fatal(err)
	}
	defer gres.Body.Close()
	if gres.StatusCode != http.StatusOK {
		t.Fatalf("GET status %d", gres.StatusCode)
	}
	var got []map[string]any
	if err := json.NewDecoder(gres.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["host"] != "web-1" || got[0]["domain"] != "cpu" {
		t.Fatalf("got %#v", got)
	}
	payload, ok := got[0]["payload"].(map[string]any)
	if !ok || payload["Seq"] != float64(7) {
		t.Fatalf("payload %#v", got[0]["payload"])
	}
}

func TestPostRejectsMissingHost(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe(cpucontracts.Domain, st)
	srv := httptest.NewServer(Handler(b, st))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/records", strings.NewReader(
		`[{"domain":"cpu","at":"2023-11-14T22:13:20Z","payload":{"Seq":7}}]`,
	))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-User-Id", "dev")
	req.Header.Set("Content-Type", "application/json")
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", res.StatusCode)
	}
	if n := len(st.List()); n != 0 {
		t.Fatalf("stored %d records, want 0", n)
	}
}

func TestMissingUserUnauthorized(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	srv := httptest.NewServer(Handler(b, st))
	t.Cleanup(srv.Close)

	res, err := srv.Client().Get(srv.URL + "/records")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", res.StatusCode)
	}
}
