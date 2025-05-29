package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
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

func (h *AuthHook) Init(config any) error {
	h.Log.Info("Initialized")
	return nil
}

func (h *AuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	username := string(cl.Properties.Username)
	password := string(pk.Connect.Password)
	res, err := h.Client.Authorize(username, password)
	if err != nil {
		h.Log.Error("Auth request failed", "err", err)
		return false
	}
	return res
}

func (h *AuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}
