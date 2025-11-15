package test

import (
	"fmt"
	"io"
	"mqtt2http/broker"
	"net/http"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPublishIsForwardedToHTTP(t *testing.T) {
	received := make(chan []byte, 1)

	clientUsername := "testClient"
	clientPassword := "testPassword"

	authSrv := createAuthSrv(t, clientUsername, clientPassword)
	defer authSrv.Close()

	pubSrv := createPubSrv(t, received)
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
	cfg.Load()

	b := broker.NewBroker(cfg)
	t.Cleanup(func() { b.Close() })

	if err := b.Start(prometheus.NewRegistry()); err != nil {
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

	// Check the content of the clients endpoint
	resp, err := http.Get(fmt.Sprintf("http://%s/clients", apiAddr))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("response read failed: %v", err)
	}
	if !strings.Contains(string(content), "\"id\":\"it-test\",\"username\":\"testClient\",\"subscriptions\":[\"devices/42/state\"],\"publications\":{\"devices/42/state\":1}") {
		t.Fatalf("unexpected content from the clients endpoint")
	}
}
