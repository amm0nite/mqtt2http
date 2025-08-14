package test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// freePortAddr returns "127.0.0.1:PORT" by binding to :0 and closing.
func freePortAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

// waitForTCP polls until a TCP connection to addr succeeds or the deadline passes.
func waitForTCP(t *testing.T, addr string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		d := net.Dialer{Timeout: 150 * time.Millisecond}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return
		}
		if ctx.Err() != nil {
			t.Fatalf("broker did not accept TCP on %s in %v", addr, timeout)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Authorize endpoint: assert Basic Auth
func createAuthSrv(t *testing.T, expectedUsername string, expectedPassword string) *httptest.Server {
	t.Helper()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("authorize: expected POST, got %s", r.Method)
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Fatal("authorize: missing basic auth")
		}
		if username != expectedUsername {
			t.Fatalf("authorize: wrong username (%q)", username)
		}
		if password != expectedPassword {
			t.Fatalf("authorize: wrong password")
		}
		w.WriteHeader(http.StatusOK)
	}))
	return authSrv
}

// Publish endpoint: capture body + basic request sanity
func createPubSrv(t *testing.T, receiveChan chan []byte) *httptest.Server {
	t.Helper()

	pubSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("publish: expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct == "" {
			t.Fatalf("publish: missing Content-Type header")
		}
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("publish: read body failed: %v", err)
		}
		receiveChan <- b
		w.WriteHeader(http.StatusNoContent)
	}))
	return pubSrv
}
