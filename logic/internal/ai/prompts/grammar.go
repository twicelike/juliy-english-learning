package prompts

import "fmt"

// BuildGrammarPrompt instructs the AI to check grammar and return structured errors.
func BuildGrammarPrompt(ctx PromptContext, text string) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English grammar teacher reviewing writing from a %s-level student whose native language is %s.

Analyse the following text for grammatical errors:

"""
%s
"""

Return ONLY a JSON object matching:
{
  "score": <integer 0–100, where 100 is perfect grammar>,
  "correctedText": "The full corrected version of the text.",
  "mistakes": [
    {
      "wrong": "the incorrect word or phrase",
      "correct": "the corrected form",
      "reason": "Short grammatical rule name (e.g. Past tense, Subject-verb agreement).",
      "explanation": "One-sentence explanation appropriate for a %s learner."
    }
  ],
  "tips": ["Tip 1 for this learner to avoid similar mistakes.", "Tip 2"]
}`, ctx.UserLevel, ctx.NativeLanguage, text, ctx.UserLevel)
}

// BuildVocabSuggestionPrompt instructs the AI to suggest better vocabulary.
func BuildVocabSuggestionPrompt(ctx PromptContext, text string) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English writing coach helping a %s-level learner whose native language is %s.

Review the following text and suggest more natural, precise, or contextually appropriate vocabulary improvements:

"""
%s
"""

Return ONLY a JSON object matching:
{
  "originalText": "%s",
  "rewrittenText": "The full text with your vocabulary improvements applied.",
  "suggestions": [
    {
      "original": "the word/phrase to replace",
      "suggested": "the improved word/phrase",
      "reason": "Why this is better.",
      "example": "A short example sentence using the suggested word."
    }
  ]
}`, ctx.UserLevel, ctx.NativeLanguage, text, text)
}
