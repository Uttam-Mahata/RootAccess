package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Uttam-Mahata/RootAccess/cli/internal/config"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Request(method, path string, body interface{}, response interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to serialize request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, path)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Token != "" {
		// Use both Header and Cookie to be safe (backend currently uses cookie)
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: c.cfg.Token,
		})
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if msg, ok := errResp["error"]; ok {
				return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
			}
			if msg, ok := errResp["message"]; ok {
				return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(path string, response interface{}) error {
	return c.Request("GET", path, nil, response)
}

func (c *Client) Post(path string, body interface{}, response interface{}) error {
	return c.Request("POST", path, body, response)
}
