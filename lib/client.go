package lib

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mochi-co/mqtt/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
	Server       *mqtt.Server
	AuthorizeURL string
	PublishURL   string
	ContentType  string
	TopicHeader  string
	Metrics      *Metrics
}

var ClientTimeout = time.Duration(5) * time.Second

func (c *Client) Authorize(username string, password string) (bool, error) {
	client := &http.Client{Timeout: ClientTimeout}

	req, err := http.NewRequest("POST", c.AuthorizeURL, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	c.Metrics.authenticateCounter.With(prometheus.Labels{"code": strconv.Itoa(res.StatusCode)}).Inc()

	success := res.StatusCode == 200 || res.StatusCode == 201
	if !success {
		return false, fmt.Errorf("request failed with status code %d", res.StatusCode)
	}

	return true, nil
}

func (c *Client) Publish(topic string, payload []byte) error {
	publishURL := strings.Replace(c.PublishURL, "{topic}", topic, 1)
	reader := bytes.NewReader(payload)

	client := &http.Client{Timeout: ClientTimeout}

	req, err := http.NewRequest("POST", publishURL, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set(c.TopicHeader, topic)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	labels := prometheus.Labels{
		"code":  strconv.Itoa(res.StatusCode),
		"topic": topic,
	}
	c.Metrics.publishCounter.With(labels).Inc()

	return nil
}
