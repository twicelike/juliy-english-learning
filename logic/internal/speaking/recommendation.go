package speaking

import (
	"context"
	"fmt"
	"math"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type recommendationService struct {
	repo    Repository
	service Service
}

func NewRecommendationService(repo Repository, service Service) RecommendationService {
	return &recommendationService{
		repo:    repo,
		service: service,
	}
}

func (s *recommendationService) GetRecommendedTask(ctx context.Context, userID string) (*RecommendedTaskResponse, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// 1. Get up to 5 most recent completed attempts with evaluations
	attempts, err := s.repo.GetRecentCompletedAttempts(ctx, uid, 5)
	if err != nil {
		return nil, fmt.Errorf("get recent attempts: %w", err)
	}

	// If not enough data, just return a random beginner task
	if len(attempts) == 0 {
		task, err := s.service.GetRandomTask(ctx, "", "beginner", "quick_practice", "")
		if err != nil || task == nil {
			return nil, fmt.Errorf("fallback task not found")
		}
		return &RecommendedTaskResponse{
			Task: task,
			Reason: &RecommendationReason{
				Skill:           "general",
				Focus:           "getting_started",
				BasedOnSessions: 0,
			},
		}, nil
	}

	// 2. Aggregate evaluations to find weakest skill
	var totalFluency, totalGrammar, totalVocab, totalPronunciation float64

	for _, a := range attempts {
		ev := a.Evaluation
		totalFluency += float64(ev.Fluency)
		totalGrammar += float64(ev.Grammar)
		totalVocab += float64(ev.Vocabulary)
		totalPronunciation += float64(ev.Pronunciation)
	}

	count := float64(len(attempts))
	avgFluency := totalFluency / count
	avgGrammar := totalGrammar / count
	avgVocab := totalVocab / count
	avgPronunciation := totalPronunciation / count

	// Find the minimum
	minScore := math.Min(math.Min(avgFluency, avgGrammar), math.Min(avgVocab, avgPronunciation))

	weakestSkill := "fluency"
	focus := "daily_conversation" // default mapping

	if minScore == avgGrammar {
		weakestSkill = "grammar"
		focus = "past_tense" // simplistic mapping for MVP
	} else if minScore == avgVocab {
		weakestSkill = "vocabulary"
		focus = "advanced_vocabulary"
	} else if minScore == avgPronunciation {
		weakestSkill = "pronunciation"
		focus = "minimal_pairs"
	} else if minScore == avgFluency {
		weakestSkill = "fluency"
		focus = "linking_words"
	}

	// 3. Query a task that addresses the focus
	// We'll try to find a scenario or free_talk task with this focus.
	task, err := s.service.GetRandomTask(ctx, "", "", "scenario", focus)
	if err != nil || task == nil {
		// Fallback without focus if DB misses
		task, err = s.service.GetRandomTask(ctx, "", "", "scenario", "")
		if err != nil || task == nil {
			// Fallback to anything
			task, err = s.service.GetRandomTask(ctx, "", "", "", "")
			if err != nil || task == nil {
				return nil, fmt.Errorf("no task available")
			}
		}
	}

	return &RecommendedTaskResponse{
		Task: task,
		Reason: &RecommendationReason{
			Skill:           weakestSkill,
			Focus:           focus,
			BasedOnSessions: len(attempts),
		},
	}, nil
}
