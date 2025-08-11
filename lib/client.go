package lib

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
	Server      *mqtt.Server
	ContentType string
	TopicHeader string
	Metrics     *Metrics
}

var ClientTimeout = time.Duration(5) * time.Second

func (c *Client) Authorize(url string, username string, password string) (bool, error) {
	client := &http.Client{Timeout: ClientTimeout}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	labels := prometheus.Labels{
		"url":  url,
		"code": strconv.Itoa(res.StatusCode),
	}
	c.Metrics.authenticateCounter.With(labels).Inc()

	success := res.StatusCode == 200 || res.StatusCode == 201
	if !success {
		return false, fmt.Errorf("auth post failed with status code %d", res.StatusCode)
	}

	return true, nil
}

func (c *Client) Publish(url string, topic string, payload []byte) error {
	publishURL := strings.Replace(url, "{topic}", topic, 1)
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
		"url":   url,
		"topic": topic,
		"code":  strconv.Itoa(res.StatusCode),
	}
	c.Metrics.publishCounter.With(labels).Inc()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		return fmt.Errorf("publish post failed with status %d", res.StatusCode)
	}

	return nil
}
