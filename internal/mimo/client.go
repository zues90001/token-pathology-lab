// Package mimo provides the shared MiMo chat-completions client used by all
// MiMo-driven agents in the Pathology Lab.
package mimo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Ledger is satisfied by storage.Diagnostics or storage.Lab; the client
// records every call so daily token aggregates are accurate.
type Ledger interface {
	RecordAnalyserCall(ctx context.Context, agent string, total int)
}

// ChatRequest is the minimal call shape used by every MiMo agent.
type ChatRequest struct {
	Agent    string
	System   string
	User     string
	Temp     float32
	MaxToks  int
	JSONMode bool
}

// Client is a minimal MiMo chat-completions client.
type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
	Ledger     Ledger
}

// New returns a Client wired to env defaults.
func New(ledger Ledger) (*Client, error) {
	key := os.Getenv("MIMO_API_KEY")
	if key == "" {
		return nil, errors.New("MIMO_API_KEY missing")
	}
	base := os.Getenv("MIMO_API_BASE")
	if base == "" {
		base = "https://platform.xiaomimimo.com"
	}
	return &Client{
		BaseURL:    base,
		APIKey:     key,
		Model:      envOr("MIMO_MODEL", "mimo-7b-rl"),
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Ledger:     ledger,
	}, nil
}

// Chat invokes MiMo and records the token usage to the ledger.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (string, error) {
	body := map[string]any{
		"model":       c.Model,
		"temperature": req.Temp,
		"max_tokens":  req.MaxToks,
		"messages": []map[string]string{
			{"role": "system", "content": req.System},
			{"role": "user", "content": req.User},
		},
	}
	if req.JSONMode {
		body["response_format"] = map[string]string{"type": "json_object"}
	}
	buf, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(
		ctx, "POST",
		c.BaseURL+"/v1/chat/completions",
		bytes.NewReader(buf),
	)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("mimo %d: %s", resp.StatusCode, raw)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			Total int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if c.Ledger != nil {
		c.Ledger.RecordAnalyserCall(ctx, req.Agent, parsed.Usage.Total)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("empty choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
