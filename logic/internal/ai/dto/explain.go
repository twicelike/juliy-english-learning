package dto

import (
	"errors"
	"fmt"
	"strings"
)

// BatchSenseExplanation represents a single definition/context of a word
type BatchSenseExplanation struct {
	SenseIndex     int      `json:"senseIndex"`
	Explanation    string   `json:"explanation"`
	Examples       []string `json:"examples"`
	CommonMistakes []string `json:"commonMistakes"`
	UsageNote      string   `json:"usageNote"`
}

// BatchExplainResponse is returned by AI for a batch of senses.
type BatchExplainResponse struct {
	Explanations []BatchSenseExplanation `json:"explanations"`
}

func (r *BatchExplainResponse) Validate(requestedIndices []int) error {
	if len(r.Explanations) != len(requestedIndices) {
		return fmt.Errorf("AI explain: expected %d explanations, got %d", len(requestedIndices), len(r.Explanations))
	}
	
	// Create map to check for missing/extra indices
	received := make(map[int]bool)
	for _, exp := range r.Explanations {
		received[exp.SenseIndex] = true
	}
	
	for _, idx := range requestedIndices {
		if !received[idx] {
			return fmt.Errorf("AI explain: missing explanation for sense %d", idx)
		}
	}
	
	return nil
}

// NuanceResponse is returned by AITutorService.ExplainNuance.
type NuanceResponse struct {
	CoreMeaning    string   `json:"coreMeaning"`
	SubtleMeaning  string   `json:"subtleMeaning"`
	UsageNotes     []string `json:"usageNotes"`
	ConfusedWith   []string `json:"confusedWith"`
}

func (r *NuanceResponse) Validate() error {
	if strings.TrimSpace(r.CoreMeaning) == "" {
		return errors.New("AI nuance: coreMeaning is empty")
	}
	return nil
}

// CompareResponse is returned by AITutorService.CompareWords.
type CompareResponse struct {
	Word1       string `json:"word1"`
	Word2       string `json:"word2"`
	Similarity  string `json:"similarity"`
	Differences []struct {
		Aspect string `json:"aspect"`
		Word1  string `json:"word1"`
		Word2  string `json:"word2"`
	} `json:"differences"`
	RecommendedUsage string `json:"recommendedUsage"`
}

func (r *CompareResponse) Validate() error {
	if r.Word1 == "" || r.Word2 == "" {
		return errors.New("AI compare: word fields are empty")
	}
	if len(r.Differences) == 0 {
		return errors.New("AI compare: no differences returned")
	}
	return nil
}

// ExamplesResponse is returned by AITutorService.GenerateExamples.
type ExamplesResponse struct {
	Word     string   `json:"word"`
	Examples []string `json:"examples"`
}

func (r *ExamplesResponse) Validate() error {
	if len(r.Examples) < 2 {
		return errors.New("AI examples: fewer than 2 examples returned")
	}
	return nil
}

