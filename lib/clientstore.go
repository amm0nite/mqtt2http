package lib

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ClientStore struct {
	clients map[string]*Client
	mutex   sync.RWMutex
	metrics *Metrics
}

func NewClientStore(metrics *Metrics) *ClientStore {
	hub := &ClientStore{metrics: metrics}
	hub.clients = make(map[string]*Client)

	return hub
}

func (s *ClientStore) Enter(id string, username string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, known := s.clients[id]
	if !known {
		s.clients[id] = NewClient(id, username)
		s.metrics.sessionGauge.Inc()
	}
}

func (s *ClientStore) Leave(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clients, id)
	s.metrics.sessionGauge.Dec()
}

func (s *ClientStore) Subscribe(id string, topics []string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, ok := s.clients[id]
	if !ok {
		return
	}

	for _, topic := range topics {
		found := false
		for _, sub := range client.Subscribtions {
			if sub == topic {
				found = true
				break
			}
		}
		if !found {
			client.Subscribtions = append(client.Subscribtions, topic)
		}
		labels := prometheus.Labels{"topic": topic}
		s.metrics.subscribeCounter.With(labels).Inc()
	}

	client.LastActivityAt = time.Now()
}

func (s *ClientStore) Publish(id string, topic string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, ok := s.clients[id]
	if !ok {
		return
	}

	value := client.Publications[topic]
	client.Publications[topic] = value + 1
	client.LastActivityAt = time.Now()

	labels := prometheus.Labels{"topic": topic}
	s.metrics.publishCounter.With(labels).Inc()
}

func (s *ClientStore) Export() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}

	return json.Marshal(clients)
}
