package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type PublishHook struct {
	mqtt.HookBase
	Client *lib.Client
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
	h.Log.Info("Initialized")
	return nil
}

func (h *PublishHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("Received from client", "client", cl.ID, "topic", pk.TopicName, "payload", string(pk.Payload))
	err := h.Client.Publish(pk.TopicName, pk.Payload)
	return pk, err
}
