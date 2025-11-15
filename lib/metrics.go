package lib

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	sessionGauge        prometheus.Gauge
	authenticateCounter *prometheus.CounterVec
	publishCounter      *prometheus.CounterVec
	forwardCounter      *prometheus.CounterVec
	subscribeCounter    *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	metrics := &Metrics{}

	metrics.sessionGauge = promauto.With(reg).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "mqtt2http",
			Name:      "sessions",
		},
	)

	metrics.authenticateCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "authenticate_count",
		},
		[]string{"url", "code"},
	)

	metrics.publishCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "publish_count",
		},
		[]string{"topic"},
	)

	metrics.forwardCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "forward_count",
		},
		[]string{"url", "code"},
	)

	metrics.subscribeCounter = promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Name:      "subscribe_count",
		},
		[]string{"topic"},
	)

	return metrics
}
