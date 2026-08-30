package dictionary

import (
	"context"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo Repository
}

// NewService creates a new dictionary service.
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Search(ctx context.Context, query string) ([]WordDTO, error) {
	if query == "" {
		return []WordDTO{}, nil
	}

	// Fetch up to 50 candidates to allow grouping
	docs, err := s.repo.SearchPrefix(ctx, query, 50)
	if err != nil {
		return nil, err
	}

	// Group by Word
	grouped := make(map[string][]WordDocument)
	var order []string
	for _, doc := range docs {
		if _, ok := grouped[doc.Word]; !ok {
			order = append(order, doc.Word)
		}
		grouped[doc.Word] = append(grouped[doc.Word], doc)
	}

	// Sort the words by relevance
	lowerQuery := strings.ToLower(query)
	sort.Slice(order, func(i, j int) bool {
		w1 := strings.ToLower(order[i])
		w2 := strings.ToLower(order[j])

		// 1. Exact match
		if w1 == lowerQuery && w2 != lowerQuery {
			return true
		}
		if w2 == lowerQuery && w1 != lowerQuery {
			return false
		}

		// 2. Prefix match
		w1Starts := strings.HasPrefix(w1, lowerQuery)
		w2Starts := strings.HasPrefix(w2, lowerQuery)

		if w1Starts && !w2Starts {
			return true
		}
		if w2Starts && !w1Starts {
			return false
		}

		// 3. Shorter words first
		if len(w1) != len(w2) {
			return len(w1) < len(w2)
		}

		// 4. Alphabetical fallback
		return w1 < w2
	})

	// Map to DTOs and limit to 10 unique words
	dtos := make([]WordDTO, 0)
	for _, w := range order {
		dtos = append(dtos, mapDocsToDTO(grouped[w]))
		if len(dtos) >= 10 {
			break
		}
	}

	return dtos, nil
}

func (s *service) GetDetail(ctx context.Context, word string) (*WordDTO, error) {
	docs, err := s.repo.GetWords(ctx, word)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil // Not found
	}

	dto := mapDocsToDTO(docs)
	return &dto, nil
}

func (s *service) GetManyByIDs(ctx context.Context, ids []bson.ObjectID) ([]WordDTO, error) {
	if len(ids) == 0 {
		return []WordDTO{}, nil
	}

	docs, err := s.repo.GetManyByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	docMap := make(map[bson.ObjectID]WordDocument)
	for _, doc := range docs {
		docMap[doc.ID] = doc
	}

	var dtos []WordDTO
	for _, id := range ids {
		if doc, ok := docMap[id]; ok {
			dtos = append(dtos, mapToDTO(&doc))
		} else {
			// Log missing word (data integrity issue)
			// For now, we just skip it, but in production we might want to log this properly.
			continue
		}
	}

	return dtos, nil
}

func (s *service) GetRandom(ctx context.Context) (*WordDTO, error) {
	doc, err := s.repo.GetRandomWord(ctx)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	// Fetch all POS variants for this random word
	docs, err := s.repo.GetWords(ctx, doc.Word)
	if err != nil || len(docs) == 0 {
		dto := mapToDTO(doc)
		return &dto, nil
	}

	dto := mapDocsToDTO(docs)
	return &dto, nil
}

func (s *service) GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDTO, error) {
	doc, err := s.repo.GetWordOfTheDay(ctx, dateStr)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	// Fetch all POS variants for this word of the day
	docs, err := s.repo.GetWords(ctx, doc.Word)
	if err != nil || len(docs) == 0 {
		dto := mapToDTO(doc)
		return &dto, nil
	}

	dto := mapDocsToDTO(docs)
	return &dto, nil
}

