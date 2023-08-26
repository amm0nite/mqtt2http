package api

import (
	"encoding/json"
	"io"
	"mqtt2http/lib"
	"net/http"

	"github.com/mochi-co/mqtt/v2"
)

type Controller struct {
	server *mqtt.Server
	client *lib.Client
}

func CreateController(server *mqtt.Server, client *lib.Client) *Controller {
	controller := &Controller{server: server, client: client}
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
		username, password, ok := request.BasicAuth()

		if !ok {
			writer.WriteHeader(http.StatusBadRequest)
			io.WriteString(writer, "Missing basic auth")
			return
		}

		authorized, err := c.client.Authorize(username, password)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			io.WriteString(writer, err.Error())
			return
		}
		if !authorized {
			writer.WriteHeader(http.StatusForbidden)
			io.WriteString(writer, "Forbidden")
			return
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

		c.server.Log.Info().Str("topic", topic).Msg("publish message")
		c.server.Publish(topic, message, false, 0)
		writer.Write(message)
	}
}
