package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// OpenRouterClient implements ai.AIClient using the OpenRouter API (OpenAI compatible).
type OpenRouterClient struct {
	apiKeys    []string
	keyIndex   uint32
	model      string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter AI client.
func NewOpenRouterClient(apiKey, model string) (*OpenRouterClient, error) {
	if model == "" {
		model = "google/gemini-2.5-flash"
	}
	
	// Support multiple comma-separated keys for backup fallback
	var keys []string
	for _, k := range strings.Split(apiKey, ",") {
		if trimmed := strings.TrimSpace(k); trimmed != "" {
			keys = append(keys, trimmed)
		}
	}
	if len(keys) == 0 {
		keys = []string{""} // Fallback to empty string to avoid panic
	}

	return &OpenRouterClient{
		apiKeys:    keys,
		model:      model,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}, nil
}

type openRouterRequest struct {
	Model          string          `json:"model"`
	Messages       []openRouterMsg `json:"messages"`
	ResponseFormat openRouterFmt   `json:"response_format,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
}

type openRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterFmt struct {
	Type string `json:"type"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *OpenRouterClient) getNextKey() string {
	idx := atomic.AddUint32(&c.keyIndex, 1)
	return c.apiKeys[(idx-1)%uint32(len(c.apiKeys))]
}

// GenerateJSON sends a prompt to OpenRouter and unmarshals the JSON response into dst.
func (c *OpenRouterClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []openRouterMsg{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: openRouterFmt{Type: "json_object"},
		Temperature:    0.2, // low temp for deterministic JSON
		MaxTokens:      4000,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("openrouter: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("openrouter: create request: %w", err)
	}

	currentKey := c.getNextKey()
	req.Header.Set("Authorization", "Bearer "+currentKey)
	req.Header.Set("Content-Type", "application/json")
	// Optional OpenRouter headers for analytics:
	req.Header.Set("HTTP-Referer", "http://localhost:3000")
	req.Header.Set("X-Title", "Vocabit")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openrouter: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openrouter: API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var orResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&orResp); err != nil {
		return fmt.Errorf("openrouter: decode response: %w", err)
	}

	if len(orResp.Choices) == 0 || orResp.Choices[0].Message.Content == "" {
		return fmt.Errorf("openrouter: empty response")
	}

	rawJSON := orResp.Choices[0].Message.Content
	rawJSON = extractJSON(rawJSON)

	if err := json.Unmarshal([]byte(rawJSON), dst); err != nil {
		return fmt.Errorf("openrouter: unmarshal json content: %w (raw: %.200s)", err, rawJSON)
	}

	return nil
}

func (o *OpenRouterClient) GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error {
	return fmt.Errorf("openrouter: multimodal generation not supported")
}

func (c *OpenRouterClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []openRouterMsg{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openrouter: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("openrouter: create request: %w", err)
	}

	currentKey := c.getNextKey()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+currentKey)
	req.Header.Set("HTTP-Referer", "https://github.com/english")
	req.Header.Set("X-Title", "English Learning App")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter: http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var orResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&orResp); err != nil {
		return "", fmt.Errorf("openrouter: decode response: %w", err)
	}

	if len(orResp.Choices) == 0 || orResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("openrouter: empty response")
	}

	return cleanResponseText(orResp.Choices[0].Message.Content), nil
}
