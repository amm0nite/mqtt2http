package hooks

import (
	"bytes"
	"io"
	"mqtt2http/lib"
	"os"

	"github.com/goccy/go-yaml"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type PublishHook struct {
	mqtt.HookBase
	Client         *lib.Client
	DefaultURL     string
	RoutesFilePath string
	routes         []lib.Route
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
	h.Log.Info("Initialized")

	err := h.loadRoutes()
	if err != nil {
		h.Log.Info("No routes loaded, defaulting to publish URL", "err", err)
		h.routes = []lib.Route{
			{
				Name:    "default",
				Pattern: ".*",
				URL:     h.DefaultURL,
			},
		}
	}

	return nil
}

func (h *PublishHook) loadRoutes() error {
	routesFile, err := os.Open(h.RoutesFilePath)
	if err != nil {
		h.Log.Info("Failed to open routes file", "err", err)
		return err
	}

	routesData, err := io.ReadAll(routesFile)
	if err != nil {
		h.Log.Error("Failed to read routes file", "err", err)
		return err
	}

	err = yaml.Unmarshal(routesData, h.routes)
	if err != nil {
		h.Log.Error("Failed to parse routes", "err", err)
		return err
	}

	return nil
}

func (h *PublishHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	h.Log.Info("Received from client", "client", cl.ID, "topic", pk.TopicName, "payload", string(pk.Payload))

	for _, route := range h.routes {
		ok, err := route.Match(pk.TopicName)
		if err != nil {
			h.Log.Error("Error while matching route pattern with topic", "err", err, "name", route.Name)
		}
		if ok {
			h.Log.Debug("Matched route", "topic", pk.TopicName, "name", route.Name)
			err := h.Client.Publish(route.URL, pk.TopicName, pk.Payload)
			if err != nil {
				h.Log.Error("Failed to post on publish", "err", err, "URL", route.URL)
			}
			break
		}
	}

	return pk, nil
}
