package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient implements ai.AIClient using Google Gemini.
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a new Gemini AI client.
// model is the model name (e.g. "gemini-2.0-flash").
func NewGeminiClient(ctx context.Context, apiKey, model string) (*GeminiClient, error) {
	if model == "" {
		model = "gemini-3.5-flash"
	}
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}
	return &GeminiClient{client: c, model: model}, nil
}

// Close releases the underlying client connection.
func (g *GeminiClient) Close() {
	g.client.Close()
}

// GenerateJSON sends prompt to Gemini and unmarshals the JSON response into dst.
func (g *GeminiClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	m := g.client.GenerativeModel(g.model)
	m.ResponseMIMEType = "application/json"

	resp, err := m.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("gemini: generate content: %w", err)
	}

	return g.parseJSONResponse(resp, dst)
}

func (g *GeminiClient) GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error {
	m := g.client.GenerativeModel(g.model)
	m.ResponseMIMEType = "application/json"

	resp, err := m.GenerateContent(ctx, genai.Text(prompt), genai.Blob{
		MIMEType: mimeType,
		Data:     data,
	})
	if err != nil {
		return fmt.Errorf("gemini: multimodal generate content: %w", err)
	}

	return g.parseJSONResponse(resp, dst)
}

func (g *GeminiClient) parseJSONResponse(resp *genai.GenerateContentResponse, dst any) error {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return fmt.Errorf("gemini: empty response")
	}

	var raw string
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			raw += string(t)
		}
	}

	raw = extractJSON(raw)

	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return fmt.Errorf("gemini: unmarshal response: %w (raw: %.200s)", err, raw)
	}
	return nil
}

func (g *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	model := g.client.GenerativeModel(g.model)
	// model.ResponseMIMEType = "application/json" // OMITTED for text

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini: generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}

	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return "", fmt.Errorf("gemini: unexpected response type")
	}

	return cleanResponseText(string(text)), nil
}
