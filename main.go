package main

import (
	"mqtt2http/hooks"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/listeners"
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
	authHook := &hooks.AuthHook{}
	err = server.AddHook(authHook, nil)
	if err != nil {
		server.Log.Error().Err(err)
	}

	// Setup publish hook
	publishHook := &hooks.PublishHook{}
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
