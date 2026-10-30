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
	"github.com/natsubrei/kosmos/internal/server/diagnose"
	"github.com/natsubrei/kosmos/internal/server/store"
)

type wireRecord struct {
	Host    string          `json:"host"`
	Domain  string          `json:"domain"`
	At      time.Time       `json:"at"`
	Payload json.RawMessage `json:"payload"`
}

// ReportLister is the diagnosis side of GET /reports. The loop holds Reports;
// this package only encodes them.
type ReportLister interface {
	ListReports() []diagnose.Report
}

// Handler serves POST /records onto b, GET /records from st, and GET /reports
// from reports. Requests need a non-empty X-User-Id header (development auth).
func Handler(b *bus.Bus, st store.Store, reports ReportLister) http.Handler {
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
	mux.HandleFunc("GET /reports", func(w http.ResponseWriter, r *http.Request) {
		if !devUser(w, r) {
			return
		}
		list := reports.ListReports()
		out := make([]wireReport, 0, len(list))
		for _, rep := range list {
			wr, err := encodeReport(rep)
			if err != nil {
				http.Error(w, "encode", http.StatusInternalServerError)
				return
			}
			out = append(out, wr)
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

type wireReport struct {
	ID         string         `json:"id"`
	ProducedAt time.Time      `json:"produced_at"`
	Host       string         `json:"host"`
	Analyzer   string         `json:"analyzer"`
	Window     wireWindow     `json:"window"`
	Conclusion wireConclusion `json:"conclusion"`
	Evidence   []wireEvidence `json:"evidence"`
}

type wireWindow struct {
	ID    string    `json:"id"`
	Host  string    `json:"host,omitempty"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type wireConclusion struct {
	Kind    string      `json:"kind"`
	Subject wireSubject `json:"subject"`
	Stance  string      `json:"stance"`
}

type wireSubject struct {
	Host      string `json:"host"`
	Pid       int32  `json:"pid"`
	Tid       int32  `json:"tid"`
	Starttime uint64 `json:"starttime"`
}

type wireEvidence struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Payload   json.RawMessage `json:"payload"`
	Predicate *wirePredicate  `json:"predicate,omitempty"`
	Citations []wireCitation  `json:"citations"`
}

type wirePredicate struct {
	Observed float64 `json:"observed"`
	Operator string  `json:"operator"`
	Bound    float64 `json:"bound"`
}

type wireCitation struct {
	Record     *wireRecordSnap    `json:"record,omitempty"`
	Window     *wireWindow        `json:"window,omitempty"`
	EvidenceID string             `json:"evidence_id,omitempty"`
	Missing    *wireMissingSource `json:"missing,omitempty"`
}

type wireRecordSnap struct {
	Domain  string          `json:"domain"`
	At      time.Time       `json:"at"`
	Payload json.RawMessage `json:"payload"`
}

type wireMissingSource struct {
	Source string `json:"source"`
	Reason string `json:"reason"`
}

func encodeReport(rep diagnose.Report) (wireReport, error) {
	ev := make([]wireEvidence, 0, len(rep.Evidence))
	for _, e := range rep.Evidence {
		raw, err := json.Marshal(e.Payload)
		if err != nil {
			return wireReport{}, err
		}
		we := wireEvidence{
			ID: e.ID, Kind: e.Kind, Payload: raw,
			Citations: make([]wireCitation, 0, len(e.Citations)),
		}
		if e.Predicate != nil {
			we.Predicate = &wirePredicate{
				Observed: e.Predicate.Observed,
				Operator: string(e.Predicate.Operator),
				Bound:    e.Predicate.Bound,
			}
		}
		for _, c := range e.Citations {
			wc, err := encodeCitation(c)
			if err != nil {
				return wireReport{}, err
			}
			we.Citations = append(we.Citations, wc)
		}
		ev = append(ev, we)
	}
	return wireReport{
		ID:         rep.ID,
		ProducedAt: rep.ProducedAt,
		Host:       rep.Host,
		Analyzer:   rep.Analyzer,
		Window: wireWindow{
			ID: rep.Window.ID, Host: rep.Window.Host,
			Start: rep.Window.Start, End: rep.Window.End,
		},
		Conclusion: wireConclusion{
			Kind:   rep.Conclusion.Kind,
			Stance: string(rep.Conclusion.Stance),
			Subject: wireSubject{
				Host:      rep.Conclusion.Subject.Host,
				Pid:       rep.Conclusion.Subject.Pid,
				Tid:       rep.Conclusion.Subject.Tid,
				Starttime: rep.Conclusion.Subject.Starttime,
			},
		},
		Evidence: ev,
	}, nil
}

func encodeCitation(c diagnose.Citation) (wireCitation, error) {
	switch {
	case c.Record != nil:
		raw, err := json.Marshal(c.Record.Payload)
		if err != nil {
			return wireCitation{}, err
		}
		return wireCitation{Record: &wireRecordSnap{
			Domain: c.Record.Domain, At: c.Record.At, Payload: raw,
		}}, nil
	case c.Window != nil:
		return wireCitation{Window: &wireWindow{
			ID: c.Window.ID, Start: c.Window.Start, End: c.Window.End,
		}}, nil
	case c.EvidenceID != "":
		return wireCitation{EvidenceID: c.EvidenceID}, nil
	case c.Missing != nil:
		return wireCitation{Missing: &wireMissingSource{
			Source: c.Missing.Source, Reason: c.Missing.Reason,
		}}, nil
	default:
		return wireCitation{}, nil
	}
}
