package broker

import (
	"io"
	"log/slog"
	"mqtt2http/lib"
	"os"

	"github.com/goccy/go-yaml"
)

type BrokerConfig struct {
	TCPAddr         string
	HTTPAddr        string
	AuthorizeURL    string
	PublishURL      string
	ContentType     string
	TopicHeader     string
	MetricsHTTPAddr string
	RoutesFilePath  string
	APIPassword     string
	Routes          []lib.Route
}

func (c *BrokerConfig) Load() {
	err := c.loadRoutes()
	if err != nil {
		slog.Info("No routes loaded", "err", err)
	}

	if len(c.Routes) == 0 && c.PublishURL != "" {
		slog.Info("Adding default route", "url", c.PublishURL)
		c.Routes = []lib.Route{
			{
				Name:    "default",
				Pattern: ".*",
				URL:     c.PublishURL,
			},
		}
	}
}

func (c *BrokerConfig) loadRoutes() error {
	routesFile, err := os.Open(c.RoutesFilePath)
	if err != nil {
		slog.Info("Failed to open routes file", "err", err)
		return err
	}

	routesData, err := io.ReadAll(routesFile)
	if err != nil {
		slog.Error("Failed to read routes file", "err", err)
		return err
	}

	err = yaml.Unmarshal(routesData, c.Routes)
	if err != nil {
		slog.Error("Failed to parse routes", "err", err)
		return err
	}

	return nil
}
