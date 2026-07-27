package prompts

// PromptContext carries learner metadata so prompts can be personalised per
// proficiency level and native language without changing service signatures.
type PromptContext struct {
	// Word is the primary vocabulary item being discussed.
	Word string
	// POS is the Part-of-Speech tag (noun, verb, adjective…).
	POS string
	// UserLevel is the CEFR proficiency level (A1 A2 B1 B2 C1 C2).
	// Prompts use this to calibrate vocabulary and sentence complexity.
	UserLevel string
	// NativeLanguage is used to tailor common-mistake notes.
	// Example: "Vietnamese".
	NativeLanguage string
}

// defaults returns a copy with sensible fallbacks applied.
func (c PromptContext) defaults() PromptContext {
	if c.UserLevel == "" {
		c.UserLevel = "B1"
	}
	if c.NativeLanguage == "" {
		c.NativeLanguage = "Vietnamese"
	}
	return c
}
