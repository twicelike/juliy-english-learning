package learning_test

import (
	"testing"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/learning"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestDiscoverNextSense(t *testing.T) {
	wid := bson.NewObjectID()

	// Helper to create a sense
	createSense := func(id string, eligible bool) dictionary.SenseDoc {
		quality := &dictionary.SenseQuality{
			QualityScore: 100,
			Category:     dictionary.CategoryPrimary,
		}
		if !eligible {
			quality.Category = "" // Will make it ineligible
		}
		return dictionary.SenseDoc{
			ID:      id,
			Quality: quality,
		}
	}

	// Helper to create progress
	createProgress := func(state string) *learning.LearningProgress {
		switch state {
		case "none":
			return nil
		case "learning":
			return &learning.LearningProgress{Interval: 5} // Not mastered
		case "mastered":
			return &learning.LearningProgress{Interval: 21} // Mastered
		default:
			return nil
		}
	}

	tests := []struct {
		name     string
		s1State  string // "none", "learning", "mastered", or "ineligible" (special flag for sense 1)
		s2State  string
		s3State  string
		expected string // Expected sense ID to be unlocked
	}{
		{"All none -> S1", "none", "none", "none", "s1"},
		{"S1 learning -> S1", "learning", "none", "none", "s1"},
		{"S1 mastered -> S2", "mastered", "none", "none", "s2"},
		{"S1 mastered, S2 learning -> S2", "mastered", "learning", "none", "s2"},
		{"S1 mastered, S2 mastered -> S3", "mastered", "mastered", "none", "s3"},
		{"All mastered -> none", "mastered", "mastered", "mastered", ""},
		{"S1 mastered, S2 ineligible, S3 none -> S3", "mastered", "ineligible", "none", "s3"},
		{"S1 ineligible, S2 none, S3 none -> S2", "ineligible", "none", "none", "s2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &dictionary.WordDocument{
				ID:   wid,
				Word: "test",
				Senses: []dictionary.SenseDoc{
					createSense("s1", tt.s1State != "ineligible"),
					createSense("s2", tt.s2State != "ineligible"),
					createSense("s3", tt.s3State != "ineligible"),
				},
			}

			progresses := map[string]*learning.LearningProgress{
				"s1": createProgress(tt.s1State),
				"s2": createProgress(tt.s2State),
				"s3": createProgress(tt.s3State),
			}

			res := learning.DiscoverNextSense(doc, progresses)

			if tt.expected == "" {
				assert.Nil(t, res)
			} else {
				if assert.NotNil(t, res) {
					assert.Equal(t, tt.expected, res.ID)
				}
			}
		})
	}
}
