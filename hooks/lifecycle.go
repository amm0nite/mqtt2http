package hooks

import (
	"bytes"

	mqtt "github.com/mochi-mqtt/server/v2"
)

type LifecycleHook struct {
	mqtt.HookBase
}

func (h *LifecycleHook) ID() string {
	return "mqtt2http-lifecycle"
}

func (h *LifecycleHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnStarted,
		mqtt.OnStopped,
	}, []byte{b})
}

func (h *LifecycleHook) Init(config any) error {
	h.Log.Debug("Initialized")
	return nil
}

func (h *LifecycleHook) OnStarted() {
	h.Log.Debug("OnStarted")
}

func (h *LifecycleHook) OnStopped() {
	h.Log.Debug("OnStopped")
}
