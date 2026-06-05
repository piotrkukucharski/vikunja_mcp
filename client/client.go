package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type VikunjaClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewVikunjaClient(baseURL, token string) *VikunjaClient {
	return &VikunjaClient{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{},
	}
}

func (c *VikunjaClient) Request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to encode body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	fullURL, err := url.JoinPath(c.BaseURL, path)
	if err != nil {
		return fmt.Errorf("failed to join url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.Token != "" {
		req.Header.Set("Authorization", c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vikunja api error: %d %s", resp.StatusCode, string(b))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}
	}
	return nil
}
