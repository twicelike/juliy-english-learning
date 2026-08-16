package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

func (g *GroqClient) transcribeAudio(ctx context.Context, audioBytes []byte, _ string) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Groq requires a filename. Extension doesn't matter too much but webm/mp4 are standard.
	fw, err := w.CreateFormFile("file", "audio.webm")
	if err != nil {
		return "", err
	}
	fw.Write(audioBytes)

	w.WriteField("model", "whisper-large-v3")
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/audio/transcriptions", &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq audio api error %d: %s", resp.StatusCode, string(respBytes))
	}

	var res struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Text, nil
}

func (g *GroqClient) GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error {
	// 1. Speech-to-Text via Whisper
	transcript, err := g.transcribeAudio(ctx, data, mimeType)
	if err != nil {
		return fmt.Errorf("groq multimodal STT failed: %w", err)
	}

	// 2. Augment prompt
	augmentedPrompt := prompt + fmt.Sprintf(`

[USER'S AUDIO TRANSCRIPT]
%s

IMPORTANT: 
Since you are only evaluating a text transcript and cannot hear the audio, you MUST NOT infer pronunciation quality from transcription errors.
Set the "pronunciation" score to 0 or neutral, and in the feedback state: "Pronunciation cannot be directly evaluated from transcript-only input."
For the "transcript" field in the JSON response, use the exact transcript provided above.
`, transcript)

	// 3. Fallback to standard Text JSON Generation
	return g.GenerateJSON(ctx, augmentedPrompt, schema, dst)
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