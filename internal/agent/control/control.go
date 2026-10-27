// Package control is the agent end of the outbound control session.
// It applies domain snapshots through the runtime kit; it does not import
// collection domains.
package control

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	defaultKeepalive = 10 * time.Second
	defaultStopAfter = 60 * time.Second
	reconnectWait    = 200 * time.Millisecond
)

type frame struct {
	Type    string   `json:"type"`
	Host    string   `json:"host,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

// Client dials POST /control, applies snapshots, and stops collection
// StopAfter after keepalive failure.
type Client struct {
	BaseURL   string
	User      string
	Host      string
	Apply     func(domains []string) error
	Keepalive time.Duration
	StopAfter time.Duration
}

// Run reconnects until ctx is done.
func (c *Client) Run(ctx context.Context) error {
	ka := c.Keepalive
	if ka == 0 {
		ka = defaultKeepalive
	}
	stopAfter := c.StopAfter
	if stopAfter == 0 {
		stopAfter = defaultStopAfter
	}

	var stop *time.Timer
	var gen atomic.Uint64
	arm := func() {
		if stop != nil {
			return
		}
		id := gen.Add(1)
		stop = time.AfterFunc(stopAfter, func() {
			if gen.Load() == id {
				_ = c.Apply(nil)
			}
		})
	}
	disarm := func() {
		gen.Add(1)
		if stop == nil {
			return
		}
		stop.Stop()
		stop = nil
	}

	for {
		if err := ctx.Err(); err != nil {
			disarm()
			return err
		}
		_ = c.session(ctx, ka, disarm)
		if ctx.Err() != nil {
			disarm()
			return ctx.Err()
		}
		arm()
		select {
		case <-ctx.Done():
			disarm()
			return ctx.Err()
		case <-time.After(reconnectWait):
		}
	}
}

func (c *Client) session(ctx context.Context, ka time.Duration, live func()) error {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		return err
	}
	defer conn.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/control", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-Id", c.User)
	if err := req.Write(conn); err != nil {
		return err
	}
	br := bufio.NewReader(conn)
	res, err := http.ReadResponse(br, req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errStatus(res.StatusCode)
	}

	enc := json.NewEncoder(conn)
	if err := enc.Encode(frame{Type: "hello", Host: c.Host}); err != nil {
		return err
	}
	live()
	dec := json.NewDecoder(br)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := conn.SetReadDeadline(time.Now().Add(3 * ka)); err != nil {
			return err
		}
		var f frame
		if err := dec.Decode(&f); err != nil {
			return err
		}
		switch f.Type {
		case "snapshot":
			if err := c.Apply(f.Domains); err != nil {
				return err
			}
		case "ping":
			if err := enc.Encode(frame{Type: "pong"}); err != nil {
				return err
			}
		}
	}
}

type statusError int

func errStatus(code int) error { return statusError(code) }

func (e statusError) Error() string { return http.StatusText(int(e)) }
