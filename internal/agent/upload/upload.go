// Package upload ships Records to the server over HTTP JSON. It is a bus
// subscriber: domain bundles do not import this package. Deliver never
// blocks on the network; a full buffer drops the Record.
package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/natsubrei/kosmos/internal/bus"
)

type wireRecord struct {
	Domain  string    `json:"domain"`
	At      time.Time `json:"at"`
	Payload any       `json:"payload"`
}

// Uploader POSTs Record batches to url. Start the worker before Subscribe.
type Uploader struct {
	url     string
	user    string
	client  *http.Client
	ch      chan bus.Record
	dropped atomic.Int64
}

// New returns an Uploader for server base URL (no trailing slash) and X-User-Id.
func New(url, user string) *Uploader {
	return &Uploader{
		url:    url,
		user:   user,
		client: &http.Client{Timeout: 2 * time.Second},
		ch:     make(chan bus.Record, 64),
	}
}

// Start runs the POST worker until ctx is done.
func (u *Uploader) Start(ctx context.Context) {
	go u.loop(ctx)
}

// Deliver implements bus.Subscriber.
func (u *Uploader) Deliver(rec bus.Record) {
	select {
	case u.ch <- rec:
	default:
		u.dropped.Add(1)
	}
}

// Dropped is the count of Records dropped because the buffer was full.
func (u *Uploader) Dropped() int64 {
	return u.dropped.Load()
}

func (u *Uploader) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rec := <-u.ch:
			u.post([]bus.Record{rec})
		}
	}
}

func (u *Uploader) post(recs []bus.Record) {
	batch := make([]wireRecord, len(recs))
	for i, rec := range recs {
		batch[i] = wireRecord{Domain: rec.Domain, At: rec.At, Payload: rec.Payload}
	}
	body, err := json.Marshal(batch)
	if err != nil {
		u.dropped.Add(int64(len(recs)))
		return
	}
	req, err := http.NewRequest(http.MethodPost, u.url+"/records", bytes.NewReader(body))
	if err != nil {
		u.dropped.Add(int64(len(recs)))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", u.user)
	res, err := u.client.Do(req)
	if err != nil {
		u.dropped.Add(int64(len(recs)))
		return
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		u.dropped.Add(int64(len(recs)))
	}
}

var _ bus.Subscriber = (*Uploader)(nil)
