package learning

import (
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AIBatchFillBlankResponse struct {
	Exercises []AIFillBlankItem `json:"exercises"`
}

type AIFillBlankItem struct {
	ReqID    string `json:"req_id"`
	Word     string `json:"word"`
	Question string `json:"question"` // Sentence with ___
	Answer   string `json:"answer"`
}

type AIBatchMCQResponse struct {
	Exercises []AIMCQItem `json:"exercises"`
}

type AIMCQItem struct {
	ReqID   string   `json:"req_id"`
	Word    string   `json:"word"`
	Choices []string `json:"choices"` // Exactly 4
	Answer  string   `json:"answer"`  // Must exactly match one of the choices
}

type WordGenerationContext struct {
	ReqID             string
	WordID            bson.ObjectID
	SenseID           string
	Word              string
	POS               string
	Definition        string
	ExistingSentences []string
}

func buildBatchFillBlankPrompt(words []WordGenerationContext) string {
	var sb strings.Builder
	sb.WriteString("You are an expert English teacher creating fill-in-the-blank exercises.\n")
	sb.WriteString("Generate EXACTLY ONE natural fill-in-the-blank sentence for each word provided.\n\n")

	sb.WriteString("Rules:\n")
	sb.WriteString("- The sentence MUST use the word naturally, optionally inflected (e.g. 'run' -> 'running').\n")
	sb.WriteString("- The blank must be exactly three underscores: ___\n")
	sb.WriteString("- The answer must be the exact word that fits in the blank.\n")
	sb.WriteString("- DO NOT use any of the Existing Sentences. Write a COMPLETELY NEW sentence with a different context, subject, or tense.\n\n")

	sb.WriteString("Input words:\n")
	for _, w := range words {
		sb.WriteString(fmt.Sprintf("- req_id: %s\n  word: %s (%s)\n  definition: %s\n", w.ReqID, w.Word, w.POS, w.Definition))
		if len(w.ExistingSentences) > 0 {
			sb.WriteString(fmt.Sprintf("  existing_sentences: [%s]\n", strings.Join(w.ExistingSentences, " | ")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Return a JSON object strictly matching this schema:
{
  "exercises": [
    {
      "req_id": "the exact req_id from the input",
      "word": "the base word",
      "question": "The sentence with ___",
      "answer": "The word that fits the blank"
    }
  ]
}`)
	return sb.String()
}

func buildBatchMCQPrompt(words []WordGenerationContext) string {
	var sb strings.Builder
	sb.WriteString("You are an expert English teacher creating Multiple Choice vocabulary questions.\n")
	sb.WriteString("Generate EXACTLY ONE multiple choice question for each word provided.\n\n")

	sb.WriteString("Rules:\n")
	sb.WriteString("- The correct answer must be the target word.\n")
	sb.WriteString("- Provide exactly 4 unique choices (1 correct, 3 distractors).\n")
	sb.WriteString("- The distractors must be the same part of speech as the target word.\n")
	sb.WriteString("- The distractors MUST NOT be synonyms that could also perfectly fit the definition.\n")
	sb.WriteString("- The choices should just be words, not definitions.\n\n")

	sb.WriteString("Input words:\n")
	for _, w := range words {
		sb.WriteString(fmt.Sprintf("- req_id: %s\n  word: %s (%s)\n  definition: %s\n\n", w.ReqID, w.Word, w.POS, w.Definition))
	}

	sb.WriteString(`Return a JSON object strictly matching this schema:
{
  "exercises": [
    {
      "req_id": "the exact req_id from the input",
      "word": "the base word",
      "choices": ["distractor1", "correct_answer", "distractor2", "distractor3"],
      "answer": "correct_answer"
    }
  ]
}`)
	return sb.String()
}

// -----------------------------------------------------------------------------
// Validation Logic
// -----------------------------------------------------------------------------

func validateMCQItem(item AIMCQItem) error {
	if len(item.Choices) != 4 {
		return errors.New("must have exactly 4 choices")
	}

	hasCorrect := false
	unique := make(map[string]bool)
	for _, c := range item.Choices {
		c = strings.TrimSpace(strings.ToLower(c))
		if unique[c] {
			return errors.New("choices must be unique")
		}
		unique[c] = true
		if c == strings.TrimSpace(strings.ToLower(item.Answer)) {
			hasCorrect = true
		}
	}

	if !hasCorrect {
		return errors.New("answer must match one of the choices exactly")
	}
	// The answer must be the target word (or at least closely match it)
	if strings.TrimSpace(strings.ToLower(item.Answer)) != strings.TrimSpace(strings.ToLower(item.Word)) {
		return errors.New("answer must be the target word")
	}

	return nil
}

func validateFillBlankItem(item AIFillBlankItem) error {
	if !strings.Contains(item.Question, "___") {
		return errors.New("question must contain ___")
	}
	if strings.TrimSpace(item.Answer) == "" {
		return errors.New("answer cannot be empty")
	}
	return nil
}
