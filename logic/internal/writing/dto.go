package writing

// MissionDTO is the data transfer object returned to the frontend.
// It hides internal fields like _id (as string), hash, created_at.
type MissionDTO struct {
	ID               string   `json:"id"`
	Topic            string   `json:"topic"`
	Scenario         string   `json:"scenario"`
	Difficulty       string   `json:"difficulty"`
	Vietnamese       []string `json:"prompt_vi"` // renamed to prompt_vi for clarity
	TargetVocabulary []string `json:"target_vocabulary"`
	GrammarFocus     []string `json:"grammar_focus,omitempty"`
	EstimatedTime    int      `json:"estimated_time"` // in minutes
}

// MapMissionToDTO converts a MongoDB WritingMission model to a safe DTO.
func MapMissionToDTO(m *WritingMission) MissionDTO {
	estimatedTime := 2 // default Easy
	if m.Difficulty == DifficultyMedium {
		estimatedTime = 5
	} else if m.Difficulty == DifficultyHard {
		estimatedTime = 10
	}

	return MissionDTO{
		ID:               m.ID.Hex(),
		Topic:            m.Topic,
		Scenario:         m.Scenario,
		Difficulty:       m.Difficulty,
		Vietnamese:       m.Vietnamese,
		TargetVocabulary: m.TargetVocabulary,
		GrammarFocus:     m.GrammarFocus,
		EstimatedTime:    estimatedTime,
	}
}

// SubmissionRequestDTO is the payload received when a user submits a writing mission.
type SubmissionRequestDTO struct {
	Content string `json:"content" binding:"required,min=5,max=2000"`
}

// EvaluationResponseDTO is the AI evaluation result returned to the frontend.
type EvaluationResponseDTO struct {
	Score             int                  `json:"score"`
	Breakdown         EvaluationBreakdown  `json:"breakdown"`
	CorrectedVersion  string               `json:"corrected_version"`
	DetailedFeedback  string               `json:"detailed_feedback"`
}

type EvaluationBreakdown struct {
	Grammar      CategoryScore `json:"grammar"`
	Vocabulary   CategoryScore `json:"vocabulary"`
	Naturalness  CategoryScore `json:"naturalness"`
	Completeness CategoryScore `json:"completeness"`
}

type CategoryScore struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// DashboardStatsDTO is the summary for the Writing dashboard.
type DashboardStatsDTO struct {
	TotalCompleted   int     `json:"total_completed"`
	AverageScore     float64 `json:"average_score"`
	MasteryMastered  int     `json:"mastery_mastered"`
	MasteryFamiliar  int     `json:"mastery_familiar"`
	MasteryBeginner  int     `json:"mastery_beginner"`
}

