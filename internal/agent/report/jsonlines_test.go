package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

func TestDeliverWritesOneJSONLine(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONLines(&buf)

	r.Deliver(bus.Record{
		Host:    "web-1",
		Domain:  "cpu",
		At:      time.Unix(1700000000, 0).UTC(),
		Payload: struct{ Seq int64 }{Seq: 7},
	})

	line := strings.TrimSuffix(buf.String(), "\n")
	if strings.Count(buf.String(), "\n") != 1 {
		t.Fatalf("output %q: want exactly one line", buf.String())
	}
	for _, want := range []string{
		`"host":"web-1"`,
		`"domain":"cpu"`,
		`"at":"2023-11-14T22:13:20Z"`,
		`"payload":{"Seq":7}`,
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("line %q missing %q", line, want)
		}
	}
}
