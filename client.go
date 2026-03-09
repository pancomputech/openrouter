package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://openrouter.ai/api/v1"

	ModelAuto    = "openrouter/auto"
	DefaultModel = ModelAuto
)

type Client struct {
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

// Chat sends a chat completion request. On non-2xx responses the status code
// and response body are included in the returned error.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	t0 := time.Now()
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openrouter: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openrouter: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openrouter: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openrouter: status %d: %s", resp.StatusCode, respBody)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("openrouter: decode response: %w", err)
	}
	duration := time.Since(t0)
	chatResp.Latency = Latency{
		Millis:   int(duration / time.Millisecond),
		Duration: duration.String(),
	}

	return &chatResp, nil
}
