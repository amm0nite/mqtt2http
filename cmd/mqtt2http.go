package main

import (
	"fmt"
	"mqtt2http/api"
	"mqtt2http/hooks"
	"mqtt2http/lib"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create the new MQTT Server.
	options := &mqtt.Options{
		InlineClient: true,
		Capabilities: &mqtt.Capabilities{
			MaximumSessionExpiryInterval: 3600,
		},
	}
	server := mqtt.New(options)

	err := start(server)
	if err != nil {
		server.Log.Error("Failed to start", "err", err)
		panic(err)
	}

	// Handle signals
	go func() {
		sig := <-sigs
		server.Log.Info("Signal received", "signal", sig.String())
		done <- true
	}()

	server.Log.Info("Awaiting signal")
	<-done
	server.Log.Info("Exiting")
}

func start(server *mqtt.Server) error {
	var err error

	// Create HTTP Client
	err = godotenv.Load()
	if err != nil {
		server.Log.Warn("Failed to read .env file", "err", err)
	}

	metrics, err := lib.NewMetrics()
	if err != nil {
		return fmt.Errorf("failed to setup metrics: %w", err)
	}

	tcpAddr := getEnv("MQTT2HTTP_MQTT_LISTEN_ADDRESS", ":1883")
	httpAddr := getEnv("MQTT2HTTP_HTTP_LISTEN_ADDRESS", ":8080")
	authorizeURL := getEnv("MQTT2HTTP_AUTHORIZE_URL", "http://example.com")
	publishURL := getEnv("MQTT2HTTP_PUBLISH_URL", "http://example.com/{topic}")
	contentType := getEnv("MQTT2HTTP_CONTENT_TYPE", "application/octet-stream")
	topicHeader := getEnv("MQTT2HTTP_TOPIC_HEADER", "X-Topic")
	metricsHttpAddr := getEnv("MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS", ":9090")
	routesFilePath := getEnv("MQTT2HTTP_ROUTES_FILE_PATH", "routes.yaml")
	apiPassword := getEnv("MQTT2HTTP_API_PASSWORD", uuid.NewString())

	client := &lib.Client{
		Server:      server,
		ContentType: contentType,
		TopicHeader: topicHeader,
		Metrics:     metrics,
	}

	// Setup auth hook
	authHook := &hooks.AuthHook{Client: client, URL: authorizeURL}
	err = server.AddHook(authHook, nil)
	if err != nil {
		return fmt.Errorf("failed to add auth hook: %w", err)
	}

	// Setup publish hook
	publishHook := &hooks.PublishHook{Client: client, DefaultURL: publishURL, RoutesFilePath: routesFilePath}
	err = server.AddHook(publishHook, map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to add publish hook: %w", err)
	}

	// Create a TCP listener on a standard port.
	options := listeners.Config{ID: "t1", Address: tcpAddr}
	tcp := listeners.NewTCP(options)
	err = server.AddListener(tcp)
	if err != nil {
		return fmt.Errorf("failed to add TCP listener: %w", err)
	}

	// Start
	err = server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// HTTP server
	go func() {
		server.Log.Info("Starting API HTTP server", "addr", httpAddr)

		controller := api.CreateController(server, client, apiPassword)

		mux := http.NewServeMux()
		mux.HandleFunc("/", controller.RootHandler())
		mux.HandleFunc("/publish", controller.PublishHandler())

		err := http.ListenAndServe(httpAddr, mux)
		if err != nil {
			server.Log.Error("API HTTP server stopped", "err", err)
		}
	}()

	// Metrics HTTP server
	go (func() {
		server.Log.Info("Starting metrics HTTP server", "addr", metricsHttpAddr)

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		err := http.ListenAndServe(metricsHttpAddr, mux)
		if err != nil {
			server.Log.Error("Metrics HTTP server stopped", "err", err)
		}
	})()

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
