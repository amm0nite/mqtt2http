package hooks

import (
	"bytes"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type PublishHook struct {
	mqtt.HookBase
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

	pkx := pk
	if string(pk.Payload) == "hello" {
		pkx.Payload = []byte("hello world")
		h.Log.Info().Str("client", cl.ID).Str("payload", string(pkx.Payload)).Msg("received modified packet from client")
	}

	return pkx, nil
}
