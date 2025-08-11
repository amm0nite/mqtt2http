package lib

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	publishCounter      *prometheus.CounterVec
	authenticateCounter *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{}

	metrics.publishCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Subsystem: "publish",
			Name:      "count",
		},
		[]string{"topic", "url", "code"},
	)

	metrics.authenticateCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Subsystem: "authenticate",
			Name:      "count",
		},
		[]string{"url", "code"},
	)

	return metrics
}
