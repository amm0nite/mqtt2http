package test

import (
	"context"
	"io"
	"mqtt2http/broker"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestPublishIsForwardedToHTTP(t *testing.T) {
	received := make(chan []byte, 1)

	clientUsername := "testClient"
	clientPassword := "testPassword"

	// Authorize endpoint: assert Basic Auth
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("authorize: expected POST, got %s", r.Method)
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Fatal("authorize: missing basic auth")
		}
		if username != clientUsername {
			t.Fatalf("authorize: wrong username (%q)", username)
		}
		if password != clientPassword {
			t.Fatalf("authorize: wrong password")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer authSrv.Close()

	// Publish endpoint: capture body + basic request sanity
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
		received <- b
		w.WriteHeader(http.StatusNoContent)
	}))
	defer pubSrv.Close()

	// Choose free ports to avoid collisions.
	mqttAddr := freePortAddr(t) // "127.0.0.1:PORT"
	apiAddr := freePortAddr(t)
	metricsAddr := freePortAddr(t)

	cfg := &broker.BrokerConfig{
		TCPAddr:         mqttAddr,
		AuthorizeURL:    authSrv.URL,
		PublishURL:      pubSrv.URL,
		ContentType:     "application/json",
		HTTPAddr:        apiAddr,
		MetricsHTTPAddr: metricsAddr,
	}
	b := broker.NewBroker(cfg)
	t.Cleanup(func() { b.Close() })

	if err := b.Start(); err != nil {
		t.Fatalf("broker start failed: %v", err)
	}

	// Optionally: wait until the MQTT port accepts TCP (reduces startup races).
	waitForTCP(t, mqttAddr, 5*time.Second)

	// Connect MQTT client with retry.
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://" + mqttAddr).
		SetClientID("it-test").
		SetUsername(clientUsername).
		SetPassword(clientPassword).
		SetConnectRetry(true).
		SetConnectRetryInterval(200 * time.Millisecond).
		SetConnectTimeout(1 * time.Second)

	client := mqtt.NewClient(opts)
	t.Cleanup(func() { client.Disconnect(250) })

	if tok := client.Connect(); !tok.WaitTimeout(5*time.Second) || tok.Error() != nil {
		t.Fatalf("connect failed: %v", tok.Error())
	}
	if !client.IsConnected() {
		t.Fatalf("client reports not connected")
	}

	// Publish and assert HTTP saw the payload.
	payload := []byte(`{"hello":"world"}`)
	if tok := client.Publish("devices/42/state", 0, false, payload); !tok.WaitTimeout(5*time.Second) || tok.Error() != nil {
		t.Fatalf("publish failed: %v", tok.Error())
	}

	select {
	case got := <-received:
		if string(got) != string(payload) {
			t.Fatalf("unexpected forwarded body\nwant: %s\ngot:  %s", payload, got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for forwarded request")
	}
}

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
