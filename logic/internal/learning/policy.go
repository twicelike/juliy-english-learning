package learning

import "github.com/english/backend/internal/dictionary"

// EligibilityPolicy defines the configurable rules for whether a sense is
// allowed to enter the learning flow. This isolates decision logic from
// dictionary metadata.
type EligibilityPolicy struct {
	AllowUnknownQuality bool
	MinQualityScore     int
}

// DefaultEligibilityPolicy is used during Phase 2 where all senses are eligible
// by default since the dictionary database has not been fully enriched yet.
var DefaultEligibilityPolicy = EligibilityPolicy{
	AllowUnknownQuality: true,
	MinQualityScore:     0,
}

// IsEligible returns true if the given sense quality meets the policy criteria.
func (p EligibilityPolicy) IsEligible(quality *dictionary.SenseQuality) bool {
	if quality == nil {
		return p.AllowUnknownQuality
	}

	if quality.QualityScore < p.MinQualityScore {
		return false
	}

	switch quality.Category {
	case dictionary.CategoryPrimary,
		dictionary.CategoryCore,
		dictionary.CategoryUseful:
		return true
	default:
		return false
	}
}
