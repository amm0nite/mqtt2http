package api

import (
	"encoding/json"
	"io"
	"mqtt2http/lib"
	"net/http"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/system"
)

type Controller struct {
	version  string
	server   *mqtt.Server
	store    *lib.ClientStore
	password string
}

type AppInfo struct {
	Version string       `json:"version"`
	Broker  *system.Info `json:"broker"`
}

func NewController(version string, server *mqtt.Server, store *lib.ClientStore, password string) *Controller {
	return &Controller{version: version, server: server, store: store, password: password}
}

func (c *Controller) RootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := &AppInfo{
			Version: c.version,
			Broker:  c.server.Info.Clone(),
		}

		data, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func (c *Controller) withAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c.password != "" {
			_, password, ok := r.BasicAuth()

			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "Missing basic auth")
				return
			}

			if password != c.password {
				w.WriteHeader(http.StatusForbidden)
				io.WriteString(w, "Forbidden")
				return
			}
		}
		next(w, r)
	}
}

func (c *Controller) PublishHandler() http.HandlerFunc {
	return c.withAuthentication(func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Missing topic")
			return
		}

		defer r.Body.Close()
		message, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
			return
		}

		c.server.Log.Info("Publish message", "topic", topic)
		c.server.Publish(topic, message, false, 0)
		w.Write(message)
	})
}

func (c *Controller) DumpHandler() http.HandlerFunc {
	return c.withAuthentication(func(w http.ResponseWriter, r *http.Request) {
		data, err := c.store.Export()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "failed to export")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
}
