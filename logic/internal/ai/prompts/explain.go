package prompts

import "fmt"

type SensePromptData struct {
	Index      int
	Definition string
}

// BuildBatchExplainPrompt returns a system+user prompt instructing the AI to explain
// a batch of dictionary senses for a vocabulary word and return structured JSON.
func BuildBatchExplainPrompt(ctx PromptContext, batchSenses []SensePromptData) string {
	ctx = ctx.defaults()

	sensesBlock := ""
	for _, s := range batchSenses {
		sensesBlock += fmt.Sprintf("Sense %d:\n%s\n\n", s.Index, s.Definition)
	}

	return fmt.Sprintf(`You are an expert English vocabulary teacher helping a %s-level learner whose native language is %s.

I will give you a batch of distinct senses/definitions for the English word "%s" (POS: %s).
You must explain ALL of these senses thoroughly.
Do NOT skip any sense.
Explain each sense independently.
Do NOT merge the meanings.
Do NOT generate Vietnamese translations unless explicitly instructed. The explanation should focus on helping the user understand and use that specific sense.

Here is the batch of senses to explain:
%s

Return ONLY a JSON object — no markdown, no preamble — that exactly matches this structure:
{
  "explanations": [
    {
      "senseIndex": <exact integer from the sense block>,
      "explanation": "A concise 1-2 sentence explanation appropriate for a %s learner focusing on this specific sense.",
      "examples": ["Example sentence 1.", "Example sentence 2."],
      "commonMistakes": ["Mistake 1 that %s speakers make with this specific sense", "Mistake 2"],
      "usageNote": "Subtle connotations, register (formal/informal), or nuances for this sense."
    }
  ]
}`, ctx.UserLevel, ctx.NativeLanguage, ctx.Word, ctx.POS, sensesBlock, ctx.UserLevel, ctx.NativeLanguage)
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
