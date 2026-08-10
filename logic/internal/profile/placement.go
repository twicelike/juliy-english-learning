package profile

// internalPlacementQuestion includes the CorrectAnswer for backend evaluation.
type internalPlacementQuestion struct {
	ID            string
	Type          QuestionType
	Question      string
	Options       []string
	CorrectAnswer string
	Topic         string // e.g. "Food", "Business", "Prepositions", "Tenses"
}

// 15 hardcoded questions for the Placement Test (5 Vocab, 5 Grammar, 5 Usage)
var placementTestQuestions = []internalPlacementQuestion{
	// Vocabulary (5)
	{
		ID:            "v1",
		Type:          TypeSynonym,
		Question:      "Which word is closest in meaning to 'abundant'?",
		Options:       []string{"plentiful", "scarce", "expensive", "small"},
		CorrectAnswer: "plentiful",
		Topic:         "Adjective",
	},
	{
		ID:            "v2",
		Type:          TypeFill,
		Question:      "I need to go to the grocery store to buy some fresh ________.",
		Options:       []string{"furniture", "produce", "vehicles", "clothing"},
		CorrectAnswer: "produce",
		Topic:         "Food",
	},
	{
		ID:            "v3",
		Type:          TypeSynonym,
		Question:      "Which word is a synonym for 'meticulous'?",
		Options:       []string{"careless", "sloppy", "careful", "rushed"},
		CorrectAnswer: "careful",
		Topic:         "Adjective",
	},
	{
		ID:            "v4",
		Type:          TypeFill,
		Question:      "The company's new policy will ________ all employees starting next month.",
		Options:       []string{"effect", "affect", "afford", "effort"},
		CorrectAnswer: "affect",
		Topic:         "Business",
	},
	{
		ID:            "v5",
		Type:          TypeSynonym,
		Question:      "What is the meaning of 'lucrative'?",
		Options:       []string{"profitable", "dangerous", "boring", "complicated"},
		CorrectAnswer: "profitable",
		Topic:         "Business",
	},

	// Grammar (5)
	{
		ID:            "g1",
		Type:          TypeGrammar,
		Question:      "He ________ to school yesterday.",
		Options:       []string{"go", "goes", "went", "going"},
		CorrectAnswer: "went",
		Topic:         "Past Tense",
	},
	{
		ID:            "g2",
		Type:          TypeGrammar,
		Question:      "I ________ living in this city for five years.",
		Options:       []string{"am", "was", "have been", "had"},
		CorrectAnswer: "have been",
		Topic:         "Perfect Tense",
	},
	{
		ID:            "g3",
		Type:          TypeGrammar,
		Question:      "If I ________ you, I would study harder.",
		Options:       []string{"am", "was", "were", "be"},
		CorrectAnswer: "were",
		Topic:         "Conditional",
	},
	{
		ID:            "g4",
		Type:          TypeGrammar,
		Question:      "She is very good ________ playing the piano.",
		Options:       []string{"in", "at", "on", "with"},
		CorrectAnswer: "at",
		Topic:         "Prepositions",
	},
	{
		ID:            "g5",
		Type:          TypeGrammar,
		Question:      "By this time next year, I ________ my degree.",
		Options:       []string{"will finish", "will have finished", "finish", "finished"},
		CorrectAnswer: "will have finished",
		Topic:         "Future Perfect",
	},

	// Usage / Phrasal Verbs (5)
	{
		ID:            "u1",
		Type:          TypeFill,
		Question:      "Can you please ________ the lights? It's too dark in here.",
		Options:       []string{"turn off", "turn on", "turn up", "turn down"},
		CorrectAnswer: "turn on",
		Topic:         "Phrasal Verbs",
	},
	{
		ID:            "u2",
		Type:          TypeFill,
		Question:      "I'm looking ________ to hearing from you soon.",
		Options:       []string{"for", "forward", "at", "up"},
		CorrectAnswer: "forward",
		Topic:         "Prepositions",
	},
	{
		ID:            "u3",
		Type:          TypeFill,
		Question:      "The meeting was called ________ because the manager was sick.",
		Options:       []string{"off", "up", "out", "down"},
		CorrectAnswer: "off",
		Topic:         "Phrasal Verbs",
	},
	{
		ID:            "u4",
		Type:          TypeFill,
		Question:      "I can't put ________ with this noise any longer!",
		Options:       []string{"up", "on", "out", "off"},
		CorrectAnswer: "up",
		Topic:         "Phrasal Verbs",
	},
	{
		ID:            "u5",
		Type:          TypeFill,
		Question:      "She takes ________ her mother; they have the same eyes and smile.",
		Options:       []string{"after", "care", "over", "off"},
		CorrectAnswer: "after",
		Topic:         "Phrasal Verbs",
	},
}

// evaluateTest maps a score out of 15 to a CEFR level.
// 0-3 (A1), 4-6 (A2), 7-9 (B1), 10-12 (B2), 13-14 (C1), 15 (C2)
func evaluateTest(answers map[string]string) (score int, cefr string, strengths, weaknesses []string) {
	topicScores := make(map[string]int)
	topicMax := make(map[string]int)

	for _, q := range placementTestQuestions {
		topicMax[q.Topic]++
		userAnswer, ok := answers[q.ID]
		if ok && userAnswer == q.CorrectAnswer {
			score++
			topicScores[q.Topic]++
		}
	}

	// Calculate CEFR
	switch {
	case score <= 3:
		cefr = "A1"
	case score <= 6:
		cefr = "A2"
	case score <= 9:
		cefr = "B1"
	case score <= 12:
		cefr = "B2"
	case score <= 14:
		cefr = "C1"
	default:
		cefr = "C2"
	}

	// Calculate strengths and weaknesses based on topic accuracy
	for topic, max := range topicMax {
		correct := topicScores[topic]
		ratio := float64(correct) / float64(max)

		// Very simplistic tag logic: if they got > 50% right in a topic, it's a strength.
		// If they got <= 50%, it's a weakness.
		if ratio > 0.5 {
			strengths = append(strengths, topic)
		} else {
			weaknesses = append(weaknesses, topic)
		}
	}

	// Ensure we don't return nil slices
	if strengths == nil {
		strengths = []string{}
	}
	if weaknesses == nil {
		weaknesses = []string{}
	}

	return score, cefr, strengths, weaknesses
}
