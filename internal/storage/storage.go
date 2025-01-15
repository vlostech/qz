package storage

import (
	"bufio"
	"github.com/vlostech/qz/internal/model"
	"io"
	"os"
)

const (
	questionPart = 1
	answerPart   = 2
)

func GetQuizItems(filePath string) ([]model.QuizSessionItem, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return extractQuizItems(file)
}

func extractQuizItems(r io.Reader) ([]model.QuizSessionItem, error) {
	curPart := questionPart

	scanner := bufio.NewScanner(r)

	var quizItems []model.QuizSessionItem
	var curQuizItem *model.QuizSessionItem

	idx := 0

	for scanner.Scan() {
		text := scanner.Text()

		switch curPart {
		case questionPart:
			{
				if text == "" {
					if curQuizItem == nil {
						continue
					}

					curPart = answerPart
				} else {
					if curQuizItem == nil {
						curQuizItem = &model.QuizSessionItem{
							Index:    idx,
							Question: text,
						}
					} else {
						curQuizItem.Question += "\n" + text
					}
				}
			}
		case answerPart:
			{
				if curQuizItem == nil {
					panic("quizItem was not initialized")
				}

				if text == "" {
					if curQuizItem.ExpectedAnswer == "" {
						continue
					} else {
						quizItems = append(quizItems, *curQuizItem)
						curQuizItem = nil
						idx++
						curPart = questionPart
					}
				} else {
					if curQuizItem.ExpectedAnswer == "" {
						curQuizItem.ExpectedAnswer = text
					} else {
						curQuizItem.ExpectedAnswer += "\n" + text
					}
				}
			}
		}
	}

	quizItems = append(quizItems, *curQuizItem)

	return quizItems, nil
}
