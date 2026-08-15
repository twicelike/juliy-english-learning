package enricher

import (
	"context"
	"fmt"

	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/translation"
	"github.com/english/backend/internal/wordset"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type sessionEnricher struct {
	transSvc translation.Service
	wsSvc    wordset.Service
}

func NewSessionEnricher(transSvc translation.Service, wsSvc wordset.Service) learning.SessionEnricher {
	return &sessionEnricher{
		transSvc: transSvc,
		wsSvc:    wsSvc,
	}
}

func (e *sessionEnricher) EnrichItems(ctx context.Context, userID bson.ObjectID, targetLang string, items []learning.SessionItemDTO) ([]learning.SessionItemDTO, error) {
	if len(items) == 0 {
		return items, nil
	}

	wordIDs := make([]bson.ObjectID, 0, len(items))
	for _, item := range items {
		wid, err := bson.ObjectIDFromHex(item.WordID)
		if err == nil {
			wordIDs = append(wordIDs, wid)
		}
	}

	// Batch Fetch Translations
	translationsMap := make(map[string]*translation.WordTranslation)
	if e.transSvc != nil {
		translations, _ := e.transSvc.GetTranslationsByWordIDs(ctx, wordIDs, targetLang)
		for i := range translations {
			key := fmt.Sprintf("%s:%d", translations[i].WordID.Hex(), translations[i].SenseIndex)
			translationsMap[key] = &translations[i]
		}
	}

	// Batch Fetch Annotations
	annotationsMap := make(map[bson.ObjectID]*wordset.UserAnnotation)
	if e.wsSvc != nil {
		anns, _ := e.wsSvc.GetAnnotationsByUserAndWordIDs(ctx, userID, wordIDs)
		annotationsMap = anns
	}

	for i := range items {
		wid, err := bson.ObjectIDFromHex(items[i].WordID)
		if err != nil {
			continue
		}

		transKey := fmt.Sprintf("%s:%d", wid.Hex(), items[i].SenseIndex)
		if t, ok := translationsMap[transKey]; ok {
			items[i].TranslationCandidates = t.Translations
		}

		if a, ok := annotationsMap[wid]; ok {
			items[i].Annotation = &learning.UserAnnotationDTO{
				Translation: a.Text,
			}
		}
	}

	return items, nil
}
