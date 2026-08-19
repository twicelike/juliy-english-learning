package writing

import (
	"context"
	"fmt"

	"github.com/english/backend/internal/evaluation"
	"github.com/english/backend/internal/events"
	"github.com/english/backend/internal/gamification"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// WritingService defines the interface for gameplay features.
type WritingService interface {
	GetNewMission(ctx context.Context, userIDStr string, topic, difficulty string) (*MissionDTO, error)
	GetReviewMission(ctx context.Context, userIDStr string, topic, difficulty string) (*MissionDTO, error)

	GetDashboardStats(ctx context.Context, userIDStr string) (*DashboardStatsDTO, error)

	// SubmitMission evaluates a user's writing submission and updates their progress.
	SubmitMission(ctx context.Context, userIDStr string, missionIDStr string, content string) (*EvaluationResponseDTO, error)
}

type service struct {
	repo      *Repository
	generator *GeneratorService
	gamiSvc   gamification.Service
	evalSvc   evaluation.Service
	publisher events.Publisher
}

func NewService(repo *Repository, generator *GeneratorService, gamiSvc gamification.Service, evalSvc evaluation.Service, publisher events.Publisher) WritingService {
	return &service{
		repo:      repo,
		generator: generator,
		gamiSvc:   gamiSvc,
		evalSvc:   evalSvc,
		publisher: publisher,
	}
}

func (s *service) GetNewMission(ctx context.Context, userIDStr string, topic, difficulty string) (*MissionDTO, error) {
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	mission, err := s.repo.GetNewMission(ctx, userID, topic, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to get new mission: %w", err)
	}

	// Fallback logic if no new mission is found
	if mission == nil {
		mission, err = s.repo.GetReviewMission(ctx, userID, topic, difficulty)
		if err != nil {
			return nil, fmt.Errorf("failed to get fallback review mission: %w", err)
		}
	}

	if mission == nil {
		return nil, fmt.Errorf("no missions available") // Even fallback is empty (DB empty)
	}

	dto := MapMissionToDTO(mission)
	return &dto, nil
}

func (s *service) GetReviewMission(ctx context.Context, userIDStr string, topic, difficulty string) (*MissionDTO, error) {
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	mission, err := s.repo.GetReviewMission(ctx, userID, topic, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to get review mission: %w", err)
	}

	if mission == nil {
		// Fallback to new mission if they haven't completed any reviewable missions yet
		mission, err = s.repo.GetNewMission(ctx, userID, topic, difficulty)
		if err != nil {
			return nil, fmt.Errorf("failed to get fallback new mission: %w", err)
		}
	}

	if mission == nil {
		return nil, fmt.Errorf("no missions available")
	}

	dto := MapMissionToDTO(mission)
	return &dto, nil
}

func (s *service) SubmitMission(ctx context.Context, userIDStr string, missionIDStr string, content string) (*EvaluationResponseDTO, error) {
	if s.generator == nil {
		return nil, fmt.Errorf("AI evaluation is currently unavailable")
	}

	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	missionID, err := bson.ObjectIDFromHex(missionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid mission ID")
	}

	// Get original mission
	mission, err := s.repo.GetMissionByID(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mission: %w", err)
	}
	if mission == nil {
		return nil, fmt.Errorf("mission not found")
	}

	// Call AI
	evalResp, err := s.generator.EvaluateMission(ctx, mission, content)
	if err != nil {
		return nil, fmt.Errorf("AI evaluation failed: %w", err)
	}

	// Ensure score bounds
	if evalResp.Score < 0 {
		evalResp.Score = 0
	}
	if evalResp.Score > 100 {
		evalResp.Score = 100
	}

	// Upsert progress
	progress, err := s.repo.GetProgress(ctx, userID, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user progress: %w", err)
	}

	if progress == nil {
		progress = &UserWritingProgress{
			UserID:    userID,
			MissionID: missionID,
		}
	}

	// Update stats
	if float64(evalResp.Score) > progress.BestScore {
		progress.BestScore = float64(evalResp.Score)
	}
	progress.LastScore = float64(evalResp.Score)
	
	totalAttempts := float64(progress.Attempts)
	progress.AverageScore = ((progress.AverageScore * totalAttempts) + float64(evalResp.Score)) / (totalAttempts + 1)
	progress.Completed = true

	// Mastery Level Logic
	if progress.AverageScore >= 90 {
		progress.MasteryLevel = "Mastered"
	} else if progress.AverageScore >= 70 {
		progress.MasteryLevel = "Familiar"
	} else {
		progress.MasteryLevel = "Beginner"
	}

	if err := s.repo.UpsertProgress(ctx, progress); err != nil {
		return nil, fmt.Errorf("failed to save progress: %w", err)
	}

	// Log the actual submission
	submission := &WritingSubmission{
		ID:               bson.NewObjectID(),
		UserID:           userID,
		MissionID:        missionID,
		UserContent:      content,
		CorrectedVersion: evalResp.CorrectedVersion,
		DetailedFeedback: evalResp.DetailedFeedback,
		Score:            evalResp.Score,
	}
	if err := s.repo.InsertSubmission(ctx, submission); err != nil {
		// Log error but don't fail the request since progress was saved
		fmt.Printf("failed to log submission history: %v\n", err)
	}

	if s.evalSvc != nil {
		ev := &evaluation.Evidence{
			UserID:     userID,
			EventID:    submission.ID.Hex(),
			Skill:      "writing", // From profile.SkillWriting
			Score:      float64(evalResp.Score),
			Weight:     evaluation.WeightWriting,
			Source:     evaluation.SourceWritingLab,
			Metadata:   bson.M{"mission_id": missionID.Hex(), "attempt_id": submission.ID.Hex()},
		}
		go func(e *evaluation.Evidence) {
			bgCtx := context.Background()
			_ = s.evalSvc.RecordEvidence(bgCtx, e)
			
			// Publish profile refresh event
			if s.publisher != nil {
				_ = s.publisher.Publish(bgCtx, events.Event{
					Type:   events.EventTypeProfileRefreshRequested,
					UserID: userIDStr,
					Reason: "writing_mission_completed",
				})
			}
		}(ev)
	}

	if s.gamiSvc != nil {
		// Fire and forget tracking
		go s.gamiSvc.TrackWritingMission(context.Background(), userIDStr, evalResp.Score)
	}

	return evalResp, nil
}

func (s *service) GetDashboardStats(ctx context.Context, userIDStr string) (*DashboardStatsDTO, error) {
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	return s.repo.GetDashboardStats(ctx, userID)
}
