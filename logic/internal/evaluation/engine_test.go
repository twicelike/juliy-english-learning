package evaluation

import (
	"math"
	"testing"
	"time"
)

func TestEvaluateSkill(t *testing.T) {
	now := time.Now()
	policy := DefaultScoringPolicy

	tests := []struct {
		name          string
		evidences     []*Evidence
		expectedScore float64
		expectedTrend SkillTrend
		expectedLevel SkillLevel
		expectedConf  float64
	}{
		{
			name:          "No evidence",
			evidences:     []*Evidence{},
			expectedScore: 0,
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelWeak,
			expectedConf:  0,
		},
		{
			name: "Only Recent",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 80, Weight: 1.0},
				{OccurredAt: now.Add(-6 * 24 * time.Hour), Score: 100, Weight: 1.0},
			},
			expectedScore: 90.0, // (80+100)/2
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelStrong,
			expectedConf:  2.0, // 2 / 100 * 100
		},
		{
			name: "Recent + Retention",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 90, Weight: 1.0},  // Recent: 90
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 60, Weight: 1.0}, // Retention: 60
			},
			// Normalization: RecentWeight(0.60) / 0.85 = 0.70588
			// RetentionWeight(0.25) / 0.85 = 0.29411
			// Score: 90 * 0.70588 + 60 * 0.29411 = 63.5294 + 17.647 = 81.176
			expectedScore: 90*(0.60/0.85) + 60*(0.25/0.85),
			expectedTrend: TrendImproving, // 90 > 60 + 5
			expectedLevel: SkillLevelStrong,
			expectedConf:  2.0,
		},
		{
			name: "All Three Buckets",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 80, Weight: 1.0},  // Recent: 80
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 80, Weight: 1.0}, // Retention: 80
				{OccurredAt: now.Add(-40 * 24 * time.Hour), Score: 80, Weight: 1.0}, // Hist: 80
			},
			expectedScore: 80.0,
			expectedTrend: TrendStable,
			expectedLevel: SkillLevelStrong,
			expectedConf:  3.0,
		},
		{
			name: "Retention + Historical",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 70, Weight: 1.0}, // Retention: 70
				{OccurredAt: now.Add(-40 * 24 * time.Hour), Score: 50, Weight: 1.0}, // Hist: 50
			},
			// Normalization: 0.25/0.40 and 0.15/0.40
			// Score: 70*(0.625) + 50*(0.375) = 43.75 + 18.75 = 62.5
			expectedScore: 62.5,
			expectedTrend: TrendInsufficientData, // No recent
			expectedLevel: SkillLevelAverage,
			expectedConf:  2.0,
		},
		{
			name: "Trend Declining",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 50, Weight: 1.0},  // Recent: 50
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 60, Weight: 1.0}, // Retention: 60
			},
			expectedScore: 50*(0.60/0.85) + 60*(0.25/0.85),
			expectedTrend: TrendDeclining, // 50 < 60 - 5
			expectedLevel: SkillLevelWeak, // 35.29 + 17.64 = 52.94
			expectedConf:  2.0,
		},
		{
			name: "Exactly Boundary 7 and 30",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-7 * 24 * time.Hour), Score: 100, Weight: 1.0}, // Included in Recent
				{OccurredAt: now.Add(-30 * 24 * time.Hour), Score: 50, Weight: 1.0}, // Included in Retention
			},
			expectedScore: 100*(0.60/0.85) + 50*(0.25/0.85),
			expectedTrend: TrendImproving, // 100 > 50 + 5
			expectedLevel: SkillLevelStrong,
			expectedConf:  2.0,
		},
		{
			name: "Score 59.99 (Weak)",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 59.99, Weight: 1.0},
			},
			expectedScore: 59.99,
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelWeak,
			expectedConf:  1.0,
		},
		{
			name: "Score 60 (Average)",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 60, Weight: 1.0},
			},
			expectedScore: 60,
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelAverage,
			expectedConf:  1.0,
		},
		{
			name: "Score 79.99 (Average)",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 79.99, Weight: 1.0},
			},
			expectedScore: 79.99,
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelAverage,
			expectedConf:  1.0,
		},
		{
			name: "Score 80 (Strong)",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 80, Weight: 1.0},
			},
			expectedScore: 80,
			expectedTrend: TrendInsufficientData,
			expectedLevel: SkillLevelStrong,
			expectedConf:  1.0,
		},
		{
			name: "Trend Threshold +5",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 65, Weight: 1.0},
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 60, Weight: 1.0},
			},
			expectedScore: 65*(0.60/0.85) + 60*(0.25/0.85),
			expectedTrend: TrendStable, // 65 == 60 + 5
			expectedLevel: SkillLevelAverage,
			expectedConf:  2.0,
		},
		{
			name: "Trend Threshold -5",
			evidences: []*Evidence{
				{OccurredAt: now.Add(-1 * 24 * time.Hour), Score: 55, Weight: 1.0},
				{OccurredAt: now.Add(-10 * 24 * time.Hour), Score: 60, Weight: 1.0},
			},
			expectedScore: 55*(0.60/0.85) + 60*(0.25/0.85),
			expectedTrend: TrendStable,    // 55 == 60 - 5
			expectedLevel: SkillLevelWeak, // 38.82 + 17.64 = 56.47
			expectedConf:  2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := EvaluateSkill(tt.evidences, policy, now)
			if err != nil {
				t.Fatalf("EvaluateSkill returned error: %v", err)
			}

			if math.Abs(res.Score-tt.expectedScore) > 0.001 {
				t.Errorf("expected score %f, got %f", tt.expectedScore, res.Score)
			}
			if res.Trend != tt.expectedTrend {
				t.Errorf("expected trend %s, got %s", tt.expectedTrend, res.Trend)
			}
			if res.Level != tt.expectedLevel {
				t.Errorf("expected level %s, got %s", tt.expectedLevel, res.Level)
			}
			if math.Abs(res.Confidence-tt.expectedConf) > 0.001 {
				t.Errorf("expected confidence %f, got %f", tt.expectedConf, res.Confidence)
			}
		})
	}

	// Confidence max out test
	t.Run("Confidence max 100", func(t *testing.T) {
		evs := make([]*Evidence, 150)
		for i := 0; i < 150; i++ {
			evs[i] = &Evidence{OccurredAt: now, Score: 100, Weight: 1.0}
		}
		res, _ := EvaluateSkill(evs, policy, now)
		if res.Confidence != 100.0 {
			t.Errorf("expected confidence 100, got %f", res.Confidence)
		}
	})
}
