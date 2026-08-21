package prom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Query(ctx context.Context, query string) ([]byte, error) {
	params := url.Values{}
	params.Set("query", query)
	return c.get(ctx, "/api/v1/query", params)
}

func (c *Client) QueryRange(ctx context.Context, query string, duration time.Duration, step time.Duration) ([]byte, error) {
	end := time.Now()
	start := end.Add(-duration)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", strconv.FormatFloat(step.Seconds(), 'f', -1, 64))
	return c.get(ctx, "/api/v1/query_range", params)
}

func (c *Client) get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	endpoint := c.baseURL + path + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query prometheus: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned %d: %s", resp.StatusCode, body)
	}

	return body, nil
}
