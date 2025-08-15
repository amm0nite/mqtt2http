package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type PublishHook struct {
	mqtt.HookBase
	Client     *lib.Client
	DefaultURL string
	Routes     []lib.Route
}

func (h *PublishHook) ID() string {
	return "mqtt2http-publish"
}

func (h *PublishHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *PublishHook) Init(config any) error {
	h.Log.Debug("Initialized")
	return nil
}

func (h *PublishHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("Received from client", "client", cl.ID, "topic", pk.TopicName, "payload", string(pk.Payload))

	matched := false
	for _, route := range h.Routes {
		ok, err := route.Match(pk.TopicName)
		if err != nil {
			h.Log.Error("Error while matching route pattern with topic", "err", err, "name", route.Name)
		}
		if ok {
			matched = true
			h.Log.Debug("Matched route", "topic", pk.TopicName, "name", route.Name)
			err := h.Client.Publish(route.URL, pk.TopicName, pk.Payload)
			if err != nil {
				h.Log.Error("Failed to post on publish", "err", err, "URL", route.URL)
			}
			break
		}
	}

	if !matched {
		h.Log.Info("No route match", "topic", pk.TopicName)
	}

	return pk, nil
}
