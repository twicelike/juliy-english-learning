package dto

import "errors"

// FillBlankResponse is returned by AITutorService.GenerateFillBlank.
// AI selects the most suitable word from the provided list and creates a
// natural sentence with a blank where that word should be placed.
type FillBlankResponse struct {
	// ChosenWord is the vocabulary word that the AI selected.
	ChosenWord string `json:"chosenWord"`
	// POS is the Part-of-Speech of the chosen word (verb, noun…).
	POS string `json:"pos"`
	// Sentence is the fill-in-the-blank sentence. The blank is represented
	// by three underscores: "___".
	Sentence string `json:"sentence"`
	// Answer is the exact form of ChosenWord that fits the blank (may be
	// inflected, e.g. "running" for word "run").
	Answer string `json:"answer"`
	// Hint is an optional short clue for the learner (e.g. "past tense").
	Hint string `json:"hint"`
}

func (r *FillBlankResponse) Validate() error {
	if r.ChosenWord == "" {
		return errors.New("AI fill-blank: chosenWord is empty")
	}
	if r.Sentence == "" {
		return errors.New("AI fill-blank: sentence is empty")
	}
	if r.Answer == "" {
		return errors.New("AI fill-blank: answer is empty")
	}
	return nil
}
