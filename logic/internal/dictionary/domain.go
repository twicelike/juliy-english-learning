package dictionary

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// WordDocument maps exactly to the Wiktionary MongoDB schema.
// No extra fields are added that don't exist in the database.
type WordDocument struct {
	ID       bson.ObjectID `bson:"_id"`
	Word     string        `bson:"word"`
	POS      string        `bson:"pos"`
	Sounds   []SoundDoc    `bson:"sounds"`
	Forms    []FormDoc     `bson:"forms"`
	Senses   []SenseDoc    `bson:"senses"`
	Synonyms []RelatedDoc  `bson:"synonyms"`
	Antonyms []RelatedDoc  `bson:"antonyms"`
}

type SoundDoc struct {
	IPA   string `bson:"ipa,omitempty"`
	Audio string `bson:"audio,omitempty"`
}

type FormDoc struct {
	Form string   `bson:"form"`
	Tags []string `bson:"tags,omitempty"`
}

type SenseDoc struct {
	Glosses  []string     `bson:"glosses,omitempty"`
	Examples []ExampleDoc `bson:"examples,omitempty"`
	Tags     []string     `bson:"tags,omitempty"`
}

type ExampleDoc struct {
	Text string `bson:"text"`
}

type RelatedDoc struct {
	Word   string `bson:"word"`
	Source string `bson:"source,omitempty"`
}

// -----------------------------------------------------------------------------
// DTOs (Data Transfer Objects) used by the Frontend
// -----------------------------------------------------------------------------

type WordDTO struct {
	Word        string          `json:"word"`
	POS         string          `json:"pos"`
	IPA         string          `json:"ipa"` // Derived from the first sound containing IPA
	Forms       []string        `json:"forms"`
	Definitions []DefinitionDTO `json:"definitions"`
	Antonyms    []string        `json:"antonyms"`
	Synonyms    []string        `json:"synonyms"`
}

type DefinitionDTO struct {
	Definition string   `json:"definition"`
	Examples   []string `json:"examples"`
}

// -----------------------------------------------------------------------------
// Interfaces
// -----------------------------------------------------------------------------

type Repository interface {
	SearchPrefix(ctx context.Context, prefix string, limit int) ([]WordDocument, error)
	GetWord(ctx context.Context, word string) (*WordDocument, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*WordDocument, error)
	GetManyByIDs(ctx context.Context, ids []bson.ObjectID) ([]WordDocument, error)
	GetManyByWords(ctx context.Context, words []string) ([]WordDocument, error)
	GetRandomWord(ctx context.Context) (*WordDocument, error)
	GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDocument, error)
	// GetDistractors returns N random words with the same POS, excluding the given id.
	GetDistractors(ctx context.Context, pos string, excludeID bson.ObjectID, n int) ([]WordDocument, error)
}

type Service interface {
	Search(ctx context.Context, query string) ([]WordDTO, error)
	GetDetail(ctx context.Context, word string) (*WordDTO, error)
	GetRandom(ctx context.Context) (*WordDTO, error)
	GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDTO, error)
}
