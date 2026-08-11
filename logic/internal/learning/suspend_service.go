package learning

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *service) SuspendWord(ctx context.Context, userID string, wordID string) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	wid, err := bson.ObjectIDFromHex(wordID)
	if err != nil {
		return err
	}

	return s.repo.SuspendWord(ctx, uid, wid)
}
