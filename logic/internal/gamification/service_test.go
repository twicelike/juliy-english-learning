package gamification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// MockRepository implements Repository for testing.
type MockRepository struct {
	Stats        *UserStats
	ProgressMap  map[string]*DailyProgress
	Achievements []UserAchievement
}

func (m *MockRepository) GetDailyProgress(ctx context.Context, userID bson.ObjectID, date string) (*DailyProgress, error) {
	dp, ok := m.ProgressMap[date]
	if ok {
		return dp, nil
	}
	return nil, nil
}

func (m *MockRepository) UpsertDailyProgress(ctx context.Context, dp *DailyProgress) error {
	m.ProgressMap[dp.Date] = dp
	return nil
}

func (m *MockRepository) GetProgressHistory(ctx context.Context, userID bson.ObjectID, startDate, endDate string) ([]DailyProgress, error) {
	var hist []DailyProgress
	for _, dp := range m.ProgressMap {
		if dp.Date >= startDate && dp.Date <= endDate {
			hist = append(hist, *dp)
		}
	}
	return hist, nil
}

func (m *MockRepository) GetUserStats(ctx context.Context, userID bson.ObjectID) (*UserStats, error) {
	return m.Stats, nil
}

func (m *MockRepository) UpsertUserStats(ctx context.Context, stats *UserStats) error {
	m.Stats = stats
	return nil
}

