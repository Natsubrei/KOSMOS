package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/diagnose"
	"github.com/natsubrei/kosmos/internal/server/store"
)

type staticReports []diagnose.Report

func (s staticReports) ListReports() []diagnose.Report { return s }

func TestPostThenGetRoundTrip(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe("example", st)
	srv := httptest.NewServer(Handler(b, st, staticReports{}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/records", strings.NewReader(
		`[{"host":"web-1","domain":"example","at":"2023-11-14T22:13:20Z","payload":{"value":7}}]`,
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
	if len(got) != 1 || got[0]["host"] != "web-1" || got[0]["domain"] != "example" {
		t.Fatalf("got %#v", got)
	}
	payload, ok := got[0]["payload"].(map[string]any)
	if !ok || payload["value"] != float64(7) {
		t.Fatalf("payload %#v", got[0]["payload"])
	}
}

func TestPostRejectsMissingHost(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	b.Subscribe("example", st)
	srv := httptest.NewServer(Handler(b, st, staticReports{}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/records", strings.NewReader(
		`[{"domain":"example","at":"2023-11-14T22:13:20Z","payload":{"value":7}}]`,
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
	srv := httptest.NewServer(Handler(b, st, staticReports{}))
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

func TestGetReportsRoundTrip(t *testing.T) {
	b := bus.New()
	st := store.NewMemory()
	rep := diagnose.Report{
		ID:         "web-1/0/10/cpu",
		ProducedAt: time.Unix(50, 0).UTC(),
		Host:       "web-1",
		Analyzer:   "cpu",
		Window: diagnose.Window{
			ID: "web-1/0/10", Host: "web-1",
			Start: time.Unix(0, 0).UTC(), End: time.Unix(10, 0).UTC(),
		},
		Conclusion: diagnose.Conclusion{
			Kind:    "example.claim",
			Subject: diagnose.Subject{Host: "web-1", Pid: 42, Tid: 42},
			Stance:  diagnose.StanceHolds,
		},
		Evidence: []diagnose.Evidence{{
			ID:      "e1",
			Kind:    "example.measure",
			Payload: map[string]any{"value": float64(1)},
			Predicate: &diagnose.Predicate{
				Observed: 1, Operator: diagnose.OpGT, Bound: 0,
			},
			Citations: []diagnose.Citation{
				{Window: &diagnose.WindowCite{
					ID: "web-1/0/10", Start: time.Unix(0, 0).UTC(), End: time.Unix(10, 0).UTC(),
				}},
				{Record: &diagnose.RecordSnap{
					Domain: "example", At: time.Unix(5, 0).UTC(), Payload: map[string]any{"value": float64(1)},
				}},
			},
		}},
	}
	srv := httptest.NewServer(Handler(b, st, staticReports{rep}))
	t.Cleanup(srv.Close)

	get, err := http.NewRequest(http.MethodGet, srv.URL+"/reports", nil)
	if err != nil {
		t.Fatal(err)
	}
	get.Header.Set("X-User-Id", "dev")
	res, err := srv.Client().Do(get)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, body)
	}
	var got []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0]["id"] != "web-1/0/10/cpu" || got[0]["analyzer"] != "cpu" {
		t.Fatalf("got %#v", got)
	}
	con, ok := got[0]["conclusion"].(map[string]any)
	if !ok || con["kind"] != "example.claim" || con["stance"] != "holds" {
		t.Fatalf("conclusion %#v", got[0]["conclusion"])
	}
}
