package lib

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	publishCounter      *prometheus.CounterVec
	authenticateCounter *prometheus.CounterVec
}

func NewMetrics() (*Metrics, error) {
	var err error
	metrics := &Metrics{}

	metrics.publishCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Subsystem: "publish",
			Name:      "count",
		},
		[]string{"topic", "url", "code"},
	)
	err = prometheus.Register(metrics.publishCounter)
	if err != nil {
		return nil, err
	}

	metrics.authenticateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mqtt2http",
			Subsystem: "authenticate",
			Name:      "count",
		},
		[]string{"url", "code"},
	)
	err = prometheus.Register(metrics.authenticateCounter)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
