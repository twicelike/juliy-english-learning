package prompts

import (
	"fmt"
	"strings"
)

// WordInput represents a vocabulary word passed to the quiz prompt builder.
type WordInput struct {
	Word string
	POS  string
	CEFR string
}

// BuildFillBlankPrompt instructs AI to select one word from the list and
// generate a natural fill-in-the-blank sentence.
func BuildFillBlankPrompt(ctx PromptContext, words []WordInput) string {
	ctx = ctx.defaults()

	var sb strings.Builder
	sb.WriteString("Available words:\n")
	for _, w := range words {
		cefr := w.CEFR
		if cefr == "" {
			cefr = "unknown"
		}
		sb.WriteString(fmt.Sprintf("  - \"%s\" (%s, CEFR: %s)\n", w.Word, w.POS, cefr))
	}

	return fmt.Sprintf(`You are an English quiz generator for a %s-level learner whose native language is %s.

%s
Choose ONE word from the list above that produces the most natural, unambiguous fill-in-the-blank sentence for a %s learner.

Create a sentence that:
  - Uses the word in its natural grammatical form (may be inflected).
  - Replaces the word with exactly three underscores: ___.
  - Is clear enough that only the chosen word is the correct answer.

Return ONLY a JSON object matching:
{
  "chosenWord": "the word you selected (base form)",
  "pos": "its part of speech",
  "sentence": "The sentence with ___ as the blank.",
  "answer": "The inflected form of the word used in the sentence.",
  "hint": "A very short grammar or meaning hint, or empty string."
}`, ctx.UserLevel, ctx.NativeLanguage, sb.String(), ctx.UserLevel)
}
