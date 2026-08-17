package prom

import (
	"context"
	"time"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
	}
}

func (c *Client) Query(ctx context.Context, query string) ([]byte, error) {
	return nil, nil
}

func (c *Client) QueryRange(ctx context.Context, query string, duration time.Duration, step time.Duration) ([]byte, error) {
	return nil, nil
}
