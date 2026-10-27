// Package report renders record-bus output for humans; the JSON-lines
// reporter is the first subscriber, used to verify the chain end to end.
package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

// JSONLines writes one JSON object per record, in the order delivered.
type JSONLines struct {
	enc *json.Encoder
}

// NewJSONLines returns a reporter writing to w.
func NewJSONLines(w io.Writer) *JSONLines {
	return &JSONLines{enc: json.NewEncoder(w)}
}

// Deliver implements bus.Subscriber. Write failures are dropped: a closed
// output ends reporting, it does not end collection.
func (r *JSONLines) Deliver(rec bus.Record) {
	_ = r.enc.Encode(struct {
		Host    string    `json:"host"`
		Domain  string    `json:"domain"`
		At      time.Time `json:"at"`
		Payload any       `json:"payload"`
	}{rec.Host, rec.Domain, rec.At, rec.Payload})
}
