package broker

import (
	"fmt"
	"mqtt2http/api"
	"mqtt2http/hooks"
	"mqtt2http/lib"
	"net/http"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Broker struct {
	config *BrokerConfig
	server *mqtt.Server
}

func NewBroker(config *BrokerConfig) *Broker {
	broker := &Broker{config: config}

	// Create the new MQTT Server.
	options := &mqtt.Options{
		InlineClient: true,
	}
	broker.server = mqtt.New(options)

	return broker
}

func (b *Broker) Start(reg prometheus.Registerer) error {
	var err error

	metrics := lib.NewMetrics(reg)

	// Create HTTP Client
	httpClient := lib.NewHTTPClient(
		b.config.ContentType,
		b.config.TopicHeader,
		b.config.AuthorizeURL,
		metrics,
	)

	// Create the client store
	clientStore := lib.NewClientStore(metrics)

	// Setup lifecycle hook
	lifecycleHook := &hooks.LifecycleHook{}
	err = b.server.AddHook(lifecycleHook, nil)
	if err != nil {
		return fmt.Errorf("failed to add lifecycle hook: %w", err)
	}

	// Setup connect-authenticate, acl, disconnect  hook
	authHook := &hooks.SessionHook{HTTPClient: httpClient, Store: clientStore}
	err = b.server.AddHook(authHook, nil)
	if err != nil {
		return fmt.Errorf("failed to add auth hook: %w", err)
	}

	// Setup publish hook
	publishHook := &hooks.PublishHook{HTTPClient: httpClient, Routes: b.config.Routes, Store: clientStore}
	err = b.server.AddHook(publishHook, nil)
	if err != nil {
		return fmt.Errorf("failed to add publish hook: %w", err)
	}

	// Create a TCP listener on a standard port.
	options := listeners.Config{ID: "t1", Address: b.config.TCPAddr}
	tcp := listeners.NewTCP(options)
	err = b.server.AddListener(tcp)
	if err != nil {
		return fmt.Errorf("failed to add TCP listener: %w", err)
	}

	// Start
	b.server.Log.Info("Starting MQTT server", "addr", b.config.TCPAddr)
	err = b.server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// HTTP server
	go func() {
		b.server.Log.Info("Starting API HTTP server", "addr", b.config.HTTPAddr)

		controller := api.NewController(b.server, clientStore, b.config.APIPassword)

		mux := http.NewServeMux()
		mux.HandleFunc("/", controller.RootHandler())
		mux.HandleFunc("/publish", controller.PublishHandler())
		mux.HandleFunc("/clients", controller.DumpHandler())

		err := http.ListenAndServe(b.config.HTTPAddr, mux)
		if err != nil {
			b.server.Log.Error("API HTTP server stopped", "err", err)
		}
	}()

	// Metrics HTTP server
	go (func() {
		b.server.Log.Info("Starting metrics HTTP server", "addr", b.config.MetricsHTTPAddr)

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		err := http.ListenAndServe(b.config.MetricsHTTPAddr, mux)
		if err != nil {
			b.server.Log.Error("Metrics HTTP server stopped", "err", err)
		}
	})()

	return nil
}

func (b *Broker) Close() {
	closed := make(chan bool)

	go func() {
		err := b.server.Close()
		if err != nil {
			b.server.Log.Error("Failed to close server", "err", err)
		}
		closed <- true
	}()

	timer := time.NewTimer(30 * time.Second)

	select {
	case <-closed:
		b.server.Log.Info("Server was gracefully shut down")
		return
	case <-timer.C:
		b.server.Log.Info("Server shutdown timeout")
		return
	}
}
