package dto

import "errors"

// SpeakingTaskResponse represents the AI-generated task for speaking practice.
type SpeakingTaskResponse struct {
	Topic               string   `json:"topic"`
	Prompt              string   `json:"prompt"`
	Scenario            string   `json:"scenario,omitempty"`
	Instructions        []string `json:"instructions"`
	Hints               []string `json:"hints,omitempty"`
	SuggestedVocabulary []string `json:"suggested_vocabulary,omitempty"`
	DurationSeconds     int      `json:"duration_seconds"`
}

func (r *SpeakingTaskResponse) Validate() error {
	if r.Prompt == "" {
		return errors.New("missing prompt")
	}
	if len(r.Instructions) == 0 {
		return errors.New("missing instructions")
	}
	if r.DurationSeconds <= 0 || r.DurationSeconds > 300 {
		return errors.New("duration must be between 1 and 300 seconds")
	}
	// Scenario requires a scenario
	// Type checking is done at business level, but let's just do schema check here.
	return nil
}

// SpeakingEvaluationResponse represents the STT and evaluation from AI.
type SpeakingEvaluationResponse struct {
	Transcript string `json:"transcript"`
	Score      int    `json:"score"`
	Fluency    struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	} `json:"fluency"`
	Grammar struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	} `json:"grammar"`
	Vocabulary struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	} `json:"vocabulary"`
	Pronunciation struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	} `json:"pronunciation"`
	TaskCompletion struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	} `json:"task_completion"`
	OverallFeedback   string        `json:"overall_feedback"`
	Strengths         []string      `json:"strengths"`
	Improvements      []string      `json:"improvements"`
	NewUsedVocabulary []NewUsedWord `json:"new_used_vocabulary,omitempty"`
}

type NewUsedWord struct {
	Word   string `json:"word"`
	Reason string `json:"reason"`
}
