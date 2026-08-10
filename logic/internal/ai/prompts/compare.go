package prompts

import "fmt"

// BuildComparePrompt returns a prompt that compares two related words.
func BuildComparePrompt(ctx PromptContext, word2 string) string {
	ctx = ctx.defaults()
	return fmt.Sprintf(`You are an expert English vocabulary teacher helping a %s-level learner whose native language is %s.

Compare the words "%s" and "%s". Focus on meaning differences, register, and when to use each.

Return ONLY a JSON object matching:
{
  "word1": "%s",
  "word2": "%s",
  "similarity": "Brief description of what they share.",
  "differences": [
    {"aspect": "meaning", "word1": "...", "word2": "..."},
    {"aspect": "register", "word1": "...", "word2": "..."},
    {"aspect": "collocation", "word1": "...", "word2": "..."}
  ],
  "recommendedUsage": "A concise rule of thumb for when to use each word."
}`, ctx.UserLevel, ctx.NativeLanguage, ctx.Word, word2, ctx.Word, word2)
}
