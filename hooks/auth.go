package hooks

import (
	"bytes"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type AuthHook struct {
	mqtt.HookBase
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
	return string(cl.Properties.Username) == "test"
}

func (h *AuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}
