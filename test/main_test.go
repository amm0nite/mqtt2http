package test

import (
	"io"
	"mqtt2http/broker"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

func TestMain(t *testing.T) {
	received := make(chan []byte, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("body read failed: %v", err)
		}
		received <- data
	})
	testHTTPServer := httptest.NewServer(handler)
	defer testHTTPServer.Close()

	addr := freePortAddr(t)
	config := &broker.BrokerConfig{
		TCPAddr:         addr,
		PublishURL:      testHTTPServer.URL,
		HTTPAddr:        freePortAddr(t),
		MetricsHTTPAddr: freePortAddr(t),
	}
	broker := broker.NewBroker(config)
	defer broker.Close()

	err := broker.Start()
	if err != nil {
		t.Fatalf("broker start failed: %v", err)
	}

	waitForMQTT(t, addr)

	opts := paho.NewClientOptions().AddBroker("tcp://" + addr).SetClientID("test")
	client := paho.NewClient(opts)

	token := client.Connect()
	if !token.WaitTimeout(5*time.Second) || token.Error() != nil {
		t.Fatalf("connect failed: %v", token.Error())
	}

	payload := []byte(`{"hello":"world"}`)
	token = client.Publish("devices/42/state", 0, false, payload)
	if !token.WaitTimeout(5*time.Second) || token.Error() != nil {
		t.Fatalf("publish failed: %v", token.Error())
	}

	select {
	case data := <-received:
		if string(data) != string(payload) {
			t.Fatalf("unexpected body.\nwant: %s\ngot:  %s", payload, data)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for forwarded request")
	}
}

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

func waitForMQTT(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatal("broker did not become ready in time")
		}
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}
