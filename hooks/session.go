package hooks

import (
	"bytes"
	"mqtt2http/lib"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type SessionHook struct {
	mqtt.HookBase
	HTTPClient *lib.HTTPClient
	URL        string
	Hub        *lib.ClientHub
}

func (h *SessionHook) ID() string {
	return "mqtt2http-session"
}

func (h *SessionHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
		mqtt.OnDisconnect,
	}, []byte{b})
}

func (h *SessionHook) Init(config any) error {
	h.Log.Debug("Initialized")
	return nil
}

func (h *SessionHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	username := string(cl.Properties.Username)
	password := string(pk.Connect.Password)

	h.Log.Debug("Client tries to connect", "username", username)
	res, err := h.HTTPClient.Authorize(username, password)
	if err != nil {
		h.Log.Error("Auth request failed", "err", err)
		return false
	}
	if !res {
		h.Log.Info("Auth denied", "client", cl.ID, "username", username)
		return false
	}

	client := &lib.Client{ID: cl.ID, Username: username}
	h.Hub.AddChan <- client

	return true
}

func (h *SessionHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	h.Log.Debug("ACLCheck", "client", cl.ID, "topic", topic, "write", write)
	return true
}

func (h *SessionHook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	h.Log.Debug("Disconnect", "client", cl.ID, "expire", expire)
	h.Hub.RemoveChan <- cl.ID
}
