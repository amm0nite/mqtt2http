package lib

import (
	"bytes"
	"net/http"

	"github.com/mochi-co/mqtt/v2"
)

type Client struct {
	Server       *mqtt.Server
	AuthorizeURL string
	PublishURL   string
}

func (c *Client) Authorize(username string, password string) bool {
	client := &http.Client{}

	req, err := http.NewRequest("GET", c.AuthorizeURL, nil)
	if err != nil {
		c.Server.Log.Error().Err(err)
		return false
	}

	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		c.Server.Log.Error().Err(err)
		return false
	}

	return res.StatusCode == 200 || res.StatusCode == 201
}

func (c *Client) Publish(topic string, payload []byte) error {
	reader := bytes.NewReader(payload)
	_, err := http.Post(c.PublishURL, "application/octet-stream", reader)
	return err
}
