package lib

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/mochi-co/mqtt/v2"
)

type Client struct {
	Server       *mqtt.Server
	AuthorizeURL string
	PublishURL   string
}

func (c *Client) Authorize(username string, password string) (bool, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", c.AuthorizeURL, nil)
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
	publishURL := strings.Replace(c.PublishURL, "{topic}", topic, 1)
	reader := bytes.NewReader(payload)

	_, err := http.Post(publishURL, "application/octet-stream", reader)
	return err
}
