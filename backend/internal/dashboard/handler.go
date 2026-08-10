package dashboard

import (
	"net/http"

	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/profile"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	profileSvc profile.Service
	gamiSvc    gamification.Service
}

func NewHandler(profileSvc profile.Service, gamiSvc gamification.Service) *Handler {
	return &Handler{
		profileSvc: profileSvc,
		gamiSvc:    gamiSvc,
	}
}

// UnifiedDashboardResponse is the grand master response for the dashboard.
type UnifiedDashboardResponse struct {
	Profile        profile.UserProfileDTO         `json:"profile"`
	Recommendation profile.RecommendationResponse `json:"recommendation"`

	Level         int                           `json:"level"`
	CurrentXP     int                           `json:"current_xp"` // XP in current level
	NextLevelXP   int                           `json:"next_level_xp"`
	TotalXP       int                           `json:"total_xp"`
	CurrentStreak int                           `json:"current_streak"`
	LongestStreak int                           `json:"longest_streak"`
	Achievements  []gamification.AchievementDef `json:"achievements"`
	TodayProgress gamification.DailyProgressDTO `json:"today_progress"`
	ChartData     []gamification.ChartDataPoint `json:"chart_data"`
}

func (h *Handler) GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := c.Request.Context()

	profileDash, err := h.profileSvc.GetDashboard(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile dashboard", "details": err.Error()})
		return
	}

	stats, dp, achievements, err := h.gamiSvc.GetStatsAndProgress(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get gamification stats", "details": err.Error()})
		return
	}

	chartData, _ := h.gamiSvc.GetChartData(ctx, userID, 7)
	level, currentXP, nextLevelXP := gamification.CalculateLevel(stats.TotalXP)

	resp := &UnifiedDashboardResponse{
		Profile:        profileDash.Profile,
		Recommendation: profileDash.Recommendation,
		Level:          level,
		CurrentXP:      currentXP,
		NextLevelXP:    nextLevelXP,
		TotalXP:        stats.TotalXP,
		CurrentStreak:  stats.CurrentStreak,
		LongestStreak:  stats.LongestStreak,
		Achievements:   achievements,
		TodayProgress: gamification.DailyProgressDTO{
			XP:               dp.XP,
			ReviewsCompleted: dp.ReviewsCompleted,
			NewWordsLearned:  dp.NewWordsLearned,
		},
		ChartData: chartData,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
