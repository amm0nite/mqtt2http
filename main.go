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
	tcpAddr := ":1883"

	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create the new MQTT Server.
	server := mqtt.New(nil)

	// Setup auth hook
	authHook := &AuthHook{}
	_ = server.AddHook(authHook, nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP("t1", tcpAddr, nil)
	err := server.AddListener(tcp)
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
