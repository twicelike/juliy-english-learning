package evaluation

import (
	"time"
)

// SkillTrend represents the trend of a skill.
type SkillTrend string

const (
	TrendImproving        SkillTrend = "improving"
	TrendStable           SkillTrend = "stable"
	TrendDeclining        SkillTrend = "declining"
	TrendInsufficientData SkillTrend = "insufficient_data"
)

// SkillLevel represents the classified level of a skill.
type SkillLevel string

const (
	SkillLevelWeak    SkillLevel = "weak"
	SkillLevelAverage SkillLevel = "average"
	SkillLevelStrong  SkillLevel = "strong"
)

// SkillAssessmentResult is the output of the Scoring Engine.
type SkillAssessmentResult struct {
	Score         float64
	Level         SkillLevel
	Confidence    float64
	EvidenceCount int
	Trend         SkillTrend
}

// EvaluateSkill calculates the skill assessment score from raw evidences based on a policy.
// It is a pure computation function.
func EvaluateSkill(evidences []*Evidence, policy ScoringPolicy, now time.Time) (SkillAssessmentResult, error) {
	if err := policy.Validate(); err != nil {
		return SkillAssessmentResult{}, err
	}

	evidenceCount := len(evidences)
	if evidenceCount == 0 {
		return SkillAssessmentResult{
			Score:         0,
			Level:         SkillLevelWeak,
			Confidence:    0,
			EvidenceCount: 0,
			Trend:         TrendInsufficientData,
		}, nil
	}

	// Time boundaries
	// We use exact time comparison based on OccurredAt and now
	// Recent: (now - RecentDays) <= OccurredAt <= now
	// Retention: (now - RetentionDays) <= OccurredAt < (now - RecentDays)
	// Historical: OccurredAt < (now - RetentionDays)
	recentCutoff := now.AddDate(0, 0, -policy.RecentDays)
	retentionCutoff := now.AddDate(0, 0, -policy.RetentionDays)

	// Wait, let's redefine bucket variables to float64 to support floating point Evidence Weights
	var rSum, retSum, histSum float64
	var rWeight, retWeight, histWeight float64

	for _, ev := range evidences {
		val := ev.Score * ev.Weight
		w := ev.Weight

		if !ev.OccurredAt.Before(recentCutoff) {
			rSum += val
			rWeight += w
		} else if !ev.OccurredAt.Before(retentionCutoff) {
			retSum += val
			retWeight += w
		} else {
			histSum += val
			histWeight += w
		}
	}

	var rAvg, retAvg, histAvg float64
	hasRecent := rWeight > 0
	hasRetention := retWeight > 0
	hasHistorical := histWeight > 0

	if hasRecent {
		rAvg = rSum / rWeight
	}
	if hasRetention {
		retAvg = retSum / retWeight
	}
	if hasHistorical {
		histAvg = histSum / histWeight
	}

	// Normalize bucket weights
	var currentRecentWeight, currentRetentionWeight, currentHistoricalWeight float64
	var totalValidWeight float64

	if hasRecent {
		totalValidWeight += policy.RecentWeight
	}
	if hasRetention {
		totalValidWeight += policy.RetentionWeight
	}
	if hasHistorical {
		totalValidWeight += policy.HistoricalWeight
	}

	if totalValidWeight > 0 {
		if hasRecent {
			currentRecentWeight = policy.RecentWeight / totalValidWeight
		}
		if hasRetention {
			currentRetentionWeight = policy.RetentionWeight / totalValidWeight
		}
		if hasHistorical {
			currentHistoricalWeight = policy.HistoricalWeight / totalValidWeight
		}
	}

	// Calculate Score
	score := (rAvg * currentRecentWeight) + (retAvg * currentRetentionWeight) + (histAvg * currentHistoricalWeight)

	// Calculate Trend
	trend := TrendInsufficientData
	if hasRecent {
		var baseline float64
		hasBaseline := false

		if hasRetention {
			baseline = retAvg
			hasBaseline = true
		} else if hasHistorical {
			baseline = histAvg
			hasBaseline = true
		}

		if hasBaseline {
			if rAvg > baseline+policy.TrendThreshold {
				trend = TrendImproving
			} else if rAvg < baseline-policy.TrendThreshold {
				trend = TrendDeclining
			} else {
				trend = TrendStable
			}
		}
	}

	// Calculate Confidence
	confidence := (float64(evidenceCount) / float64(policy.TargetEvidenceCount)) * 100.0
	if confidence > 100.0 {
		confidence = 100.0
	}

	return SkillAssessmentResult{
		Score:         score,
		Level:         SkillLevelFromScore(score),
		Confidence:    confidence,
		EvidenceCount: evidenceCount,
		Trend:         trend,
	}, nil
}

// SkillLevelFromScore maps a score to a SkillLevel.
func SkillLevelFromScore(score float64) SkillLevel {
	if score < 60 {
		return SkillLevelWeak
	}
	if score < 80 {
		return SkillLevelAverage
	}
	return SkillLevelStrong
}
