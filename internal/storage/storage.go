package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vlostech/qz/internal/model"
)

const (
	questionPart = 1
	answerPart   = 2
)

var (
	ErrInvalidFilePath = errors.New("invalid file path")
)

// GetQuizItems returns all pairs of questions and answers from a given file.
//
// Any file should have alternating questions and answers (question goes first) that are separated by an empty line.
// Both can have 1..* rows.
//
// Example:
//
//	Question 1
//
//	Answer 1
//
//	Question 2 - Row 1
//	Question 2 - Row 2
//
//	Answer 2 - Row 1
//
// In example above, GetQuizItems returns two model.QuizSessionItem. The second item contains the question that consists
// of two rows.
func GetQuizItems(filePath string) ([]model.QuizSessionItem, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return extractQuizItems(file)
}

// SaveQuizItems saves questions to the given file. If the file does not exist, it will be created. If the file already
// exists, questions will be added to the end of the file. Returns an absolute path of the file. Returns
// ErrInvalidFilePath if the given file path is invalid.
func SaveQuizItems(filePath string, quizItems []model.QuizSessionItem) (string, error) {
	absFilePath, err := getAbsolutePath(filePath)

	if err != nil {
		return "", err
	}

	err = createMissingDirectoriesForFile(absFilePath)

	var pathErr *fs.PathError

	if err != nil {
		if errors.As(err, &pathErr) {
			return "", ErrInvalidFilePath
		}

		return "", err
	}

	file, err := os.OpenFile(absFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if errors.As(err, &pathErr) {
		return "", ErrInvalidFilePath
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			fmt.Printf("Error while closing file %s: %v\n", file.Name(), closeErr)
		}
	}()

	if err != nil {
		return "", err
	}

	builder := strings.Builder{}

	for _, quizItem := range quizItems {
		builder.WriteString(quizItem.Question)
		builder.WriteString("\n\n")
		builder.WriteString(quizItem.ExpectedAnswer)
		builder.WriteString("\n\n")
	}

	_, err = file.Write([]byte(builder.String()))

	if err != nil {
		return "", err
	}

	return absFilePath, nil
}

func extractQuizItems(r io.Reader) ([]model.QuizSessionItem, error) {
	curPart := questionPart
	isPreviousLineEmpty := true

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
					if isPreviousLineEmpty {
						continue
					}

					if curQuizItem == nil {
						continue
					}

					curPart = answerPart
					isPreviousLineEmpty = true
				} else {
					if curQuizItem == nil {
						curQuizItem = &model.QuizSessionItem{
							Index:    idx,
							Question: text,
						}
					} else {
						curQuizItem.Question += "\n" + text
					}

					isPreviousLineEmpty = false
				}
			}
		case answerPart:
			{
				if curQuizItem == nil {
					panic("quizItem was not initialized")
				}

				if text == "" {
					if isPreviousLineEmpty {
						continue
					}

					if curQuizItem.ExpectedAnswer == "" {
						continue
					}

					quizItems = append(quizItems, *curQuizItem)
					curQuizItem = nil
					idx++
					curPart = questionPart
					isPreviousLineEmpty = true
				} else {
					if curQuizItem.ExpectedAnswer == "" {
						curQuizItem.ExpectedAnswer = text
					} else {
						curQuizItem.ExpectedAnswer += "\n" + text
					}

					isPreviousLineEmpty = false
				}
			}
		}
	}

	if curQuizItem != nil {
		quizItems = append(quizItems, *curQuizItem)
	}

	return quizItems, nil
}

func getAbsolutePath(filePath string) (string, error) {
	if filepath.IsAbs(filePath) {
		return filePath, nil
	}

	if strings.HasPrefix(filePath, fmt.Sprintf("~%c", os.PathSeparator)) {
		homeDir, err := os.UserHomeDir()

		if err != nil {
			return "", err
		}

		absPath := path.Join(homeDir, filePath[2:])
		return absPath, nil
	}

	return filepath.Abs(filePath)
}

func createMissingDirectoriesForFile(filePath string) error {
	dirPath := filepath.Dir(filePath)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0644)
	}

	return nil
}
