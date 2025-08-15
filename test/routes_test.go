package test

import (
	"mqtt2http/broker"
	"mqtt2http/lib"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
)

func TestPublishIsForwardedToDifferentHTTPRoutes(t *testing.T) {
	receivedA := make(chan []byte, 1)
	receivedB := make(chan []byte, 1)

	clientUsername := "testClient"
	clientPassword := "testPassword"

	authSrv := createAuthSrv(t, clientUsername, clientPassword)
	defer authSrv.Close()

	routeASrv := createPubSrv(t, receivedA)
	defer routeASrv.Close()

	routeBSrv := createPubSrv(t, receivedB)
	defer routeASrv.Close()

	// Choose free ports to avoid collisions.
	mqttAddr := freePortAddr(t) // "127.0.0.1:PORT"
	apiAddr := freePortAddr(t)
	metricsAddr := freePortAddr(t)

	cfg := &broker.BrokerConfig{
		TCPAddr:         mqttAddr,
		AuthorizeURL:    authSrv.URL,
		ContentType:     "application/json",
		HTTPAddr:        apiAddr,
		MetricsHTTPAddr: metricsAddr,
	}
	routeA := lib.Route{
		Name:    "RouteA",
		Pattern: "topicA",
		URL:     routeASrv.URL,
	}
	routeB := lib.Route{
		Name:    "RouteB",
		Pattern: "topicB",
		URL:     routeBSrv.URL,
	}
	cfg.Routes = append(cfg.Routes, routeA, routeB)

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

	topicA := "topicA"
	topicB := "topicB"

	payloadA := []byte(`payloadA`)
	payloadB := []byte(`payloadB`)

	if tok := client.Publish(topicA, 0, false, payloadA); !tok.WaitTimeout(5*time.Second) || tok.Error() != nil {
		t.Fatalf("publish failed: %v", tok.Error())
	}
	if tok := client.Publish(topicB, 0, false, payloadB); !tok.WaitTimeout(5*time.Second) || tok.Error() != nil {
		t.Fatalf("publish failed: %v", tok.Error())
	}

	select {
	case got := <-receivedA:
		if string(got) != string(payloadA) {
			t.Fatalf("unexpected forwarded body\nwant: %s\ngot:  %s", payloadA, got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for forwarded request A")
	}

	select {
	case got := <-receivedB:
		if string(got) != string(payloadB) {
			t.Fatalf("unexpected forwarded body\nwant: %s\ngot:  %s", payloadB, got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for forwarded request B")
	}
}
