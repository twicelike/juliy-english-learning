package learning

// EvaluateLearningEvidence evaluates a WordResult and calculates the final quality
// to be passed to the SM-2 scheduler.
func EvaluateLearningEvidence(result WordResult) LearningAssessment {
	baseQuality := result.FlashcardQuality
	mcPenalty := 0
	fbPenalty := 0

	// Multiple Choice penalty
	if result.MultipleChoiceCorrect != nil && !*result.MultipleChoiceCorrect {
		mcPenalty = 1
	}

	// Fill Blank penalty
	if result.FillBlankCorrect != nil && !*result.FillBlankCorrect {
		fbPenalty = 1
	}

	finalQuality := baseQuality - mcPenalty - fbPenalty

	// Clamp final quality between 1 (Again) and baseQuality
	if finalQuality < 1 {
		finalQuality = 1
	}
	if finalQuality > baseQuality {
		finalQuality = baseQuality
	}

	return LearningAssessment{
		BaseQuality:  baseQuality,
		MCPenalty:    mcPenalty,
		FBPenalty:    fbPenalty,
		FinalQuality: finalQuality,
	}
}
