package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockService struct {
	Service
}

func (m *mockService) GetDashboard(ctx context.Context, userID string) (*DashboardResponse, error) {
	return &DashboardResponse{
		Profile: UserProfileDTO{
			CurrentCEFR: "A2",
			DailyGoal:   20,
		},
		Statistics: DashboardStatistics{
			WordsLearned: 50,
			ReviewsDue:   10,
			Accuracy:     85,
		},
		Charts: []interface{}{ // Mocking chart data points
			map[string]interface{}{
				"date":   "2026-08-01",
				"xp":     100,
				"words":  10,
				"active": true,
			},
		},
		Gamification: DashboardGamification{
			Level:         3,
			CurrentXP:     150,
			NextLevelXP:   300,
			Streak:        5,
			DailyGoalMet:  false,
			TodayProgress: 10,
			DailyGoal:     20,
			Achievements: []interface{}{
				map[string]interface{}{
					"id":    "FIRST_REVIEW",
					"title": "First Step",
				},
			},
		},
		Recommendation: RecommendationResponse{
			Action:      ActionReview,
			Title:       "Review Due Words",
			Description: "You have 10 words waiting.",
			Count:       10,
		},
	}, nil
}

func TestGetDashboard_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	handler := NewHandler(&mockService{}, nil)

	// Simulate auth middleware setting user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "some-object-id")
		c.Next()
	})
	router.GET("/dashboard", handler.GetDashboard)

	req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify top-level structure
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data field in response")
	}

	requiredFields := []string{"profile", "statistics", "charts", "gamification", "recommendation"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			t.Errorf("missing field in response data: %s", field)
		}
	}
}
