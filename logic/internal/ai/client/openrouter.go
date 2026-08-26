package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenRouterClient implements ai.AIClient using the OpenRouter API (OpenAI compatible).
type OpenRouterClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter AI client.
func NewOpenRouterClient(apiKey, model string) (*OpenRouterClient, error) {
	if model == "" {
		model = "google/gemini-2.5-flash"
	}
	return &OpenRouterClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
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

// GenerateJSON sends a prompt to OpenRouter and unmarshals the JSON response into dst.
func (c *OpenRouterClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []openRouterMsg{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: openRouterFmt{Type: "json_object"},
		Temperature:    0.2, // low temp for deterministic JSON
		MaxTokens:      1000,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("openrouter: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("openrouter: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
	
	// Clean JSON (remove <think> block and markdown formatting)
	rawJSON = strings.TrimSpace(rawJSON)
	if idx := strings.Index(rawJSON, "</think>"); idx != -1 {
		rawJSON = rawJSON[idx+len("</think>"):]
		rawJSON = strings.TrimSpace(rawJSON)
	}
	if strings.HasPrefix(rawJSON, "```json") {
		rawJSON = strings.TrimPrefix(rawJSON, "```json")
	} else if strings.HasPrefix(rawJSON, "```") {
		rawJSON = strings.TrimPrefix(rawJSON, "```")
	}
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("openrouter: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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

	return orResp.Choices[0].Message.Content, nil
}
