package session

import (
	"github.com/vlostech/qz/internal/model"
	"github.com/vlostech/qz/internal/random"
	"github.com/vlostech/qz/internal/ranges"
	"github.com/vlostech/qz/internal/storage"
)

// CreateSession creates a new session with random questions.
func CreateSession(
	filePath string,
	count int,
	rangeQuery ranges.RangeQuery,
) (*model.QuizSession, error) {
	items, err := storage.GetQuizItems(filePath)

	if err != nil {
		return nil, err
	}

	indexes := getQuestionIndexes(rangeQuery, len(items))

	if count <= 0 {
		count = len(items)
	}

	randomIndexes := random.Randomize(indexes, count)
	randomItems := make([]model.QuizSessionItem, len(randomIndexes))

	for i, idx := range randomIndexes {
		randomItems[i] = items[idx]
	}

	session := &model.QuizSession{
		Items: randomItems,
	}

	return session, nil
}

func getQuestionIndexes(rangeQuery ranges.RangeQuery, totalCount int) []int {
	var indexes []int

	if len(rangeQuery.Parts) != 0 {
		for _, part := range rangeQuery.Parts {
			if part.CloseIndex == -1 {
				for i := part.OpenIndex; i < totalCount; i++ {
					indexes = append(indexes, i)
				}
			} else {
				for i := part.OpenIndex; i < part.CloseIndex; i++ {
					if i == totalCount {
						break
					}

					indexes = append(indexes, i)
				}
			}
		}
	} else {
		indexes = make([]int, totalCount)

		for i := range totalCount {
			indexes[i] = i
		}
	}

	return indexes
}
