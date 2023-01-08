package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
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
	h.Log.Info().Msg("initialised")
	return nil
}

func (h *PublishHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info().Str("client", cl.ID).Str("payload", string(pk.Payload)).Msg("received from client")
	err := h.Client.Publish(pk.TopicName, pk.Payload)
	return pk, err
}
