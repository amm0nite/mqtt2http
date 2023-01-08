package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type AuthHook struct {
	mqtt.HookBase
	Client *lib.Client
}

func (h *AuthHook) ID() string {
	return "mqtt2http-auth"
}

func (h *AuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
	}, []byte{b})
}

func (h *AuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	username := string(cl.Properties.Username)
	password := string(pk.Connect.Password)
	res, err := h.Client.Authorize(username, password)
	if err != nil {
		h.Log.Error().Err(err).Msg("Auth request failed")
	}
	return res
}

func (h *AuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}
