package dto

import (
	"errors"
	"strings"
)

// GrammarMistake describes a single grammatical error.
type GrammarMistake struct {
	Wrong       string `json:"wrong"`
	Correct     string `json:"correct"`
	Reason      string `json:"reason"`
	Explanation string `json:"explanation"`
}

// GrammarResponse is returned by AITutorService.CheckGrammar.
type GrammarResponse struct {
	Score         int              `json:"score"` // 0–100
	CorrectedText string           `json:"correctedText"`
	Mistakes      []GrammarMistake `json:"mistakes"`
	Tips          []string         `json:"tips"`
}

func (r *GrammarResponse) Validate() error {
	if r.Score < 0 || r.Score > 100 {
		return errors.New("AI grammar: score out of range [0, 100]")
	}
	if strings.TrimSpace(r.CorrectedText) == "" {
		return errors.New("AI grammar: correctedText is empty")
	}
	return nil
}

// VocabSuggestion is a single vocabulary improvement suggestion.
type VocabSuggestion struct {
	Original  string `json:"original"`
	Suggested string `json:"suggested"`
	Reason    string `json:"reason"`
	Example   string `json:"example"`
}

// VocabSuggestionResponse is returned by AITutorService.SuggestWritingVocab.
type VocabSuggestionResponse struct {
	OriginalText  string            `json:"originalText"`
	RewrittenText string            `json:"rewrittenText"`
	Suggestions   []VocabSuggestion `json:"suggestions"`
}

func (r *VocabSuggestionResponse) Validate() error {
	if strings.TrimSpace(r.RewrittenText) == "" {
		return errors.New("AI vocab: rewrittenText is empty")
	}
	return nil
}
