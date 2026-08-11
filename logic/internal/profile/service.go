package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/english/backend/internal/ai"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/learning"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo            Repository
	learningService learning.Service
	aiClient        ai.AIClient
	gamiService     gamification.Service
}

func NewService(repo Repository, learningSvc learning.Service, aiClient ai.AIClient, gamiSvc gamification.Service) Service {
	return &service{
		repo:            repo,
		learningService: learningSvc,
		aiClient:        aiClient,
		gamiService:     gamiSvc,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*UserProfileDTO, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, oid)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		// Return empty/default profile if not found
		return &UserProfileDTO{
			CurrentCEFR:        "Unknown",
			DailyGoal:          20,
			CompletedPlacement: false,
			Strengths:          []string{},
			Weaknesses:         []string{},
		}, nil
	}

	return &UserProfileDTO{
		CurrentCEFR:        profile.CurrentCEFR,
		DailyGoal:          profile.DailyGoal,
		CompletedPlacement: profile.CompletedPlacement,
		Strengths:          profile.Strengths,
		Weaknesses:         profile.Weaknesses,
	}, nil
}

func (s *service) GetPlacementTest(ctx context.Context) (*PlacementTestResponse, error) {
	// Convert internal questions to DTOs (strip CorrectAnswer)
	var dtos []PlacementQuestionDTO
	for _, q := range placementTestQuestions {
		dtos = append(dtos, PlacementQuestionDTO{
			ID:       q.ID,
			Type:     q.Type,
			Question: q.Question,
			Options:  q.Options,
		})
	}
	return &PlacementTestResponse{Questions: dtos}, nil
}

func (s *service) SubmitPlacementTest(ctx context.Context, userID string, req *SubmitTestRequest) (*SubmitTestResponse, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	score, cefr, strengths, weaknesses := evaluateTest(req.Answers)

	profile, err := s.repo.GetByUserID(ctx, oid)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		profile = &UserProfile{
			ID:     bson.NewObjectID(),
			UserID: oid,
		}
	}

	profile.PlacementScore = score
	profile.CurrentCEFR = cefr
	profile.Strengths = strengths
	profile.Weaknesses = weaknesses
	profile.CompletedPlacement = true
	profile.LastAssessmentAt = time.Now()
	profile.UpdatedAt = time.Now()
	if profile.DailyGoal == 0 {
		profile.DailyGoal = 20
	}
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	if err := s.repo.Upsert(ctx, profile); err != nil {
		return nil, err
	}

	return &SubmitTestResponse{
		Score: score,
		CEFR:  cefr,
	}, nil
}

type aiRecommendation struct {
	Message string `json:"message"`
}

func (s *service) GetDashboard(ctx context.Context, userID string) (*DashboardResponse, error) {
	// 1. Load Profile
	profileDTO, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Load Learning Progress
	learningDash, err := s.learningService.GetDashboard(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("learning service error: %w", err)
	}

	// 3. Load Gamification
	stats, dp, unlockedDefs, err := s.gamiService.GetStatsAndProgress(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("gamification service error: %w", err)
	}
	charts, err := s.gamiService.GetChartData(ctx, userID, 7)
	if err != nil {
		return nil, fmt.Errorf("chart data error: %w", err)
	}

	// 4. Logic to determine Next Action
	var action RecommendationAction
	var title, desc string
	var count int

	if learningDash.DueToday > 0 {
		action = ActionReview
		title = "Review Due Words"
		desc = fmt.Sprintf("You have %d words waiting to be reviewed.", learningDash.DueToday)
		count = learningDash.DueToday
	} else if learningDash.PendingCount > 0 {
		action = ActionLearnNew
		title = "Continue Learning"
		desc = fmt.Sprintf("You have %d new words you haven't finished learning.", learningDash.PendingCount)
		count = learningDash.PendingCount
	} else if learningDash.NewCount > 0 {
		action = ActionLearnNew
		title = "Learn New Words"
		desc = fmt.Sprintf("You have %d new bookmarked words.", learningDash.NewCount)
		count = learningDash.NewCount
	} else if len(profileDTO.Weaknesses) > 0 {
		action = ActionGrammar
		title = "Improve Weaknesses"
		desc = fmt.Sprintf("Practice your weak topics like %s.", profileDTO.Weaknesses[0])
		count = 0
	} else {
		action = ActionDone
		title = "All caught up!"
		desc = "You have met your daily goals. Keep it up!"
		count = 0
	}

	// 5. Optionally Generate AI Summary Message
	aiMessage := "Keep up the good work!"
	if s.aiClient != nil && profileDTO.CompletedPlacement {
		prompt := fmt.Sprintf(`You are a concise, encouraging English tutor.
The user is at CEFR level %s.
Due words today: %d.
New words to learn: %d.
Weaknesses: %v.
Strengths: %v.

Write a very short, encouraging advice sentence (under 30 words) for the user on what to focus on next.
Return ONLY a JSON object with a "message" string field.`, 
		profileDTO.CurrentCEFR, learningDash.DueToday, learningDash.NewCount, profileDTO.Weaknesses, profileDTO.Strengths)
		
		var aiResp aiRecommendation
		if err := s.aiClient.GenerateJSON(ctx, prompt, nil, &aiResp); err == nil && aiResp.Message != "" {
			aiMessage = aiResp.Message
		}
	}

	// Calculate accuracy (dummy value or based on DP)
	accuracy := 0
	if dp.ReviewsCompleted > 0 {
		accuracy = (dp.CorrectReviews * 100) / dp.ReviewsCompleted
	}

	totalActionsToday := dp.ReviewsCompleted + dp.NewWordsLearned

	_, currentXP, nextLevelXP := gamification.CalculateLevel(stats.TotalXP)

	resp := &DashboardResponse{
		Profile: *profileDTO,
		Statistics: DashboardStatistics{
			WordsLearned: learningDash.LearnedCount,
			ReviewsDue:   learningDash.DueToday,
			PendingWords: learningDash.PendingCount,
			NewWords:     learningDash.NewCount,
			Accuracy:     accuracy,
		},
		Charts: charts,
		Gamification: DashboardGamification{
			Level:         stats.Level,
			CurrentXP:     currentXP,
			NextLevelXP:   nextLevelXP,
			Streak:        stats.CurrentStreak,
			DailyGoalMet:  dp.DailyGoalReached,
			TodayProgress: totalActionsToday,
			DailyGoal:     profileDTO.DailyGoal,
			Achievements:  unlockedDefs,
		},
		Recommendation: RecommendationResponse{
			Action:      action,
			Title:       title,
			Description: desc,
			Count:       count,
			AIMessage:   aiMessage,
		},
	}

	return resp, nil
}
