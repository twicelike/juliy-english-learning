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
	doc, err := s.repo.GetWord(ctx, word)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // Not found
	}
	
	dto := mapToDTO(doc)
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
	
	dto := mapToDTO(doc)
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
	
	dto := mapToDTO(doc)
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

	// 1. Extract IPA (first sound with ipa)
	for _, sound := range doc.Sounds {
		if sound.IPA != "" {
			dto.IPA = sound.IPA
			break
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
				Definition: gloss,
				Examples:   []string{},
			}
			for _, ex := range sense.Examples {
				if ex.Text != "" {
					defDTO.Examples = append(defDTO.Examples, ex.Text)
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
