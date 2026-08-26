package profile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/english/backend/internal/ai"
	"github.com/english/backend/internal/evaluation"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/learning"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo            Repository
	learningService learning.Service
	aiClient        ai.AIClient
	gamiService     gamification.Service
	evalService     evaluation.Service
}

func NewService(repo Repository, learningSvc learning.Service, aiClient ai.AIClient, gamiSvc gamification.Service, evalSvc evaluation.Service) Service {
	return &service{
		repo:            repo,
		learningService: learningSvc,
		aiClient:        aiClient,
		gamiService:     gamiSvc,
		evalService:     evalSvc,
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
			Skills:             []SkillAssessment{},
			Goals:              []string{},
			FocusAreas:         []string{},
		}, nil
	}

	dto := &UserProfileDTO{
		CurrentCEFR:        profile.CurrentCEFR,
		DailyGoal:          profile.DailyGoal,
		CompletedPlacement: profile.CompletedPlacement,
		Strengths:          profile.Strengths,
		Weaknesses:         profile.Weaknesses,
		Skills:             profile.Skills,
		Goals:              profile.Goals,
		FocusAreas:         profile.FocusAreas,
	}
	if !profile.LastAssessmentAt.IsZero() {
		dto.LastAssessmentAt = &profile.LastAssessmentAt
	}

	if dto.Strengths == nil {
		dto.Strengths = []string{}
	}
	if dto.Weaknesses == nil {
		dto.Weaknesses = []string{}
	}
	if dto.Skills == nil {
		dto.Skills = []SkillAssessment{}
	}
	if dto.Goals == nil {
		dto.Goals = []string{}
	}
	if dto.FocusAreas == nil {
		dto.FocusAreas = []string{}
	}

	return dto, nil
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

	score, cefr, skillScores, skillMax := evaluateTest(req.Answers)

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

	// Create and record evidence for DE5
	if s.evalService != nil {
		var evidences []*evaluation.Evidence
		eventID := fmt.Sprintf("assessment_%d", time.Now().Unix())

		for skillStr, max := range skillMax {
			correct := skillScores[skillStr]
			// Normalized score (0-100)
			skillScore := 0.0
			if max > 0 {
				skillScore = (float64(correct) / float64(max)) * 100
			}

			evidences = append(evidences, &evaluation.Evidence{
				UserID:     oid,
				EventID:    eventID,
				Skill:      skillStr,
				Score:      skillScore,
				Weight:     evaluation.WeightReview, // standard weight for tests
				Source:     "assessment_test",
				Metadata:   bson.M{"correct": correct, "max": max},
				OccurredAt: time.Now(),
			})
		}

		if len(evidences) > 0 {
			go func(evs []*evaluation.Evidence) {
				bgCtx := context.Background()
				_ = s.evalService.RecordEvidences(bgCtx, evs)
			}(evidences)
		}
	}

	// Fire async profile refresh event
	// To do this, we need events.Publisher injected into profile.Handler, but profile.Service doesn't have it.
	// Wait, we can inject EventPublisher into ProfileService too. Or we can just let the handler publish it?
	// For SubmitPlacementTest, it's better to let the handler publish it because Handler has Publisher.

	return &SubmitTestResponse{
		Score: score,
		CEFR:  cefr,
	}, nil
}

