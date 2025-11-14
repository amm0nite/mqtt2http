package lib

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	publishCounter      *prometheus.CounterVec
	authenticateCounter *prometheus.CounterVec
	sessionGauge        prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	metrics := &Metrics{}

	metrics.publishCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "publish_count",
		},
		[]string{"topic", "url", "code"},
	)

	metrics.authenticateCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "authenticate_count",
		},
		[]string{"url", "code"},
	)

	metrics.sessionGauge = promauto.With(reg).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mqtt2http",
			Name:      "sessions",
		},
	)

	return metrics
}
