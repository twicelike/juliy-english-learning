package learning

import (
	"math"
	"time"
)

// SM2Scheduler implements the Scheduler interface using the SuperMemo-2 algorithm.
type SM2Scheduler struct{}

func NewSM2Scheduler() Scheduler {
	return &SM2Scheduler{}
}

// CalculateNext calculates the new state for a word given the review quality.
func (s *SM2Scheduler) CalculateNext(p *LearningProgress, q Quality) *LearningProgress {
	next := &LearningProgress{
		ID:             p.ID,
		UserID:         p.UserID,
		WordID:         p.WordID,
		CreatedAt:      p.CreatedAt,
		LastReviewDate: time.Now(),
	}

	// Calculate Ease Factor
	// EF' = EF + (0.1 - (5 - q) * (0.08 + (5 - q) * 0.02))
	ef := p.EaseFactor
	if ef < 1.3 {
		ef = 2.5 // Default if it was unset or invalid
	}
	
	newEF := ef + (0.1 - float64(5-q)*(0.08+float64(5-q)*0.02))
	if newEF < 1.3 {
		newEF = 1.3
	}
	next.EaseFactor = newEF

	// Calculate Repetition & Interval
	if q < QualityHard {
		// Failed or blacked out
		next.Repetition = 0
		next.Interval = 1
	} else {
		// Passed
		next.Repetition = p.Repetition + 1
		
		if next.Repetition == 1 {
			next.Interval = 1
		} else if next.Repetition == 2 {
			next.Interval = 6
		} else {
			// I(n) = I(n-1) * EF
			next.Interval = int(math.Round(float64(p.Interval) * ef))
		}
	}

	// Calculate Next Review Date
	next.NextReviewDate = next.LastReviewDate.AddDate(0, 0, next.Interval)
	next.UpdatedAt = next.LastReviewDate

	return next
}

// IsDue reports whether a progress record is due for review.
func IsDue(p *LearningProgress) bool {
	return !time.Now().Before(p.NextReviewDate)
}

// IsLearned reports whether the word is considered mature/learned (interval >= 21).
func IsLearned(p *LearningProgress) bool {
	return p.Interval >= 21
}
