package dictionary

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// WordDocument maps exactly to the Wiktionary MongoDB schema.
// No extra fields are added that don't exist in the database.
type WordDocument struct {
	ID             bson.ObjectID `bson:"_id"`
	Word           string        `bson:"word"`
	POS            string        `bson:"pos"`
	FrequencyScore int           `bson:"frequency_score,omitempty"`
	Sounds         []SoundDoc    `bson:"sounds"`
	Forms          []FormDoc     `bson:"forms"`
	Senses         []SenseDoc    `bson:"senses"`
	Synonyms       []RelatedDoc  `bson:"synonyms"`
	Antonyms       []RelatedDoc  `bson:"antonyms"`
}

type SoundDoc struct {
	IPA   string `bson:"ipa,omitempty"`
	Audio string `bson:"audio,omitempty"`
}

type FormDoc struct {
	Form string   `bson:"form"`
	Tags []string `bson:"tags,omitempty"`
}

type SenseCategory string

const (
	CategoryPrimary SenseCategory = "PRIMARY"
	CategoryCore    SenseCategory = "CORE"
	CategoryUseful  SenseCategory = "USEFUL"
	CategoryNiche   SenseCategory = "NICHE"
	CategoryRare    SenseCategory = "RARE"
	CategoryIgnore  SenseCategory = "IGNORE"
)

type SenseQuality struct {
	QualityScore      int           `bson:"quality_score,omitempty" json:"quality_score"`
	FrequencyScore    int           `bson:"frequency_score,omitempty" json:"frequency_score"`
	Category          SenseCategory `bson:"category,omitempty" json:"category"`
	EnrichmentVersion string        `bson:"enrichment_version,omitempty" json:"enrichment_version,omitempty"`
	EnrichedAt        *time.Time    `bson:"enriched_at,omitempty" json:"enriched_at,omitempty"`
}

type SenseDoc struct {
	ID       string        `bson:"id,omitempty"` // Stable Sense ID (UUID)
	Glosses  []string      `bson:"glosses,omitempty"`
	Examples []ExampleDoc  `bson:"examples,omitempty"`
	Tags     []string      `bson:"tags,omitempty"`
	Quality  *SenseQuality `bson:"quality,omitempty"`
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
	ID             string          `json:"id"`
	Word           string          `json:"word"`
	POS            string          `json:"pos"`
	FrequencyScore int             `json:"frequency_score,omitempty"`
	IPA            string          `json:"ipa"` // Derived from the first sound containing IPA
	Audio          string          `json:"audio"`
	Forms          []string        `json:"forms"`
	Definitions    []DefinitionDTO `json:"definitions"`
	Antonyms       []string        `json:"antonyms"`
	Synonyms       []string        `json:"synonyms"`
}

type DefinitionDTO struct {
	WordID       string   `json:"word_id,omitempty"`
	SenseID      string   `json:"sense_id"`    // Stable Sense ID
	SenseIndex   int      `json:"sense_index"` // Deprecated: keep for backward compatibility
	PartOfSpeech string   `json:"partOfSpeech,omitempty"`
	Category     string   `json:"category,omitempty"` // Added for filtering during translation
	Definition   string   `json:"definition"`
	Examples     []string `json:"examples"`
}

// -----------------------------------------------------------------------------
// Interfaces
// -----------------------------------------------------------------------------

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	SearchPrefix(ctx context.Context, prefix string, limit int) ([]WordDocument, error)
	GetWord(ctx context.Context, word string) (*WordDocument, error)
	GetWords(ctx context.Context, word string) ([]WordDocument, error)
	GetByID(ctx context.Context, id bson.ObjectID) (*WordDocument, error)
	GetManyByIDs(ctx context.Context, ids []bson.ObjectID) ([]WordDocument, error)
	GetManyByWords(ctx context.Context, words []string) ([]WordDocument, error)
	GetTopWordsNotIn(ctx context.Context, excludeIDs []bson.ObjectID, limit int) ([]WordDocument, error)
	GetRandomWord(ctx context.Context) (*WordDocument, error)
	GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDocument, error)
	// GetDistractors returns N random words with the same POS, excluding the given id.
	GetDistractors(ctx context.Context, pos string, excludeID bson.ObjectID, n int) ([]WordDocument, error)
	// GetDistractorsBulk returns N random words with the same POS, excluding the given list of ids.
	GetDistractorsBulk(ctx context.Context, pos string, excludeIDs []bson.ObjectID, n int) ([]WordDocument, error)
}

type Service interface {
	Search(ctx context.Context, query string) ([]WordDTO, error)
	GetDetail(ctx context.Context, word string) (*WordDTO, error)
	GetManyByIDs(ctx context.Context, ids []bson.ObjectID) ([]WordDTO, error)
	GetRandom(ctx context.Context) (*WordDTO, error)
	GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDTO, error)
}
