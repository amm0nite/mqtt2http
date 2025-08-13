package main

import (
	"log/slog"
	"mqtt2http/broker"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Info("Did not load .env file", "err", err)
	}

	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	config := &broker.BrokerConfig{
		TCPAddr:         getEnv("MQTT2HTTP_MQTT_LISTEN_ADDRESS", ":1883"),
		HTTPAddr:        getEnv("MQTT2HTTP_HTTP_LISTEN_ADDRESS", ":8080"),
		AuthorizeURL:    getEnv("MQTT2HTTP_AUTHORIZE_URL", "http://example.com"),
		PublishURL:      getEnv("MQTT2HTTP_PUBLISH_URL", "http://example.com/{topic}"),
		ContentType:     getEnv("MQTT2HTTP_CONTENT_TYPE", "application/octet-stream"),
		TopicHeader:     getEnv("MQTT2HTTP_TOPIC_HEADER", "X-Topic"),
		MetricsHTTPAddr: getEnv("MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS", ":9090"),
		RoutesFilePath:  getEnv("MQTT2HTTP_ROUTES_FILE_PATH", "routes.yaml"),
		APIPassword:     getEnv("MQTT2HTTP_API_PASSWORD", uuid.NewString()),
	}

	broker := broker.NewBroker(config)
	err = broker.Start()
	if err != nil {
		slog.Error("Failed to start", "err", err)
		panic(err)
	}

	// Handle signals
	go func() {
		sig := <-sigs
		slog.Info("Signal received", "signal", sig.String())
		broker.Close()
		done <- true
	}()

	<-done
	slog.Info("Exiting")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
