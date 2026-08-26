package translation

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type WordTranslation struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	WordID         bson.ObjectID `bson:"word_id" json:"word_id"`
	SenseID        string        `bson:"sense_id" json:"sense_id"`
	SourceLanguage string        `bson:"source_language" json:"source_language"`
	TargetLanguage string        `bson:"target_language" json:"target_language"`
	Translations   []string      `bson:"translations" json:"translations"`
	Provider       string        `bson:"provider" json:"provider"`
	Model          string        `bson:"model" json:"model"`
	Version        int           `bson:"version" json:"version"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
}
