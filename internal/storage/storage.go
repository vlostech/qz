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

	scanner := bufio.NewScanner(file)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return extractQuizItems(lines)
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

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		return "", err
	}

	fileText := string(fileBytes)

	lineBreak := getLineBreakForFileText(fileText)
	lines := strings.Split(fileText, lineBreak)

	existingQuizItems, err := extractQuizItems(lines)

	if err != nil {
		return "", err
	}

	m := make(map[string]bool)

	for _, existingQuizItem := range existingQuizItems {
		m[existingQuizItem.Question] = true
	}

	var newQuizItems []model.QuizSessionItem

	for _, quizItem := range quizItems {
		if m[quizItem.Question] {
			continue
		}

		newQuizItems = append(newQuizItems, quizItem)
	}

	builder := strings.Builder{}

	// This block appends additional line breaks to the end of the current file to make sure that
	// existing content will be separated from the new content:
	//
	// * If the last line of the file contains text, then we add two line breaks to the end of the
	//   line.
	//
	// * If the last line is empty, but the line above has text, then we add one line break.
	if len(lines) > 0 && lines[len(lines)-1] != "" {
		builder.WriteString(lineBreak + lineBreak)
	} else if len(lines) > 1 && lines[len(lines)-2] != "" {
		builder.WriteString(lineBreak)
	}

	for i, newQuizItem := range newQuizItems {
		builder.WriteString(strings.ReplaceAll(newQuizItem.Question, "\n", lineBreak))
		builder.WriteString(lineBreak + lineBreak)
		builder.WriteString(strings.ReplaceAll(newQuizItem.ExpectedAnswer, "\n", lineBreak))

		// Adds line breaks only between question-answer pairs.
		if i < len(newQuizItems)-1 {
			builder.WriteString(lineBreak + lineBreak)
		}
	}

	_, err = file.Write([]byte(builder.String()))

	if err != nil {
		return "", err
	}

	return absFilePath, nil
}

func extractQuizItems(lines []string) ([]model.QuizSessionItem, error) {
	curPart := questionPart
	isPreviousLineEmpty := true

	var quizItems []model.QuizSessionItem
	var curQuizItem *model.QuizSessionItem

	idx := 0

	for _, line := range lines {
		switch curPart {
		case questionPart:
			{
				if line == "" {
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
							Question: line,
						}
					} else {
						curQuizItem.Question += "\n" + line
					}

					isPreviousLineEmpty = false
				}
			}
		case answerPart:
			{
				if curQuizItem == nil {
					panic("quizItem was not initialized")
				}

				if line == "" {
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
						curQuizItem.ExpectedAnswer = line
					} else {
						curQuizItem.ExpectedAnswer += "\n" + line
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

func getLineBreakForFileText(fileText string) string {
	for _, char := range fileText {
		if char == '\r' {
			return "\r\n"
		}

		if char == '\n' {
			return "\n"
		}
	}

	return "\n"
}
