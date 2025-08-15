package api

import (
	"encoding/json"
	"io"
	"mqtt2http/lib"
	"net/http"

	mqtt "github.com/mochi-mqtt/server/v2"
)

type Controller struct {
	server   *mqtt.Server
	client   *lib.Client
	password string
}

func CreateController(server *mqtt.Server, client *lib.Client, password string) *Controller {
	controller := &Controller{server: server, client: client, password: password}
	return controller
}

func (c *Controller) RootHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		info, _ := json.Marshal(c.server.Info)
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(info)
	}
}

func (c *Controller) PublishHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if c.password != "" {
			_, password, ok := request.BasicAuth()

			if !ok {
				writer.WriteHeader(http.StatusBadRequest)
				io.WriteString(writer, "Missing basic auth")
				return
			}

			if password != c.password {
				writer.WriteHeader(http.StatusForbidden)
				io.WriteString(writer, "Forbidden")
				return
			}
		}

		topic := request.URL.Query().Get("topic")
		if topic == "" {
			writer.WriteHeader(http.StatusBadRequest)
			io.WriteString(writer, "Missing topic")
			return
		}

		defer request.Body.Close()
		message, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			io.WriteString(writer, err.Error())
			return
		}

		c.server.Log.Info("Publish message", "topic", topic)
		c.server.Publish(topic, message, false, 0)
		writer.Write(message)
	}
}
