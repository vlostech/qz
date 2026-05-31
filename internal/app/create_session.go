package app

import (
	"github.com/vlostech/qz/internal/domain"
)

type CreateSessionInput struct {
	FilePath   string
	Count      int
	RangeQuery domain.RangeGroup
}

type CreateSessionOutput struct {
	Session domain.QuizSession
}

func (u *UseCase) CreateSession(input CreateSessionInput) (CreateSessionOutput, error) {
	items, err := u.fileStorage.GetItems(input.FilePath)

	if err != nil {
		return CreateSessionOutput{}, err
	}

	indexes := u.prepareIndexes(input.RangeQuery, len(items))

	var count int

	if input.Count > 0 {
		count = input.Count
	} else {
		count = len(items)
	}

	randomIndexes := u.randomizationService.Randomize(indexes, count)

	randomSessionItems := make([]domain.SessionItem, len(randomIndexes))

	for i, idx := range randomIndexes {
		item := items[idx]

		randomSessionItems[i] = domain.SessionItem{
			CollectionItem: item,
			ActualAnswer:   "",
		}
	}

	session := domain.QuizSession{
		SessionItems: randomSessionItems,
	}

	output := CreateSessionOutput{
		Session: session,
	}

	return output, nil
}

func (u *UseCase) prepareIndexes(rangeQuery domain.RangeGroup, totalCount int) []int {
	var indexes []int

	if len(rangeQuery.Ranges) != 0 {
		for _, part := range rangeQuery.Ranges {
			if part.LastIndex == -1 {
				for i := part.FirstIndex; i < totalCount; i++ {
					indexes = append(indexes, i)
				}
			} else {
				for i := part.FirstIndex; i <= part.LastIndex; i++ {
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
