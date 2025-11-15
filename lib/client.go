package lib

import "time"

type Client struct {
	ID             string           `json:"id"`
	Username       string           `json:"username"`
	Subscribtions  []string         `json:"subscriptions"`
	Publications   map[string]int64 `json:"publications"`
	ConnectedAt    time.Time        `json:"connected_at"`
	LastActivityAt time.Time        `json:"last_activity_at"`
}

func NewClient(id string, username string) *Client {
	client := &Client{ID: id, Username: username}
	client.Publications = make(map[string]int64)
	client.ConnectedAt = time.Now()
	client.LastActivityAt = time.Now()
	return client
}
