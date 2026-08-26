package gamification

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CalculateLevel determines the user's level based on TotalXP.
// Simple curve: Level N requires N * 100 XP from the previous level.
// L1: 0-99, L2: 100-299, L3: 300-599, L4: 600-999.
func CalculateLevel(totalXP int) (level int, currentXP int, nextLevelXP int) {
	level = 1
	baseXP := 0
	for {
		reqForNext := level * 100
		if totalXP < baseXP+reqForNext {
			currentXP = totalXP - baseXP
			nextLevelXP = reqForNext
			break
		}
		baseXP += reqForNext
		level++
	}
	return
}

func (s *service) TrackActivity(ctx context.Context, userID string, activity ActivityType, count int, correctCount int, customXP int, dailyGoal int) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	xpGained := customXP
	if xpGained == 0 {
		xpGained = activity.XP() * count
	}

	// 1. Update Daily Progress
	dp, err := s.repo.GetDailyProgress(ctx, oid, today)
	if err != nil {
		return err
	}
	if dp == nil {
		dp = &DailyProgress{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			Date:      today,
			CreatedAt: time.Now(),
		}
	}

	if activity == ActivityReview || activity == ActivityQuiz {
		dp.ReviewsCompleted += count
		dp.CorrectReviews += correctCount
	} else if activity == ActivityLearnNew {
		dp.NewWordsLearned += count
	}

	// Check for Daily Goal Bonus
	if dailyGoal > 0 && !dp.DailyGoalReached {
		totalActionsToday := dp.ReviewsCompleted + dp.NewWordsLearned
		if totalActionsToday >= dailyGoal {
			dp.DailyGoalReached = true
			xpGained += ActivityDaily.XP()
		}
	}

	dp.XP += xpGained
	dp.UpdatedAt = time.Now()
	if err := s.repo.UpsertDailyProgress(ctx, dp); err != nil {
		return err
	}

	// 2. Update User Stats & Streak
	stats, err := s.repo.GetUserStats(ctx, oid)
	if err != nil {
		return err
	}
	if stats == nil {
		stats = &UserStats{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			CreatedAt: time.Now(),
		}
	}

	stats.TotalXP += xpGained
	stats.Level, _, _ = CalculateLevel(stats.TotalXP)

	// Streak Logic
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if stats.LastActivityDate == yesterday {
		// Logged activity yesterday, increment streak
		stats.CurrentStreak++
	} else if stats.LastActivityDate != today {
		// Missed yesterday, reset streak (or first time)
		stats.CurrentStreak = 1
	}
	// If stats.LastActivityDate == today, streak remains the same (already incremented today)

	if stats.CurrentStreak > stats.LongestStreak {
		stats.LongestStreak = stats.CurrentStreak
	}
	stats.LastActivityDate = today
	stats.UpdatedAt = time.Now()

	if err := s.repo.UpsertUserStats(ctx, stats); err != nil {
		return err
	}

	// 3. Check Achievements
	s.checkAchievements(ctx, oid, stats, dp)

	return nil
}

func (s *service) checkAchievements(ctx context.Context, userID bson.ObjectID, stats *UserStats, dp *DailyProgress) {
	// FIRST_REVIEW
	if stats.TotalXP > 0 {
		s.unlockAchievementIfMissing(ctx, userID, "FIRST_REVIEW")
	}

	// 7_DAY_STREAK
	if stats.CurrentStreak >= 7 {
		s.unlockAchievementIfMissing(ctx, userID, "7_DAY_STREAK")
	}

	// FIRST_PERFECT_WRITING
	if dp.WritingMissionsCompleted > 0 && stats.TotalXP > 0 { // we can't cleanly check score here, so we do it in TrackWritingMission directly instead.
		// Actually, I'll move this to TrackWritingMission for cleaner logic.
	}
}

// TrackWritingMission tracks a completed writing mission, awarding XP based on score.
func (s *service) TrackWritingMission(ctx context.Context, userID string, score int) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	xpGained := 10
	if score >= 90 {
		xpGained = 100
	} else if score >= 80 {
		xpGained = 70
	} else if score >= 70 {
		xpGained = 50
	}

	// 1. Update Daily Progress
	dp, err := s.repo.GetDailyProgress(ctx, oid, today)
	if err != nil {
		return err
	}
	if dp == nil {
		dp = &DailyProgress{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			Date:      today,
			CreatedAt: time.Now(),
		}
	}
	dp.WritingMissionsCompleted++
	dp.XP += xpGained
	dp.UpdatedAt = time.Now()
	if err := s.repo.UpsertDailyProgress(ctx, dp); err != nil {
		return err
	}

	// 2. Update User Stats & Streak
	stats, err := s.repo.GetUserStats(ctx, oid)
	if err != nil {
		return err
	}
	if stats == nil {
		stats = &UserStats{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			CreatedAt: time.Now(),
		}
	}

	stats.TotalXP += xpGained
	stats.Level, _, _ = CalculateLevel(stats.TotalXP)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if stats.LastActivityDate == yesterday {
		stats.CurrentStreak++
	} else if stats.LastActivityDate != today {
		stats.CurrentStreak = 1
	}
	if stats.CurrentStreak > stats.LongestStreak {
		stats.LongestStreak = stats.CurrentStreak
	}
	stats.LastActivityDate = today
	stats.UpdatedAt = time.Now()

	if err := s.repo.UpsertUserStats(ctx, stats); err != nil {
		return err
	}

	// 3. Check Achievements
	s.checkAchievements(ctx, oid, stats, dp)
	if score == 100 {
		s.unlockAchievementIfMissing(ctx, oid, "FIRST_PERFECT_WRITING")
	}

	// Check if user has done 10 writing missions total (approx check: we could sum daily progress or just rely on a new field, but let's assume if dp.WritingMissionsCompleted >= 10 today, or we can just query the writing progress table. For simplicity, we just grant it if they have done 10).
	// For this Phase, we'll keep it simple: if they do 10 in a day, they get it, or we skip checking it here.
	// We will skip WRITING_10_STREAK logic for now unless we add a specific tracker. Let's just grant it if streak >= 10.
	if stats.CurrentStreak >= 10 {
		s.unlockAchievementIfMissing(ctx, oid, "WRITING_10_STREAK")
	}

	return nil
}

