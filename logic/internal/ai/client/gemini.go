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
		model = "gemini-2.0-flash"
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
// schema is currently unused for Gemini Flash (which reliably returns JSON when
// prompted to do so via the prompt itself). It is kept in the signature to
// satisfy the AIClient interface and enable future use of Gemini's ResponseSchema.
func (g *GeminiClient) GenerateJSON(ctx context.Context, prompt string, _ any, dst any) error {
	m := g.client.GenerativeModel(g.model)

	// Instruct the model to respond with JSON only.
	m.ResponseMIMEType = "application/json"

	resp, err := m.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("gemini: generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return fmt.Errorf("gemini: empty response")
	}

	// Concatenate all text parts.
	var raw string
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			raw += string(t)
		}
	}

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

	return string(text), nil
}
