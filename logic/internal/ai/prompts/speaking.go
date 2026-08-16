package prompts

import (
	"fmt"
	"strings"
)

// BuildSpeakingTaskPrompt generates the prompt for AI to create a speaking task.
func BuildSpeakingTaskPrompt(topic, taskType, difficulty string) string {
	var builder strings.Builder

	builder.WriteString("You are an expert English language tutor creating a speaking practice task.\n\n")
	builder.WriteString(fmt.Sprintf("Topic: %s\n", topic))
	builder.WriteString(fmt.Sprintf("Difficulty level: %s (adjust vocabulary and expectations accordingly)\n", difficulty))
	builder.WriteString(fmt.Sprintf("Task Type: %s\n\n", taskType))

	builder.WriteString("Generate a JSON response that precisely matches the requested schema.\n")
	builder.WriteString("Requirements based on Task Type:\n")

	switch taskType {
	case "scenario":
		builder.WriteString("- 'scenario': Provide a detailed roleplay context (e.g., 'You are checking into a hotel...').\n")
		builder.WriteString("- 'prompt': State clearly what the user needs to say to start the conversation.\n")
		builder.WriteString("- 'instructions': Provide 3-4 bullet points of what they must include.\n")
		builder.WriteString("- 'hints': Provide 2-3 helpful hints on how to approach the situation.\n")
		builder.WriteString("- 'suggested_vocabulary': Provide 4-6 useful words or phrases.\n")
		builder.WriteString("- 'duration_seconds': Around 60 to 90 seconds.\n")
	case "quick_practice":
		builder.WriteString("- 'scenario': Leave this empty.\n")
		builder.WriteString("- 'prompt': Provide a single, clear question related to the topic.\n")
		builder.WriteString("- 'instructions': Provide 2-3 bullet points guiding their answer.\n")
		builder.WriteString("- 'hints': Provide 1-2 quick tips.\n")
		builder.WriteString("- 'suggested_vocabulary': Provide 3-4 related words.\n")
		builder.WriteString("- 'duration_seconds': Around 45 to 60 seconds.\n")
	case "free_talk":
		builder.WriteString("- 'scenario': Leave this empty.\n")
		builder.WriteString("- 'prompt': Provide an open-ended discussion prompt.\n")
		builder.WriteString("- 'instructions': Provide 1-2 broad ideas to talk about.\n")
		builder.WriteString("- 'hints': Leave this empty.\n")
		builder.WriteString("- 'suggested_vocabulary': Provide 3-5 words that might spark ideas.\n")
		builder.WriteString("- 'duration_seconds': Around 120 seconds.\n")
	}

	builder.WriteString("\nEnsure the output is strictly valid JSON matching the schema.")
	return builder.String()
}

// BuildSpeakingEvaluationPrompt generates the prompt for AI to evaluate a speaking recording.
func BuildSpeakingEvaluationPrompt(taskSnapshotJSON string) string {
	var builder strings.Builder

	builder.WriteString("You are an expert IELTS and English language evaluator.\n")
	builder.WriteString("You have been provided with an audio recording of a student speaking.\n\n")
	
	builder.WriteString("Here is the context of the task they were given:\n")
	builder.WriteString(taskSnapshotJSON)
	builder.WriteString("\n\n")

	builder.WriteString("Your job is to:\n")
	builder.WriteString("1. Listen to the audio and accurately transcribe exactly what the student said.\n")
	builder.WriteString("2. Evaluate their speaking performance based on the task context.\n")
	builder.WriteString("3. Extract 2-3 advanced or notable vocabulary words that the student actually used naturally in their speech.\n\n")

	builder.WriteString("Output a JSON object that EXACTLY matches the following structure. Do not include markdown blocks, just the JSON.\n")
	builder.WriteString(`{
  "transcript": "Exact transcription of the audio. If unintelligible, write '[unintelligible]'.",
  "score": <overall score from 0 to 100>,
  "fluency": { "score": <0-100>, "feedback": "Detailed feedback on fluency and coherence" },
  "grammar": { "score": <0-100>, "feedback": "Detailed feedback on grammar range and accuracy" },
  "vocabulary": { "score": <0-100>, "feedback": "Detailed feedback on lexical resource" },
  "pronunciation": { "score": <0-100>, "feedback": "Detailed feedback on pronunciation and intonation" },
  "task_completion": { "score": <0-100>, "feedback": "Did they answer the prompt and follow instructions?" },
  "overall_feedback": "A summary of their performance.",
  "strengths": ["list of 2-3 strengths"],
  "improvements": ["list of 2-3 areas to improve"],
  "new_used_vocabulary": [
    {
      "word": "a notable or advanced word the student used naturally",
      "reason": "Why this word was good in this context (e.g., 'Good B2-level word used naturally in context')"
    }
  ]
}`)
	
	return builder.String()
}
