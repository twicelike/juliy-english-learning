package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GroqClient implements ai.AIClient using the Groq API (OpenAI compatible).
type GroqClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGroqClient creates a new Groq AI client.
// model is the model name (e.g. "llama3-8b-8192").
func NewGroqClient(apiKey, model string) (*GroqClient, error) {
	if model == "" {
		model = "llama-3.1-8b-instant"
	}
	return &GroqClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

type groqRequest struct {
	Model          string           `json:"model"`
	Messages       []groqMessage    `json:"messages"`
	ResponseFormat *groqResponseFmt `json:"response_format,omitempty"`
	Temperature    float64          `json:"temperature"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponseFmt struct {
	Type string `json:"type"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateJSON sends a prompt to Groq and unmarshals the JSON response into dst.
func (g *GroqClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	reqBody := groqRequest{
		Model: g.model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &groqResponseFmt{Type: "json_object"},
		Temperature:    0.2, // low temp for deterministic JSON
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("groq: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("groq: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("groq: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("groq: API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var gResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return fmt.Errorf("groq: decode response: %w", err)
	}

	if len(gResp.Choices) == 0 || gResp.Choices[0].Message.Content == "" {
		return fmt.Errorf("groq: empty response")
	}

	rawJSON := gResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(rawJSON), dst); err != nil {
		return fmt.Errorf("groq: unmarshal json content: %w (raw: %.200s)", err, rawJSON)
	}

	return nil
}


func (g *GroqClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	reqBody := groqRequest{
		Model: g.model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
		
		Temperature:    0.2, // low temp for deterministic JSON
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("groq: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("groq: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq: API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var gResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return "", fmt.Errorf("groq: decode response: %w", err)
	}

	if len(gResp.Choices) == 0 || gResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("groq: empty response")
	}

	rawJSON := gResp.Choices[0].Message.Content
	return rawJSON, nil
}