func (s *service) UpdatePreferences(ctx context.Context, userID string, req *UpdatePreferencesRequest) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, oid)
	if err != nil {
		return err
	}

	if profile == nil {
		profile = &UserProfile{
			ID:     bson.NewObjectID(),
			UserID: oid,
		}
	}

	profile.Goals = req.Goals
	profile.FocusAreas = req.FocusAreas
	profile.UpdatedAt = time.Now()

	return s.repo.Upsert(ctx, profile)
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
	} else if len(profileDTO.FocusAreas) > 0 {
		// Priority 2: User Intent (Focus Areas)
		// Pick the first focus area for simplicity
		focus := strings.ToLower(profileDTO.FocusAreas[0])
		if strings.Contains(focus, "writing") {
			action = ActionWriting
			title = "Focus on Writing"
			desc = "Practice writing to achieve your goals."
			count = 0
		} else if strings.Contains(focus, "grammar") {
			action = ActionGrammar
			title = "Focus on Grammar"
			desc = "Master grammar rules for better fluency."
			count = 0
		} else {
			// default for vocabulary or general
			action = ActionLearnNew
			title = "Focus on Vocabulary"
			desc = "Expand your vocabulary based on your goals."
			count = learningDash.PendingCount + learningDash.NewCount
		}
	} else {
		// Priority 3: Evidence-based assessment (Weak Skills)
		var weakestSkill *SkillAssessment
		for i, skill := range profileDTO.Skills {
			if skill.Level == SkillLevelWeak {
				if weakestSkill == nil || skill.Score < weakestSkill.Score {
					weakestSkill = &profileDTO.Skills[i]
				}
			}
		}

		if weakestSkill != nil {
			switch weakestSkill.Skill {
			case SkillWriting:
				action = ActionWriting
				title = "Improve Writing"
				desc = "Your writing skill needs some practice."
				count = 0
			case SkillGrammar:
				action = ActionGrammar
				title = "Improve Grammar"
				desc = "Let's review some grammar rules."
				count = 0
			case SkillVocabulary:
				action = ActionLearnNew
				title = "Improve Vocabulary"
				desc = "Learn more words to strengthen your vocabulary."
				count = learningDash.PendingCount + learningDash.NewCount
			default:
				// Fallback for listening/reading if no specific action
				action = ActionLearnNew
				title = "Keep Learning"
				desc = "Practice more to improve your weak skills."
				count = learningDash.PendingCount + learningDash.NewCount
			}
		} else if learningDash.PendingCount > 0 {
			// Priority 4: Learning fallback
			action = ActionLearnNew
			title = "Continue Learning"
			desc = fmt.Sprintf("You have %d new words you haven't finished learning.", learningDash.PendingCount)
			count = learningDash.PendingCount
		} else if learningDash.NewCount > 0 {
			action = ActionLearnNew
			title = "Learn New Words"
			desc = fmt.Sprintf("You have %d new bookmarked words.", learningDash.NewCount)
			count = learningDash.NewCount
		} else {
			// Priority 5: Done
			action = ActionDone
			title = "All caught up!"
			desc = "You have met your daily goals. Keep it up!"
			count = 0
		}
	}

	// 5. Optionally Generate AI Summary Message
	// Removed AI quote generation to drastically improve page load times.
	aiMessage := ""

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

func (s *service) RefreshSkills(ctx context.Context, userIDStr string) error {
	oid, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	profile, err := s.repo.GetByUserID(ctx, oid)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("profile not found")
	}

	skillsToEvaluate := []Skill{SkillVocabulary, SkillGrammar, SkillReading, SkillWriting, SkillListening, SkillSpeaking}
	now := time.Now()

	// Convert existing skills to a map for easy updates
	skillMap := make(map[Skill]SkillAssessment)
	for _, sk := range profile.Skills {
		skillMap[sk.Skill] = sk
	}

	changed := false

	for _, sk := range skillsToEvaluate {
		res, err := s.evalService.CalculateSkillAssessment(ctx, userIDStr, string(sk))
		if err != nil {
			// Log error but continue with other skills
			fmt.Printf("failed to calculate skill %s: %v\n", sk, err)
			continue
		}

		// Update if we have sufficient data (or if trend is insufficient but we want to initialize it)
		if res.Trend != evaluation.TrendInsufficientData || res.Confidence > 0 {
			skillMap[sk] = SkillAssessment{
				Skill:           sk,
				Score:           res.Score,
				Level:           SkillLevel(res.Level),
				Confidence:      res.Confidence,
				EvidenceCount:   res.EvidenceCount,
				LastEvaluatedAt: now,
				Trend:           SkillTrend(res.Trend),
			}
			changed = true
		} else {
			// If insufficient data, we only insert if it doesn't exist, to not overwrite a valid old assessment
			// As user requested: "Nếu hôm nay chỉ có vocabulary evidence, việc refresh không được làm mất assessment writing cũ."
			if _, exists := skillMap[sk]; !exists {
				skillMap[sk] = SkillAssessment{
					Skill:           sk,
					Score:           0,
					Level:           SkillLevelWeak,
					Confidence:      0,
					EvidenceCount:   0,
					LastEvaluatedAt: now,
					Trend:           SkillTrendInsufficient,
				}
				changed = true
			}
		}
	}

	if changed {
		var newSkills []SkillAssessment
		for _, sk := range skillsToEvaluate {
			if val, ok := skillMap[sk]; ok {
				newSkills = append(newSkills, val)
			}
		}
		profile.Skills = newSkills
		profile.UpdatedAt = now

		if err := s.repo.Upsert(ctx, profile); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}
	}

	return nil
}
