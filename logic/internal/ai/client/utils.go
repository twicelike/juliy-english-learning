package client

import (
	"regexp"
	"strings"
)

var providerPrefixRegex = regexp.MustCompile(`^(?i)(google|anthropic|openai|meta|mistral|cohere|deepseek|xai|grok|llama|gemini|claude|gpt|openrouter|provider|model)\s*:\s*`)

// cleanResponseText removes known provider prefixes from the AI response.
func cleanResponseText(raw string) string {
	raw = strings.TrimSpace(raw)
	return providerPrefixRegex.ReplaceAllString(raw, "")
}

// extractJSON extracts the JSON block from an AI response string.
// It handles cases where the AI appends non-JSON text (like warnings or <think> tags).
func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// Remove <think> blocks
	if idx := strings.Index(raw, "</think>"); idx != -1 {
		raw = raw[idx+len("</think>"):]
		raw = strings.TrimSpace(raw)
	}

	// Remove markdown code blocks
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimPrefix(raw, "```json")
	} else if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
	}
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// Extract JSON block in case of extra text (like "User Safety: safe")
	firstBrace := strings.Index(raw, "{")
	firstBracket := strings.Index(raw, "[")
	firstIdx := -1
	if firstBrace != -1 && firstBracket != -1 {
		if firstBrace < firstBracket {
			firstIdx = firstBrace
		} else {
			firstIdx = firstBracket
		}
	} else if firstBrace != -1 {
		firstIdx = firstBrace
	} else if firstBracket != -1 {
		firstIdx = firstBracket
	}

	lastBrace := strings.LastIndex(raw, "}")
	lastBracket := strings.LastIndex(raw, "]")
	lastIdx := -1
	if lastBrace != -1 && lastBracket != -1 {
		if lastBrace > lastBracket {
			lastIdx = lastBrace
		} else {
			lastIdx = lastBracket
		}
	} else if lastBrace != -1 {
		lastIdx = lastBrace
	} else if lastBracket != -1 {
		lastIdx = lastBracket
	}

	if firstIdx != -1 && lastIdx != -1 && lastIdx >= firstIdx {
		raw = raw[firstIdx : lastIdx+1]
	}

	return raw
}
