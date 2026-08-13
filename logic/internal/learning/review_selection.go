package learning

import (
	"errors"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReviewableWords struct {
	Due      []bson.ObjectID
	Learning []bson.ObjectID
	New      []bson.ObjectID
}

var ErrNoReviewableWords = errors.New("no_reviewable_words")
var ErrInvalidWordCount = errors.New("invalid_word_count")

// SelectReviewWords selects a list of words to review based on the given strategy and limit.
func SelectReviewWords(
	due []bson.ObjectID,
	learning []bson.ObjectID,
	newWords []bson.ObjectID,
	strategy ReviewStrategy,
	limit int,
) ([]bson.ObjectID, error) {
	if limit < 5 || limit > 25 {
		return nil, ErrInvalidWordCount
	}

	var totalAvailable int
	if strategy == ReviewStrategyFresh {
		totalAvailable = len(due) + len(newWords)
	} else {
		totalAvailable = len(due) + len(learning) + len(newWords)
	}

	if totalAvailable == 0 {
		return nil, ErrNoReviewableWords
	}

	// Adjust limit if there are fewer words available than requested
	if limit > totalAvailable {
		limit = totalAvailable
	}

	// Shuffle each category to avoid selecting the same words repeatedly
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	shuffle := func(slice []bson.ObjectID) {
		rng.Shuffle(len(slice), func(i, j int) {
			slice[i], slice[j] = slice[j], slice[i]
		})
	}

	// Make copies so we don't modify the originals
	dueCopy := append([]bson.ObjectID(nil), due...)
	learningCopy := append([]bson.ObjectID(nil), learning...)
	newWordsCopy := append([]bson.ObjectID(nil), newWords...)

	shuffle(dueCopy)
	shuffle(learningCopy)
	shuffle(newWordsCopy)

	var selected []bson.ObjectID

	take := func(source *[]bson.ObjectID, amount int) {
		if amount <= 0 {
			return
		}
		if amount > len(*source) {
			amount = len(*source)
		}
		selected = append(selected, (*source)[:amount]...)
		*source = (*source)[amount:]
	}

	remaining := func() int {
		return limit - len(selected)
	}

	switch strategy {
	case ReviewStrategyContinue, ReviewStrategySmart:
		// Priority: Due -> Learning -> New
		take(&dueCopy, remaining())
		take(&learningCopy, remaining())
		take(&newWordsCopy, remaining())

	case ReviewStrategyFresh:
		// Priority: Due -> New -> (Learning as fallback only if we absolutely have to, wait, spec says:
		// "Không lấy Learning trừ khi sau này product decision thay đổi."
		// Let's strictly take Due -> New.
		take(&dueCopy, remaining())
		take(&newWordsCopy, remaining())
	}

	return selected, nil
}
