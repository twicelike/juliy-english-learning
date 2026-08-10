package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HuggingFaceClient implements ai.AIClient using the HuggingFace Router API (OpenAI compatible).
type HuggingFaceClient struct {
	token      string
	model      string
	httpClient *http.Client
}

// NewHuggingFaceClient creates a new HuggingFace AI client.
func NewHuggingFaceClient(token, model string) (*HuggingFaceClient, error) {
	if model == "" {
		model = "deepseek-ai/DeepSeek-V4-Flash:novita"
	}
	return &HuggingFaceClient{
		token:      token,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

type hfRequest struct {
	Model          string  `json:"model"`
	Messages       []hfMsg `json:"messages"`
	ResponseFormat *hfFmt  `json:"response_format,omitempty"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      int     `json:"max_tokens"`
	Stream         bool    `json:"stream"`
}

type hfMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hfFmt struct {
	Type string `json:"type"`
}

type hfResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GenerateJSON sends a prompt to HuggingFace and unmarshals the JSON response into dst.
func (c *HuggingFaceClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	// DeepSeek-V4-Flash over HF usually handles formatting reasonably well,
	// but we explicitly tell it to output JSON in the prompt as well (handled by generator).
	reqBody := hfRequest{
		Model: c.model,
		Messages: []hfMsg{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &hfFmt{Type: "json_object"},
		Temperature:    0.2, // low temp for deterministic JSON
		MaxTokens:      2048,
		Stream:         false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("huggingface: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://router.huggingface.co/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("huggingface: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("huggingface: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("huggingface: API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var hfResp hfResponse
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return fmt.Errorf("huggingface: decode response: %w", err)
	}

	if len(hfResp.Choices) == 0 || hfResp.Choices[0].Message.Content == "" {
		return fmt.Errorf("huggingface: empty response")
	}

	rawJSON := hfResp.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(rawJSON), dst); err != nil {
		return fmt.Errorf("huggingface: unmarshal json content: %w (raw: %.200s)", err, rawJSON)
	}

	return nil
}
