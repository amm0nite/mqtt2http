package lib

import (
	"context"
	"log/slog"
)

type ClientHub struct {
	clients    map[string]*Client
	AddChan    chan *Client
	RemoveChan chan string
	metrics    *Metrics
}

func NewClientHub(metrics *Metrics) *ClientHub {
	hub := &ClientHub{metrics: metrics}
	hub.clients = make(map[string]*Client)
	hub.AddChan = make(chan *Client)
	hub.RemoveChan = make(chan string)
	return hub
}

func (h *ClientHub) Run(ctx context.Context) {
	slog.Info("Start client hub loop")

	for {
		select {
		case client := <-h.AddChan:
			_, known := h.clients[client.ID]
			if !known {
				h.clients[client.ID] = client
				h.metrics.sessionGauge.Inc()
			}
		case id := <-h.RemoveChan:
			delete(h.clients, id)
			h.metrics.sessionGauge.Dec()
		case <-ctx.Done():
			slog.Info("Stop client hub loop")
			return
		}
	}
}
