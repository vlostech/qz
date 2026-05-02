package app

import "github.com/vlostech/qz/internal/domain"

type SaveQuestionsInput struct {
	FilePath string
	Items    []domain.CollectionItem
}

type SaveQuestionsOutput struct {
	AbsoluteFilePath string
}

func (u *UseCase) SaveQuestions(input SaveQuestionsInput) (SaveQuestionsOutput, error) {
	absoluteFilePath, err := u.fileStorage.SaveQuizItems(input.FilePath, input.Items)

	if err != nil {
		return SaveQuestionsOutput{}, err
	}

	output := SaveQuestionsOutput{
		AbsoluteFilePath: absoluteFilePath,
	}

	return output, nil
}
