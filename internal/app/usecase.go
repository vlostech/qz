package app

import "github.com/vlostech/qz/internal/domain"

type RandomizationService interface {
	Randomize(nums []int, count int) []int
}

type FileStorage interface {
	GetItems(filePath string) ([]domain.CollectionItem, error)
	SaveQuizItems(filePath string, items []domain.CollectionItem) (string, error)
}

type UseCase struct {
	fileStorage          FileStorage
	randomizationService RandomizationService
}

func NewUseCase(fileStorage FileStorage, randomizationService RandomizationService) *UseCase {
	return &UseCase{
		fileStorage:          fileStorage,
		randomizationService: randomizationService,
	}
}
