package dto

import (
	"errors"
	"strings"
)

// ExplainResponse is returned by AITutorService.ExplainWord.
type ExplainResponse struct {
	Summary       string   `json:"summary"`
	Nuance        string   `json:"nuance"`
	CommonMistakes []string `json:"commonMistakes"`
	Examples      []string `json:"examples"`
	Register      string   `json:"register"` // e.g. "formal", "informal", "neutral"
}

func (r *ExplainResponse) Validate() error {
	if strings.TrimSpace(r.Summary) == "" {
		return errors.New("AI explain: summary is empty")
	}
	if len(r.Examples) == 0 {
		return errors.New("AI explain: no examples returned")
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

// MnemonicResponse is returned by AITutorService.CreateMnemonic.
type MnemonicResponse struct {
	Word     string `json:"word"`
	Mnemonic string `json:"mnemonic"`
	Story    string `json:"story"`
	Visual   string `json:"visual"`
}

func (r *MnemonicResponse) Validate() error {
	if strings.TrimSpace(r.Mnemonic) == "" {
		return errors.New("AI mnemonic: mnemonic is empty")
	}
	return nil
}
