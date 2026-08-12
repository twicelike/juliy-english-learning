package dictionary

import (
	"context"
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

	docs, err := s.repo.SearchPrefix(ctx, query, 10)
	if err != nil {
		return nil, err
	}

	dtos := make([]WordDTO, 0, len(docs))
	for _, doc := range docs {
		dtos = append(dtos, mapToDTO(&doc))
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
		Word:        doc.Word,
		POS:         doc.POS,
		Forms:       []string{},
		Definitions: []DefinitionDTO{},
		Antonyms:    []string{},
		Synonyms:    []string{},
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
	for _, sense := range doc.Senses {
		for _, gloss := range sense.Glosses {
			defDTO := DefinitionDTO{
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
		for _, sense := range doc.Senses {
			for _, gloss := range sense.Glosses {
				defDTO := DefinitionDTO{
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
