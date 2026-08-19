package evaluation

import "errors"

// ScoringPolicy defines the weights and thresholds used by the Scoring Engine.
type ScoringPolicy struct {
	RecentDays          int
	RetentionDays       int
	RecentWeight        float64
	RetentionWeight     float64
	HistoricalWeight    float64
	TargetEvidenceCount int
	TrendThreshold      float64
}

// DefaultScoringPolicy is the recommended policy for MVP.
var DefaultScoringPolicy = ScoringPolicy{
	RecentDays:          7,
	RetentionDays:       30,
	RecentWeight:        0.60,
	RetentionWeight:     0.25,
	HistoricalWeight:    0.15,
	TargetEvidenceCount: 100,
	TrendThreshold:      5.0,
}

// Validate checks if the policy configuration is mathematically sound.
func (p ScoringPolicy) Validate() error {
	if p.RecentDays <= 0 {
		return errors.New("RecentDays must be > 0")
	}
	if p.RetentionDays <= p.RecentDays {
		return errors.New("RetentionDays must be > RecentDays")
	}
	if p.RecentWeight < 0 || p.RetentionWeight < 0 || p.HistoricalWeight < 0 {
		return errors.New("weights cannot be negative")
	}
	if (p.RecentWeight + p.RetentionWeight + p.HistoricalWeight) <= 0 {
		return errors.New("sum of weights must be > 0")
	}
	if p.TargetEvidenceCount <= 0 {
		return errors.New("TargetEvidenceCount must be > 0")
	}
	if p.TrendThreshold < 0 {
		return errors.New("TrendThreshold cannot be negative")
	}
	return nil
}
