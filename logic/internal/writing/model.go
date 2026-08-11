package writing

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Difficulty levels
const (
	DifficultyEasy   = "Easy"
	DifficultyMedium = "Medium"
	DifficultyHard   = "Hard"
)

type WritingMission struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Topic            string        `bson:"topic" json:"topic"`
	Scenario         string        `bson:"scenario" json:"scenario"`
	Difficulty       string        `bson:"difficulty" json:"difficulty"`
	Vietnamese       []string      `bson:"vietnamese" json:"vietnamese"`
	TargetVocabulary []string      `bson:"target_vocabulary" json:"target_vocabulary"`
	GrammarFocus     []string      `bson:"grammar_focus,omitempty" json:"grammar_focus,omitempty"`
	Hash             string        `bson:"hash" json:"hash"` // MD5 or SHA256 of normalized vietnamese content
	CreatedBy        string        `bson:"created_by" json:"created_by"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
}

type UserWritingProgress struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           bson.ObjectID `bson:"user_id" json:"user_id"`
	MissionID        bson.ObjectID `bson:"mission_id" json:"mission_id"`
	Attempts         int           `bson:"attempts" json:"attempts"`
	BestScore        float64       `bson:"best_score" json:"best_score"`
	LastScore        float64       `bson:"last_score" json:"last_score"`
	AverageScore     float64       `bson:"average_score" json:"average_score"`
	Completed        bool          `bson:"completed" json:"completed"`
	MasteryLevel     string        `bson:"mastery_level" json:"mastery_level"` // e.g., Beginner, Familiar, Mastered
	NextReviewAt     *time.Time    `bson:"next_review_at,omitempty" json:"next_review_at,omitempty"`
	FirstCompletedAt *time.Time    `bson:"first_completed_at,omitempty" json:"first_completed_at,omitempty"`
	LastCompletedAt  *time.Time    `bson:"last_completed_at,omitempty" json:"last_completed_at,omitempty"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"updated_at"`
}

type WritingSubmission struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           bson.ObjectID `bson:"user_id" json:"user_id"`
	MissionID        bson.ObjectID `bson:"mission_id" json:"mission_id"`
	UserContent      string        `bson:"user_content" json:"user_content"`
	CorrectedVersion string        `bson:"corrected_version" json:"corrected_version"`
	DetailedFeedback string        `bson:"detailed_feedback" json:"detailed_feedback"`
	Score            int           `bson:"score" json:"score"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
}

// Config lists for generation
var SupportedTopics = []string{
	"Daily Life", "Education", "Work", "Travel",
	"Technology", "Food", "Finance", "Health",
	"Entertainment", "Relationships",
}

var SupportedScenarios = map[string][]string{
	"Daily Life":    {"Morning Routine", "Shopping", "Cooking", "Weekend", "Household Chores"},
	"Education":     {"Homework", "Classroom", "University", "Presentation", "Exam Prep"},
	"Work":          {"Office", "Meeting", "Deadline", "Job Interview", "Business Trip"},
	"Travel":        {"Airport", "Hotel", "Taxi", "Restaurant", "Sightseeing", "Lost Luggage"},
	"Technology":    {"Smartphone", "Internet", "AI", "Cyber Security", "Social Media"},
	"Food":          {"Restaurant", "Grocery Shopping", "Cooking", "Cafe", "Recipe"},
	"Finance":       {"Saving Money", "Banking", "Credit Card", "Investment", "Budgeting"},
	"Health":        {"Exercise", "Hospital", "Pharmacy", "Healthy Lifestyle", "Diet"},
	"Entertainment": {"Movies", "Books", "Music", "Sports", "Gaming"},
	"Relationships": {"Friends", "Family", "Dating", "Wedding", "Colleagues"},
}

var SupportedDifficulties = []string{
	DifficultyEasy,
	DifficultyMedium,
	DifficultyHard,
}