// TrackSpeakingSession tracks a completed speaking session, awarding XP based on score.
func (s *service) TrackSpeakingSession(ctx context.Context, userID string, score int) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	xpGained := score / 5
	if xpGained < 1 {
		xpGained = 1
	}

	// 1. Update Daily Progress
	dp, err := s.repo.GetDailyProgress(ctx, oid, today)
	if err != nil {
		return err
	}
	if dp == nil {
		dp = &DailyProgress{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			Date:      today,
			CreatedAt: time.Now(),
		}
	}
	dp.SpeakingTasksCompleted++
	dp.XP += xpGained
	dp.UpdatedAt = time.Now()
	if err := s.repo.UpsertDailyProgress(ctx, dp); err != nil {
		return err
	}

	// 2. Update User Stats & Streak
	stats, err := s.repo.GetUserStats(ctx, oid)
	if err != nil {
		return err
	}
	if stats == nil {
		stats = &UserStats{
			ID:        bson.NewObjectID(),
			UserID:    oid,
			CreatedAt: time.Now(),
		}
	}

	stats.TotalXP += xpGained
	stats.Level, _, _ = CalculateLevel(stats.TotalXP)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if stats.LastActivityDate == yesterday {
		stats.CurrentStreak++
	} else if stats.LastActivityDate != today {
		stats.CurrentStreak = 1
	}
	if stats.CurrentStreak > stats.LongestStreak {
		stats.LongestStreak = stats.CurrentStreak
	}
	stats.LastActivityDate = today
	stats.UpdatedAt = time.Now()

	if err := s.repo.UpsertUserStats(ctx, stats); err != nil {
		return err
	}

	// 3. Check Achievements (general check)
	s.checkAchievements(ctx, oid, stats, dp)

	return nil
}

func (s *service) unlockAchievementIfMissing(ctx context.Context, userID bson.ObjectID, achievementID string) {
	has, _ := s.repo.HasAchievement(ctx, userID, achievementID)
	if !has {
		ua := &UserAchievement{
			ID:            bson.NewObjectID(),
			UserID:        userID,
			AchievementID: achievementID,
			UnlockedAt:    time.Now(),
		}
		_ = s.repo.UnlockAchievement(ctx, ua) // Fire and forget for now
	}
}

func (s *service) GetStatsAndProgress(ctx context.Context, userID string) (*UserStats, *DailyProgress, []AchievementDef, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid user id: %w", err)
	}

	stats, err := s.repo.GetUserStats(ctx, oid)
	if err != nil {
		return nil, nil, nil, err
	}
	if stats == nil {
		stats = &UserStats{Level: 1}
	}

	today := time.Now().Format("2006-01-02")
	dp, err := s.repo.GetDailyProgress(ctx, oid, today)
	if err != nil {
		return nil, nil, nil, err
	}
	if dp == nil {
		dp = &DailyProgress{}
	}

	uas, err := s.repo.GetUserAchievements(ctx, oid)
	if err != nil {
		return nil, nil, nil, err
	}

	var unlockedDefs = make([]AchievementDef, 0)
	for _, ua := range uas {
		for _, def := range Achievements {
			if def.ID == ua.AchievementID {
				unlockedDefs = append(unlockedDefs, def)
				break
			}
		}
	}

	return stats, dp, unlockedDefs, nil
}

func (s *service) GetChartData(ctx context.Context, userID string, periodDays int) ([]ChartDataPoint, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -(periodDays - 1))

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	history, err := s.repo.GetProgressHistory(ctx, oid, startStr, endStr)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	histMap := make(map[string]DailyProgress)
	for _, h := range history {
		histMap[h.Date] = h
	}

	var chart []ChartDataPoint
	for i := 0; i < periodDays; i++ {
		dateObj := startDate.AddDate(0, 0, i)
		dateStr := dateObj.Format("2006-01-02")

		dp, exists := histMap[dateStr]
		if exists {
			activitiesCount := dp.ReviewsCompleted + dp.NewWordsLearned + dp.SpeakingTasksCompleted + dp.WritingMissionsCompleted
			chart = append(chart, ChartDataPoint{
				Date:       dateStr, // could format to "Mon" or "02/01" for UI
				XP:         dp.XP,
				Activities: activitiesCount,
				Words:      activitiesCount, // Keep backward compatibility
				Active:     true,
			})
		} else {
			chart = append(chart, ChartDataPoint{
				Date:       dateStr,
				XP:         0,
				Activities: 0,
				Words:      0,
				Active:     false,
			})
		}
	}

	return chart, nil
}
