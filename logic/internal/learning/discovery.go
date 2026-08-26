package learning

import (
	"github.com/english/backend/internal/dictionary"
)

// IsMastered defines the threshold for a sense to be considered fully learned (Mature).
func IsMastered(p *LearningProgress) bool {
	// A sense is considered mastered if its SM-2 interval reaches 21 days (Mature).
	return p != nil && p.Interval >= 21
}

// DiscoverNextSense implements the Progressive Sense Discovery policy.
// It iterates through eligible senses in dictionary order.
// If a sense has no progress, it unlocks it.
// If a sense is currently being learned (not mastered), it blocks further discovery and returns it.
// If a sense is mastered, it continues to the next eligible sense.
// If all eligible senses are mastered, it returns nil.
func DiscoverNextSense(doc *dictionary.WordDocument, progresses map[string]*LearningProgress) *dictionary.SenseDoc {
	if doc == nil {
		return nil
	}

	policy := DefaultEligibilityPolicy

	for i := range doc.Senses {
		sense := &doc.Senses[i]

		// Filter out obscure or ineligible senses
		if !policy.IsEligible(sense.Quality) {
			continue
		}

		p, exists := progresses[sense.ID]
		if !exists {
			// No progress -> this is the next unlearned sense to unlock
			return sense
		}

		if !IsMastered(p) {
			// Progress exists but is not mastered -> user is currently learning this sense.
			// Stop unlocking and return this sense as the active frontier.
			return sense
		}

		// Progress exists and is mastered -> continue to the next eligible sense
	}

	// All eligible senses have been mastered
	return nil
}
