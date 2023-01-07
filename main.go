package main

import (
	"bytes"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/listeners"
	"github.com/mochi-co/mqtt/v2/packets"
)

func main() {
	var err error
	tcpAddr := ":1883"

	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create the new MQTT Server.
	server := mqtt.New(nil)

	// Setup auth hook
	authHook := &AuthHook{}
	err = server.AddHook(authHook, nil)
	if err != nil {
		server.Log.Error().Err(err)
	}

	// Setup publish hook
	publishHook := &PublishHook{}
	err = server.AddHook(publishHook, map[string]any{})
	if err != nil {
		server.Log.Error().Err(err)
	}

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP("t1", tcpAddr, nil)
	err = server.AddListener(tcp)
	if err != nil {
		server.Log.Error().Err(err)
	}

	// Start
	err = server.Serve()
	if err != nil {
		server.Log.Error().Err(err)
	}

	// Handle signals
	go func() {
		sig := <-sigs
		server.Log.Info().Msg(sig.String())
		done <- true
	}()

	server.Log.Info().Msg("awaiting signal")
	<-done
	server.Log.Info().Msg("exiting")
}

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