// mapToDTO converts a WordDocument to a WordDTO exactly matching the frontend needs.
func mapToDTO(doc *WordDocument) WordDTO {
	dto := WordDTO{
		ID:             doc.ID.Hex(),
		Word:           doc.Word,
		POS:            doc.POS,
		FrequencyScore: doc.FrequencyScore,
		Forms:          []string{},
		Definitions:    []DefinitionDTO{},
		Antonyms:       []string{},
		Synonyms:       []string{},
	}

	// 1. Extract IPA and Audio (first available)
	for _, sound := range doc.Sounds {
		if sound.IPA != "" && dto.IPA == "" {
			dto.IPA = sound.IPA
		}
		if sound.Audio != "" && dto.Audio == "" {
			dto.Audio = sound.Audio
		}
	}

	// 2. Extract Forms
	for _, form := range doc.Forms {
		dto.Forms = append(dto.Forms, form.Form)
	}

	// 3. Extract Definitions and Examples
	for i, sense := range doc.Senses {
		for _, gloss := range sense.Glosses {
			defDTO := DefinitionDTO{
				WordID:       doc.ID.Hex(),
				SenseID:      sense.ID,
				SenseIndex:   i,
				PartOfSpeech: doc.POS,
				Definition:   gloss,
				Examples:     []string{},
			}
			
			if sense.Quality != nil {
				defDTO.Category = string(sense.Quality.Category)
			}
			count := 0
			for _, ex := range sense.Examples {
				if ex.Text != "" {
					defDTO.Examples = append(defDTO.Examples, ex.Text)
					count++
					if count >= 2 {
						break
					}
				}
			}
			dto.Definitions = append(dto.Definitions, defDTO)
		}
	}

	// 4. Extract Antonyms
	for _, ant := range doc.Antonyms {
		dto.Antonyms = append(dto.Antonyms, ant.Word)
	}

	// 5. Extract Synonyms
	for _, syn := range doc.Synonyms {
		dto.Synonyms = append(dto.Synonyms, syn.Word)
	}

	return dto
}

// mapDocsToDTO merges multiple documents of the same word (e.g. different POS) into a single WordDTO.
func mapDocsToDTO(docs []WordDocument) WordDTO {
	if len(docs) == 0 {
		return WordDTO{}
	}

	dto := WordDTO{
		ID:          docs[0].ID.Hex(),
		Word:        docs[0].Word,
		Forms:       []string{},
		Definitions: []DefinitionDTO{},
		Antonyms:    []string{},
		Synonyms:    []string{},
	}

	formSet := make(map[string]bool)
	antonymSet := make(map[string]bool)
	synonymSet := make(map[string]bool)

	var posList []string
	posSet := make(map[string]bool)

	for _, doc := range docs {
		if doc.POS != "" && !posSet[doc.POS] {
			posSet[doc.POS] = true
			posList = append(posList, doc.POS)
		}

		// 1. Extract IPA and Audio (first available across all docs)
		for _, sound := range doc.Sounds {
			if sound.IPA != "" && dto.IPA == "" {
				dto.IPA = sound.IPA
			}
			if sound.Audio != "" && dto.Audio == "" {
				dto.Audio = sound.Audio
			}
		}

		// 2. Extract Forms
		for _, form := range doc.Forms {
			if !formSet[form.Form] {
				formSet[form.Form] = true
				dto.Forms = append(dto.Forms, form.Form)
			}
		}

		// 3. Extract Definitions and Examples
		for i, sense := range doc.Senses {
			for _, gloss := range sense.Glosses {
				defDTO := DefinitionDTO{
					WordID:       doc.ID.Hex(),
					SenseID:      sense.ID,
					SenseIndex:   i,
					PartOfSpeech: doc.POS,
					Definition:   gloss,
					Examples:     []string{},
				}
				count := 0
				for _, ex := range sense.Examples {
					if ex.Text != "" {
						defDTO.Examples = append(defDTO.Examples, ex.Text)
						count++
						if count >= 2 {
							break
						}
					}
				}
				dto.Definitions = append(dto.Definitions, defDTO)
			}
		}

		// 4. Extract Antonyms
		for _, ant := range doc.Antonyms {
			if !antonymSet[ant.Word] {
				antonymSet[ant.Word] = true
				dto.Antonyms = append(dto.Antonyms, ant.Word)
			}
		}

		// 5. Extract Synonyms
		for _, syn := range doc.Synonyms {
			if !synonymSet[syn.Word] {
				synonymSet[syn.Word] = true
				dto.Synonyms = append(dto.Synonyms, syn.Word)
			}
		}
	}

	// Join all POS together
	importStrings := false
	if len(posList) > 0 {
		importStrings = true
	}
	_ = importStrings
	// I will just use string concatenation to avoid needing to import strings if it's not imported
	posStr := ""
	for i, p := range posList {
		if i > 0 {
			posStr += ", "
		}
		posStr += p
	}
	dto.POS = posStr

	return dto
}
