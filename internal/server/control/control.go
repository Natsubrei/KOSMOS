// Package control is the server end of the agent-outbound control session.
// Domain policy is injected; this package does not import collection domains.
package control

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

const defaultKeepalive = 10 * time.Second

type frame struct {
	Type    string   `json:"type"`
	Host    string   `json:"host,omitempty"`
	Domains []string `json:"domains,omitempty"`
}

// Hub holds one session per host. A new session replaces the old connection.
type Hub struct {
	domains   func(host string) []string
	Keepalive time.Duration

	mu       sync.Mutex
	sessions map[string]net.Conn
}

// NewHub returns a hub. domains is the snapshot to send when a session appears.
func NewHub(domains func(host string) []string) *Hub {
	return &Hub{domains: domains, sessions: make(map[string]net.Conn)}
}

// Handler serves POST /control. Development auth is a non-empty X-User-Id.
func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(h.serve)
}

func (h *Hub) serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("X-User-Id") == "" {
		http.Error(w, "X-User-Id", http.StatusUnauthorized)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack", http.StatusInternalServerError)
		return
	}
	conn, brw, err := hj.Hijack()
	if err != nil {
		return
	}
	if _, err := brw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n"); err != nil {
		conn.Close()
		return
	}
	if err := brw.Flush(); err != nil {
		conn.Close()
		return
	}
	go h.session(conn, brw)
}

func (h *Hub) session(conn net.Conn, brw *bufio.ReadWriter) {
	defer conn.Close()
	dec := json.NewDecoder(brw)
	var hello frame
	if err := dec.Decode(&hello); err != nil || hello.Type != "hello" || hello.Host == "" {
		return
	}

	h.mu.Lock()
	if old, ok := h.sessions[hello.Host]; ok {
		old.Close()
	}
	h.sessions[hello.Host] = conn
	snap := h.domains(hello.Host)
	h.mu.Unlock()

	enc := json.NewEncoder(brw)
	var wmu sync.Mutex
	write := func(f frame) error {
		wmu.Lock()
		defer wmu.Unlock()
		if err := enc.Encode(f); err != nil {
			return err
		}
		return brw.Flush()
	}
	if err := write(frame{Type: "snapshot", Domains: snap}); err != nil {
		h.drop(hello.Host, conn)
		return
	}

	ka := h.Keepalive
	if ka == 0 {
		ka = defaultKeepalive
	}
	tick := time.NewTicker(ka)
	defer tick.Stop()

	readErr := make(chan error, 1)
	go func() {
		for {
			var f frame
			if err := dec.Decode(&f); err != nil {
				readErr <- err
				return
			}
		}
	}()

	for {
		select {
		case <-readErr:
			h.drop(hello.Host, conn)
			return
		case <-tick.C:
			if err := write(frame{Type: "ping"}); err != nil {
				h.drop(hello.Host, conn)
				return
			}
		}
	}
}

// Close ends every session. In-flight handlers return as their connections drop.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for host, c := range h.sessions {
		c.Close()
		delete(h.sessions, host)
	}
}

func (h *Hub) drop(host string, conn net.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[host] == conn {
		delete(h.sessions, host)
	}
}
