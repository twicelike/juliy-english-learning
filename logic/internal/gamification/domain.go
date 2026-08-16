package gamification

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// -----------------------------------------------------------------------------
// Constants
// -----------------------------------------------------------------------------

type ActivityType string

const (
	ActivityReview         ActivityType = "review"          // +5 XP
	ActivityLearnNew       ActivityType = "learn_new"       // +10 XP
	ActivityQuiz           ActivityType = "quiz"            // +15 XP
	ActivityWritingMission ActivityType = "writing_mission" // Dynamic XP
	ActivityDaily          ActivityType = "daily_goal"      // +20 XP (bonus)
)

func (a ActivityType) XP() int {
	switch a {
	case ActivityReview:
		return 5
	case ActivityLearnNew:
		return 10
	case ActivityQuiz:
		return 15
	case ActivityDaily:
		return 20
	default:
		return 0
	}
}

// -----------------------------------------------------------------------------
// Domain Models (DB Structs)
// -----------------------------------------------------------------------------

// DailyProgress logs activity for a specific day.
type DailyProgress struct {
	ID               bson.ObjectID `bson:"_id,omitempty"`
	UserID           bson.ObjectID `bson:"user_id"`
	Date             string        `bson:"date"` // Format: YYYY-MM-DD
	XP               int           `bson:"xp"`
	ReviewsCompleted int           `bson:"reviews_completed"`
	CorrectReviews   int           `bson:"correct_reviews"`
	NewWordsLearned          int           `bson:"new_words_learned"`
	StudyMinutes             int           `bson:"study_minutes"`
	WritingMissionsCompleted int           `bson:"writing_missions_completed"`
	SpeakingTasksCompleted   int           `bson:"speaking_tasks_completed"`
	DailyGoalReached         bool          `bson:"daily_goal_reached"`
	CreatedAt        time.Time     `bson:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at"`
}

// UserStats stores aggregate gamification metrics.
type UserStats struct {
	ID               bson.ObjectID `bson:"_id,omitempty"`
	UserID           bson.ObjectID `bson:"user_id"`
	TotalXP          int           `bson:"total_xp"`
	Level            int           `bson:"level"`
	CurrentStreak    int           `bson:"current_streak"`
	LongestStreak    int           `bson:"longest_streak"`
	LastActivityDate string        `bson:"last_activity_date"` // Format: YYYY-MM-DD
	CreatedAt        time.Time     `bson:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at"`
}

// UserAchievement tracks unlocked badges.
type UserAchievement struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	UserID        bson.ObjectID `bson:"user_id"`
	AchievementID string        `bson:"achievement_id"`
	UnlockedAt    time.Time     `bson:"unlocked_at"`
}

// Hardcoded Achievement Definitions
type AchievementDef struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

var Achievements = []AchievementDef{
	{ID: "FIRST_REVIEW", Title: "First Step", Description: "Completed your first review.", Icon: "🌱"},
	{ID: "100_WORDS", Title: "Centurion", Description: "Learned 100 words.", Icon: "💯"},
	{ID: "7_DAY_STREAK", Title: "Consistent", Description: "Maintained a 7-day learning streak.", Icon: "🔥"},
	{ID: "FIRST_PERFECT_WRITING", Title: "Shakespeare", Description: "Scored a perfect 100 on a writing mission.", Icon: "✍️"},
	{ID: "WRITING_10_STREAK", Title: "Dedicated Writer", Description: "Completed 10 writing missions in total.", Icon: "📚"},
}

// -----------------------------------------------------------------------------
// DTOs
// -----------------------------------------------------------------------------

// ChartDataPoint represents one bar/point on the chart.
type ChartDataPoint struct {
	Date       string `json:"date"`
	XP         int    `json:"xp"`
	Activities int    `json:"activities"`
	Words      int    `json:"words"` // Kept for backward compatibility
	Active     bool   `json:"active"`
}



type DailyProgressDTO struct {
	XP               int `json:"xp"`
	ReviewsCompleted int `json:"reviews_completed"`
	NewWordsLearned  int `json:"new_words_learned"`
}

// -----------------------------------------------------------------------------
// Interfaces
// -----------------------------------------------------------------------------

type Repository interface {
	GetDailyProgress(ctx context.Context, userID bson.ObjectID, date string) (*DailyProgress, error)
	UpsertDailyProgress(ctx context.Context, dp *DailyProgress) error
	GetProgressHistory(ctx context.Context, userID bson.ObjectID, startDate, endDate string) ([]DailyProgress, error)

	GetUserStats(ctx context.Context, userID bson.ObjectID) (*UserStats, error)
	UpsertUserStats(ctx context.Context, stats *UserStats) error

	HasAchievement(ctx context.Context, userID bson.ObjectID, achievementID string) (bool, error)
	UnlockAchievement(ctx context.Context, ua *UserAchievement) error
	GetUserAchievements(ctx context.Context, userID bson.ObjectID) ([]UserAchievement, error)

	EnsureIndexes(ctx context.Context) error
}

type Service interface {
	// Facade method for tracking all user actions
	TrackActivity(ctx context.Context, userID string, activity ActivityType, isCorrect bool, dailyGoal int) error
	TrackWritingMission(ctx context.Context, userID string, score int) error
	TrackSpeakingSession(ctx context.Context, userID string, score int) error
	
	GetStatsAndProgress(ctx context.Context, userID string) (*UserStats, *DailyProgress, []AchievementDef, error)
	GetChartData(ctx context.Context, userID string, periodDays int) ([]ChartDataPoint, error)
}
