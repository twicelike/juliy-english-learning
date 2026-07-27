package prompts

import "fmt"

// BuildExplainPrompt returns a system+user prompt instructing the AI to explain
// a vocabulary word and return structured JSON.
func BuildExplainPrompt(ctx PromptContext) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English vocabulary teacher helping a %s-level learner whose native language is %s.

Explain the English word "%s" (%s) thoroughly.

Return ONLY a JSON object — no markdown, no preamble — that exactly matches this structure:
{
  "summary": "A concise 1–2 sentence explanation appropriate for a %s learner.",
  "nuance": "Subtle connotations or usage nuances.",
  "commonMistakes": ["Mistake 1 that %s speakers make", "Mistake 2"],
  "examples": ["Example sentence 1.", "Example sentence 2.", "Example sentence 3."],
  "register": "formal | informal | neutral"
}`, ctx.UserLevel, ctx.NativeLanguage, ctx.Word, ctx.POS,
		ctx.UserLevel, ctx.NativeLanguage)
}

// BuildNuancePrompt returns a prompt focused on subtle meaning distinctions.
func BuildNuancePrompt(ctx PromptContext) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English vocabulary teacher helping a %s-level learner whose native language is %s.

Explain the nuances of the English word "%s" (%s) in detail.

Return ONLY a JSON object matching:
{
  "coreMeaning": "The fundamental meaning.",
  "subtleMeaning": "Subtle or extended meanings.",
  "usageNotes": ["Note 1", "Note 2"],
  "confusedWith": ["word1", "word2"]
}`, ctx.UserLevel, ctx.NativeLanguage, ctx.Word, ctx.POS)
}

// BuildExamplesPrompt returns a prompt to generate rich example sentences.
func BuildExamplesPrompt(ctx PromptContext) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English vocabulary teacher.

Generate 5 natural, varied example sentences for the word "%s" (%s), appropriate for a %s-level learner.

Return ONLY a JSON object matching:
{
  "word": "%s",
  "examples": ["Sentence 1.", "Sentence 2.", "Sentence 3.", "Sentence 4.", "Sentence 5."]
}`, ctx.Word, ctx.POS, ctx.UserLevel, ctx.Word)
}

// BuildMnemonicPrompt returns a prompt that creates a memorable mnemonic.
func BuildMnemonicPrompt(ctx PromptContext) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are a memory coach helping a %s-level English learner whose native language is %s.

Create a creative mnemonic to remember the word "%s" (%s).

Return ONLY a JSON object matching:
{
  "word": "%s",
  "mnemonic": "A catchy phrase or acronym.",
  "story": "A very short memorable story (2–3 sentences).",
  "visual": "Describe a vivid mental image to associate with this word."
}`, ctx.UserLevel, ctx.NativeLanguage, ctx.Word, ctx.POS, ctx.Word)
}
