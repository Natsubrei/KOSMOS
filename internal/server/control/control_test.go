package control

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestHandlerRequiresUser(t *testing.T) {
	h := NewHub(func(string) []string { return []string{"cpu"} })
	srv := httptest.NewServer(h.Handler())
	t.Cleanup(srv.Close)

	res, err := srv.Client().Post(srv.URL+"/control", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestSnapshotOnHelloAndReplaceHost(t *testing.T) {
	h := NewHub(func(host string) []string {
		if host != "web-1" {
			t.Errorf("host %q", host)
		}
		return []string{"cpu"}
	})
	h.Keepalive = time.Hour
	srv := httptest.NewServer(h.Handler())
	t.Cleanup(srv.Close)

	c1, dec1 := dialControl(t, srv.URL, "web-1")
	t.Cleanup(func() { c1.Close() })
	got := readFrame(t, dec1)
	if got.Type != "snapshot" || len(got.Domains) != 1 || got.Domains[0] != "cpu" {
		t.Fatalf("first snapshot %#v", got)
	}

	c2, dec2 := dialControl(t, srv.URL, "web-1")
	t.Cleanup(func() { c2.Close() })
	got = readFrame(t, dec2)
	if got.Type != "snapshot" || len(got.Domains) != 1 || got.Domains[0] != "cpu" {
		t.Fatalf("second snapshot %#v", got)
	}

	if _, err := c1.Read(make([]byte, 1)); err == nil {
		t.Fatal("old session still readable")
	}
}

func dialControl(t *testing.T, base, host string) (net.Conn, *json.Decoder) {
	t.Helper()
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(conn, "POST /control HTTP/1.1\r\nHost: "+u.Host+"\r\nX-User-Id: dev\r\nContent-Length: 0\r\n\r\n")
	if err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(conn)
	res, err := http.ReadResponse(br, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(frame{Type: "hello", Host: host}); err != nil {
		t.Fatal(err)
	}
	return conn, json.NewDecoder(br)
}

func readFrame(t *testing.T, dec *json.Decoder) frame {
	t.Helper()
	var f frame
	if err := dec.Decode(&f); err != nil {
		t.Fatal(err)
	}
	return f
}