func (m *MockRepository) HasAchievement(ctx context.Context, userID bson.ObjectID, achievementID string) (bool, error) {
	for _, a := range m.Achievements {
		if a.AchievementID == achievementID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockRepository) UnlockAchievement(ctx context.Context, ua *UserAchievement) error {
	m.Achievements = append(m.Achievements, *ua)
	return nil
}

func (m *MockRepository) GetUserAchievements(ctx context.Context, userID bson.ObjectID) ([]UserAchievement, error) {
	return m.Achievements, nil
}

func (m *MockRepository) EnsureIndexes(ctx context.Context) error {
	return nil
}

func TestTrackActivity_XPAndStreak(t *testing.T) {
	repo := &MockRepository{
		ProgressMap: make(map[string]*DailyProgress),
	}
	svc := NewService(repo)
	ctx := context.Background()
	userID := bson.NewObjectID().Hex()
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 1. Initial Review (New day, 0 to 1 streak)
	err := svc.TrackActivity(ctx, userID, ActivityReview, true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.Stats.CurrentStreak != 1 {
		t.Errorf("expected streak to be 1, got %d", repo.Stats.CurrentStreak)
	}
	if repo.Stats.TotalXP != ActivityReview.XP() {
		t.Errorf("expected XP %d, got %d", ActivityReview.XP(), repo.Stats.TotalXP)
	}
	if repo.ProgressMap[today].CorrectReviews != 1 {
		t.Errorf("expected 1 correct review today")
	}

	// 2. Second review on the same day (Streak remains 1)
	err = svc.TrackActivity(ctx, userID, ActivityLearnNew, true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Stats.CurrentStreak != 1 {
		t.Errorf("expected streak to remain 1, got %d", repo.Stats.CurrentStreak)
	}
	if repo.Stats.TotalXP != ActivityReview.XP()+ActivityLearnNew.XP() {
		t.Errorf("expected XP %d, got %d", ActivityReview.XP()+ActivityLearnNew.XP(), repo.Stats.TotalXP)
	}
	if repo.ProgressMap[today].NewWordsLearned != 1 {
		t.Errorf("expected 1 new word learned today")
	}

	// 3. Simulate activity yesterday, to check if streak resets if we act today?
	// Wait, we already acted today. Let's manipulate LastActivityDate to two days ago to simulate a gap.
	twoDaysAgo := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	repo.Stats.LastActivityDate = twoDaysAgo
	repo.Stats.CurrentStreak = 5
	repo.Stats.LongestStreak = 5

	err = svc.TrackActivity(ctx, userID, ActivityQuiz, false, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Streak should reset to 1 because yesterday was missed.
	if repo.Stats.CurrentStreak != 1 {
		t.Errorf("expected streak to reset to 1, got %d", repo.Stats.CurrentStreak)
	}
	// Longest streak should remain 5
	if repo.Stats.LongestStreak != 5 {
		t.Errorf("expected longest streak to be 5, got %d", repo.Stats.LongestStreak)
	}

	// 4. Manipulate to yesterday to see if it increments to 2
	repo.Stats.LastActivityDate = yesterday
	repo.Stats.CurrentStreak = 1
	err = svc.TrackActivity(ctx, userID, ActivityReview, true, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Stats.CurrentStreak != 2 {
		t.Errorf("expected streak to increment to 2, got %d", repo.Stats.CurrentStreak)
	}
}

func TestCalculateLevel(t *testing.T) {
	// L1: 0-99
	l, c, n := CalculateLevel(50)
	if l != 1 || c != 50 || n != 100 {
		t.Errorf("Failed L1: got %d, %d, %d", l, c, n)
	}
	// L2: 100-299
	l, c, n = CalculateLevel(150)
	if l != 2 || c != 50 || n != 200 {
		t.Errorf("Failed L2: got %d, %d, %d", l, c, n)
	}
	// L3: 300-599
	l, c, n = CalculateLevel(300)
	if l != 3 || c != 0 || n != 300 {
		t.Errorf("Failed L3: got %d, %d, %d", l, c, n)
	}
	l, c, n = CalculateLevel(600)
	if l != 4 || c != 0 || n != 400 {
		t.Errorf("Failed L4: got %d, %d, %d", l, c, n)
	}
}

func TestTrackActivity_DailyGoalBonus(t *testing.T) {
	repo := &MockRepository{
		ProgressMap: make(map[string]*DailyProgress),
	}
	svc := NewService(repo)
	ctx := context.Background()
	userID := bson.NewObjectID().Hex()
	today := time.Now().Format("2006-01-02")

	// Set daily goal to 3
	dailyGoal := 3

	// 1st action -> no bonus
	_ = svc.TrackActivity(ctx, userID, ActivityReview, true, dailyGoal)
	if repo.Stats.TotalXP != ActivityReview.XP() {
		t.Errorf("expected %d, got %d", ActivityReview.XP(), repo.Stats.TotalXP)
	}

	// 2nd action -> no bonus
	_ = svc.TrackActivity(ctx, userID, ActivityReview, true, dailyGoal)

	// 3rd action -> meets goal -> gets bonus
	_ = svc.TrackActivity(ctx, userID, ActivityReview, true, dailyGoal)

	expectedXP := ActivityReview.XP()*3 + ActivityDaily.XP()
	if repo.Stats.TotalXP != expectedXP {
		t.Errorf("expected XP %d after goal met, got %d", expectedXP, repo.Stats.TotalXP)
	}
	if !repo.ProgressMap[today].DailyGoalReached {
		t.Errorf("expected DailyGoalReached to be true")
	}

	// 4th action -> goal already met -> no extra bonus
	_ = svc.TrackActivity(ctx, userID, ActivityReview, true, dailyGoal)
	expectedXP += ActivityReview.XP() // only normal review XP
	if repo.Stats.TotalXP != expectedXP {
		t.Errorf("expected XP %d after exceeding goal, got %d", expectedXP, repo.Stats.TotalXP)
	}
}

func TestTrackSpeakingSession(t *testing.T) {
	repo := &MockRepository{
		ProgressMap: make(map[string]*DailyProgress),
	}
	svc := NewService(repo)
	ctx := context.Background()
	userID := bson.NewObjectID().Hex()
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		score      int
		expectedXP int
	}{
		{score: 20, expectedXP: 4},
		{score: 50, expectedXP: 10},
		{score: 80, expectedXP: 16},
		{score: 100, expectedXP: 20},
		{score: 0, expectedXP: 1}, // Minimum XP is 1
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Score_%d", tt.score), func(t *testing.T) {
			repo.ProgressMap = make(map[string]*DailyProgress) // Reset state
			repo.Stats = nil

			err := svc.TrackSpeakingSession(ctx, userID, tt.score)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			dp, exists := repo.ProgressMap[today]
			if !exists {
				t.Fatal("DailyProgress not created")
			}
			if dp.SpeakingTasksCompleted != 1 {
				t.Errorf("Expected 1 speaking task, got %d", dp.SpeakingTasksCompleted)
			}
			if dp.XP != tt.expectedXP {
				t.Errorf("Expected %d XP in DP, got %d", tt.expectedXP, dp.XP)
			}

			if repo.Stats == nil {
				t.Fatal("UserStats not created")
			}
			if repo.Stats.TotalXP != tt.expectedXP {
				t.Errorf("Expected %d total XP, got %d", tt.expectedXP, repo.Stats.TotalXP)
			}
		})
	}
}
