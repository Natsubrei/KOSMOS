// Package httpapi is the JSON HTTP face of the server: bytes become Records
// on the bus; GET lists the store. Domain bundles and analyzers do not import
// this package.
package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
	"github.com/natsubrei/kosmos/internal/server/store"
)

type wireRecord struct {
	Host    string          `json:"host"`
	Domain  string          `json:"domain"`
	At      time.Time       `json:"at"`
	Payload json.RawMessage `json:"payload"`
}

// Handler serves POST /records onto b and GET /records from st.
// Requests need a non-empty X-User-Id header (development auth).
func Handler(b *bus.Bus, st store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /records", func(w http.ResponseWriter, r *http.Request) {
		if !devUser(w, r) {
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		var batch []wireRecord
		if err := json.Unmarshal(body, &batch); err != nil {
			http.Error(w, "json", http.StatusBadRequest)
			return
		}
		for _, wr := range batch {
			if wr.Host == "" {
				http.Error(w, "host", http.StatusBadRequest)
				return
			}
			if wr.Domain == "" {
				http.Error(w, "domain", http.StatusBadRequest)
				return
			}
			var payload any
			if len(wr.Payload) > 0 {
				if err := json.Unmarshal(wr.Payload, &payload); err != nil {
					http.Error(w, "payload", http.StatusBadRequest)
					return
				}
			}
			b.Publish(bus.Record{Host: wr.Host, Domain: wr.Domain, At: wr.At, Payload: payload})
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /records", func(w http.ResponseWriter, r *http.Request) {
		if !devUser(w, r) {
			return
		}
		list := st.List()
		out := make([]wireRecord, 0, len(list))
		for _, rec := range list {
			raw, err := json.Marshal(rec.Payload)
			if err != nil {
				http.Error(w, "encode", http.StatusInternalServerError)
				return
			}
			out = append(out, wireRecord{Host: rec.Host, Domain: rec.Domain, At: rec.At, Payload: raw})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
	return mux
}

func devUser(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-User-Id") == "" {
		http.Error(w, "X-User-Id", http.StatusUnauthorized)
		return false
	}
	return true
}
