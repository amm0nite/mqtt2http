package lib

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const clientTimeout = time.Duration(5) * time.Second

type HTTPClient struct {
	ContentType  string
	TopicHeader  string
	AuthorizeURL string
	Metrics      *Metrics
}

func NewHTTPClient(contentType string, topicHeader string, authorizeURL string, metrics *Metrics) *HTTPClient {
	return &HTTPClient{
		ContentType:  contentType,
		TopicHeader:  topicHeader,
		AuthorizeURL: authorizeURL,
		Metrics:      metrics,
	}
}

func (c *HTTPClient) Authorize(username string, password string) (bool, error) {
	client := &http.Client{Timeout: clientTimeout}

	req, err := http.NewRequest("POST", c.AuthorizeURL, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(username, password)

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	labels := prometheus.Labels{
		"url":  c.AuthorizeURL,
		"code": strconv.Itoa(res.StatusCode),
	}
	c.Metrics.authenticateCounter.With(labels).Inc()

	success := res.StatusCode == 200 || res.StatusCode == 201
	if !success {
		return false, fmt.Errorf("auth post failed with status code %d", res.StatusCode)
	}

	return true, nil
}

func (c *HTTPClient) Publish(url string, topic string, payload []byte) error {
	publishURL := strings.Replace(url, "{topic}", topic, 1)
	reader := bytes.NewReader(payload)

	client := &http.Client{Timeout: clientTimeout}

	req, err := http.NewRequest("POST", publishURL, reader)
	if err != nil {
		return err
	}

	if c.ContentType != "" {
		req.Header.Set("Content-Type", c.ContentType)
	}
	if c.TopicHeader != "" {
		req.Header.Set(c.TopicHeader, topic)
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	labels := prometheus.Labels{
		"url":  url,
		"code": strconv.Itoa(res.StatusCode),
	}
	c.Metrics.forwardCounter.With(labels).Inc()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("publish post failed with status %d", res.StatusCode)
	}

	return nil
}

func (c *HTTPClient) NoMatch(topic string) {
	labels := prometheus.Labels{
		"topic": topic,
	}
	c.Metrics.noMatchCounter.With(labels).Inc()
}
