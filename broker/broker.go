package broker

import (
	"fmt"
	"mqtt2http/api"
	"mqtt2http/hooks"
	"mqtt2http/lib"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type BrokerConfig struct {
	TCPAddr         string
	HTTPAddr        string
	AuthorizeURL    string
	PublishURL      string
	ContentType     string
	TopicHeader     string
	MetricsHTTPAddr string
	RoutesFilePath  string
	APIPassword     string
}

type Broker struct {
	config *BrokerConfig
	server *mqtt.Server
}

func NewBroker(config *BrokerConfig) *Broker {
	broker := &Broker{config: config}

	// Create the new MQTT Server.
	options := &mqtt.Options{
		InlineClient: true,
		Capabilities: &mqtt.Capabilities{
			MaximumSessionExpiryInterval: 3600,
		},
	}
	broker.server = mqtt.New(options)

	return broker
}

func (b *Broker) Start() error {
	var err error

	// Create HTTP Client
	err = godotenv.Load()
	if err != nil {
		b.server.Log.Warn("Failed to read .env file", "err", err)
	}

	metrics := lib.NewMetrics()

	client := &lib.Client{
		Server:      b.server,
		ContentType: b.config.ContentType,
		TopicHeader: b.config.TopicHeader,
		Metrics:     metrics,
	}

	// Setup auth hook
	authHook := &hooks.AuthHook{Client: client, URL: b.config.AuthorizeURL}
	err = b.server.AddHook(authHook, nil)
	if err != nil {
		return fmt.Errorf("failed to add auth hook: %w", err)
	}

	// Setup publish hook
	publishHook := &hooks.PublishHook{Client: client, DefaultURL: b.config.PublishURL, RoutesFilePath: b.config.RoutesFilePath}
	err = b.server.AddHook(publishHook, map[string]any{})
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
	err = b.server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// HTTP server
	go func() {
		b.server.Log.Info("Starting API HTTP server", "addr", b.config.HTTPAddr)

		controller := api.CreateController(b.server, client, b.config.APIPassword)

		mux := http.NewServeMux()
		mux.HandleFunc("/", controller.RootHandler())
		mux.HandleFunc("/publish", controller.PublishHandler())

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
