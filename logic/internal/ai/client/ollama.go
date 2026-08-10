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

// OllamaClient implements ai.AIClient using the native Ollama API.
type OllamaClient struct {
	endpoint   string // e.g., http://localhost:11434
	model      string
	httpClient *http.Client
}

// NewOllamaClient creates a new Ollama AI client.
func NewOllamaClient(endpoint, model string) (*OllamaClient, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	// Remove trailing slash if present
	endpoint = strings.TrimSuffix(endpoint, "/")
	
	if model == "" {
		model = "qwen2.5:7b" // Fallback model
	}
	return &OllamaClient{
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   string          `json:"format"`
	Options  ollamaOptions   `json:"options"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message ollamaMessage `json:"message"`
}

// GenerateJSON sends a prompt to Ollama and unmarshals the JSON response into dst.
func (o *OllamaClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	reqBody := ollamaRequest{
		Model: o.model,
		Messages: []ollamaMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
		Format: "json",
		Options: ollamaOptions{
			Temperature: 0.2, // Low temp for deterministic JSON
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", o.endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama: API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var oResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return fmt.Errorf("ollama: decode response: %w", err)
	}

	rawJSON := oResp.Message.Content
	if rawJSON == "" {
		return fmt.Errorf("ollama: empty response")
	}

	if err := json.Unmarshal([]byte(rawJSON), dst); err != nil {
		return fmt.Errorf("ollama: unmarshal json content: %w (raw: %.200s)", err, rawJSON)
	}

	return nil
}
