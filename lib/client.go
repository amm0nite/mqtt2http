package lib

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/mochi-co/mqtt/v2"
)

type Client struct {
	Server       *mqtt.Server
	AuthorizeURL string
	PublishURL   string
}

func (c *Client) Authorize(username string, password string) (bool, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", c.AuthorizeURL, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	success := res.StatusCode == 200 || res.StatusCode == 201
	if !success {
		return false, fmt.Errorf("request failed with status code %d", res.StatusCode)
	}

	return true, nil
}

func (c *Client) Publish(topic string, payload []byte) error {
	reader := bytes.NewReader(payload)
	_, err := http.Post(c.PublishURL, "application/octet-stream", reader)
	return err
}
