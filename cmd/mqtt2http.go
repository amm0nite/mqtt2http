package main

import (
	"mqtt2http/api"
	"mqtt2http/hooks"
	"mqtt2http/lib"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/listeners"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var err error

	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create the new MQTT Server.
	server := mqtt.New(nil)

	// Create HTTP Client
	err = godotenv.Load()
	if err != nil {
		server.Log.Warn().Err(err).Msg("Failed to read .env file")
	}

	tcpAddr := getEnv("MQTT2HTTP_MQTT_LISTEN_ADDRESS", ":1883")
	httpAddr := getEnv("MQTT2HTTP_HTTP_LISTEN_ADDRESS", ":8080")
	authorizeURL := getEnv("MQTT2HTTP_AUTHORIZE_URL", "http://example.com")
	publishURL := getEnv("MQTT2HTTP_PUBLISH_URL", "http://example.com/{topic}")
	contentType := getEnv("MQTT2HTTP_CONTENT_TYPE", "application/octet-stream")
	topicHeader := getEnv("MQTT2HTTP_TOPIC_HEADER", "X-Topic")
	metricsHttpAddr := getEnv("MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS", ":9090")

	client := &lib.Client{
		Server:       server,
		AuthorizeURL: authorizeURL,
		PublishURL:   publishURL,
		ContentType:  contentType,
		TopicHeader:  topicHeader,
	}

	// Setup auth hook
	authHook := &hooks.AuthHook{Client: client}
	err = server.AddHook(authHook, nil)
	if err != nil {
		server.Log.Error().Err(err).Msg("Failed to add auth hook")
	}

	// Setup publish hook
	publishHook := &hooks.PublishHook{Client: client}
	err = server.AddHook(publishHook, map[string]any{})
	if err != nil {
		server.Log.Error().Err(err).Msg("Failed to add publish hook")
	}

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP("t1", tcpAddr, nil)
	err = server.AddListener(tcp)
	if err != nil {
		server.Log.Error().Err(err).Msg("Failed to add TCP listener")
	}

	// Start
	err = server.Serve()
	if err != nil {
		server.Log.Error().Err(err).Msg("Failed to start server")
	}

	// Handle signals
	go func() {
		sig := <-sigs
		server.Log.Info().Msg(sig.String())
		done <- true
	}()

	// HTTP server
	go func() {
		server.Log.Info().Str("Addr", httpAddr).Msg("Starting API HTTP server")

		controller := api.CreateController(server, client)

		mux := http.NewServeMux()
		mux.HandleFunc("/", controller.RootHandler())
		mux.HandleFunc("/publish", controller.PublishHandler())

		err := http.ListenAndServe(httpAddr, mux)
		if err != nil {
			server.Log.Error().Err(err).Msg("API HTTP server error")
		}
	}()

	// Metrics HTTP server
	go (func() {
		server.Log.Info().Str("Addr", metricsHttpAddr).Msg("Starting metrics HTTP server")

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		err := http.ListenAndServe(metricsHttpAddr, mux)
		if err != nil {
			server.Log.Error().Err(err).Msg("Metrics HTTP server error")
		}
	})()

	server.Log.Info().Msg("awaiting signal")
	<-done
	server.Log.Info().Msg("exiting")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
