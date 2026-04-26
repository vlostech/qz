package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vlostech/qz/internal/ioext"
	"github.com/vlostech/qz/internal/model"
	"github.com/vlostech/qz/internal/ranges"
	"github.com/vlostech/qz/internal/session"
	"github.com/vlostech/qz/internal/storage"
)

var (
	filePath    string
	count       int
	rangeString string
)

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run a quiz",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCommand()
	},
}

func init() {
	Command.PersistentFlags().StringVarP(
		&filePath,
		"file",
		"f",
		"",
		"Path to the file that contains questions",
	)
	Command.PersistentFlags().IntVarP(
		&count,
		"count",
		"c",
		0,
		"Number of questions",
	)
	Command.PersistentFlags().StringVarP(
		&rangeString,
		"range",
		"r",
		"",
		"Range of questions",
	)
}

func runCommand() error {
	rangeQuery, err := ranges.ParseRange(rangeString)

	if err != nil {
		return err
	}

	s, err := session.CreateSession(filePath, count, rangeQuery)

	if err != nil {
		return err
	}

	err = runFirstPhase(s)

	if err != nil {
		return err
	}

	err = runSecondPhase(s)

	if err != nil {
		return err
	}

	return runSavePhase(s)
}

func runFirstPhase(session *model.QuizSession) error {
	fmt.Println("PHASE 1 - QUESTIONS")
	fmt.Println()

	for i := 0; i < len(session.Items); i++ {
		fmt.Printf("%v/%v\n", i+1, len(session.Items))
		fmt.Println()
		fmt.Println(session.Items[i].Question)
		fmt.Println()
		fmt.Println("Write your answer:")

		answer, err := ioext.GetMultilineString()

		if err != nil {
			return err
		}

		session.Items[i].ActualAnswer = answer

		fmt.Println()
	}

	return nil
}

func runSecondPhase(session *model.QuizSession) error {
	fmt.Println("PHASE 2 - ANSWERS")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; i < len(session.Items); i++ {
		fmt.Printf("%v/%v\n", i+1, len(session.Items))
		fmt.Println()
		fmt.Println("Question:")
		fmt.Println(session.Items[i].Question)
		fmt.Println()
		fmt.Println("Expected answer:")
		fmt.Println(session.Items[i].ExpectedAnswer)
		fmt.Println()
		fmt.Println("Actual answer:")
		fmt.Println(session.Items[i].ActualAnswer)
		fmt.Println()
		fmt.Println("Press Enter to continue...")

		if !scanner.Scan() {
			return scanner.Err()
		}
	}

	return nil
}

func runSavePhase(session *model.QuizSession) error {
	shouldSave, err := askIfQuestionsShouldBeSaved()

	if err != nil {
		return err
	}

	if !shouldSave {
		return nil
	}

	questionIndices, shouldQuit, err := askForQuestionsToSave(session)

	if err != nil {
		return err
	}

	if shouldQuit {
		return nil
	}

	chosenQuestions := make([]model.QuizSessionItem, len(questionIndices))

	for i, idx := range questionIndices {
		chosenQuestions[i] = session.Items[idx]
	}

	var absolutePath string

	for {
		path, err := askForPath()

		if err != nil {
			return err
		}

		absolutePath, err = storage.SaveQuizItems(path, chosenQuestions)

		if err != nil {
			if errors.Is(err, storage.ErrInvalidFilePath) {
				fmt.Println("ERROR: Provided path is invalid.")
				continue
			}

			return err
		}

		break
	}

	fmt.Println()
	fmt.Println("Questions are saved in the following file:")
	fmt.Println(absolutePath)

	return nil
}

func askIfQuestionsShouldBeSaved() (bool, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Would you like to save questions to a separate file (y/n)?")

		if !scanner.Scan() {
			return false, scanner.Err()
		}

		answer := strings.TrimSpace(scanner.Text())

		switch strings.ToLower(answer) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		default:
			fmt.Println("Provide valid answer.")
		}
	}
}

func askForPath() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Enter a file path:")

		if !scanner.Scan() {
			return "", scanner.Err()
		}

		path := strings.TrimSpace(scanner.Text())

		if path == "" {
			fmt.Println("Provide a valid file path.")
			continue
		}

		return path, nil
	}
}

func askForQuestionsToSave(session *model.QuizSession) ([]int, bool, error) {
	chosenQuestionIndices := make(map[int]struct{})

	for {
		showStatus(session, chosenQuestionIndices)
		fmt.Println()
		selectionRange, shouldQuit, shouldProceed, err := askForRange()

		if err != nil {
			return nil, false, err
		}

		if shouldQuit {
			return []int{}, true, err
		}

		toggleSelection(session, chosenQuestionIndices, selectionRange)

		if shouldProceed {
			if len(chosenQuestionIndices) == 0 {
				fmt.Println("No questions to save.")
				return []int{}, true, nil
			}

			break
		}
	}

	indices := make([]int, 0, len(chosenQuestionIndices))

	for idx := range chosenQuestionIndices {
		indices = append(indices, idx)
	}

	return indices, false, nil
}

func askForRange() (ranges.RangeQuery, bool, bool, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Choose questions or enter command (f, q, ?):")

		if !scanner.Scan() {
			return ranges.RangeQuery{}, false, false, scanner.Err()
		}

		answer := strings.TrimSpace(scanner.Text())

		if answer == "" {
			fmt.Println("Provide a valid request.")
			continue
		}

		switch strings.ToLower(answer) {
		case "q":
			return ranges.RangeQuery{}, true, false, nil
		case "f":
			return ranges.RangeQuery{}, false, true, nil
		case "?":
			showHelpForRange()
			return ranges.RangeQuery{}, false, false, nil
		}

		parsedRange, err := ranges.ParseRange(answer)

		if err != nil {
			fmt.Println("Provide a valid range.")
			continue
		}

		return parsedRange, false, false, nil
	}
}

func showHelpForRange() {
	const str = "Available commands:\n\n" +
		"f - Finish\n" +
		"q - Quit\n" +
		"? - Show this help\n\n" +
		"You can control selection using the following features:\n\n" +
		"1. Type N to select the corresponding element (N is an element number).\n\n" +
		"   Example: 42 (selects the element 42)\n\n" +
		"2. Type N..M to select elements from N to M inclusive.\n\n" +
		"   Example: 5..10 (selects the elements 5, 6, 7, 8, 9, and 10)\n\n" +
		"3. Type N.. to select N and all elements after.\n\n" +
		"   Example: 5.. (selects the element 5 and all elements after)\n\n" +
		"4. Type ..N to select all elements before N inclusive.\n\n" +
		"   Example: ..5 (selects the elements 1, 2, 3, 4, and 5)\n\n" +
		"5. Type .. to select all elements.\n\n" +
		"6. You can combine multiple expressions separated by a comma.\n\n" +
		"   Example: ..10, 20, 30..40, 50..\n\n" +
		"7. Select the same element twice to deselect it.\n\n"

	fmt.Print(str)
}

func showStatus(session *model.QuizSession, chosenIndices map[int]struct{}) {
	fmt.Println("Questions:")

	for i, item := range session.Items {
		var selected string

		if _, ok := chosenIndices[i]; ok {
			selected = "+"
		} else {
			selected = " "
		}

		question := ""

		for line := range strings.Lines(item.Question) {
			question = strings.TrimSuffix(line, "\n")
			break
		}

		runes := []rune(question)

		if len(runes) > 50 {
			runes = runes[:50]
			question = string(runes) + "..."
		}

		fmt.Printf("%v [%v] %v\n", selected, i, question)
	}
}

func toggleSelection(
	session *model.QuizSession,
	chosenIndices map[int]struct{},
	selectionRange ranges.RangeQuery,
) {
	for _, rangePart := range selectionRange.Parts {
		left := rangePart.OpenIndex

		var right int

		if rangePart.CloseIndex == -1 {
			right = len(session.Items)
		} else {
			right = rangePart.CloseIndex
		}

		for left < right {
			if _, ok := chosenIndices[left]; !ok {
				if left < len(session.Items) {
					chosenIndices[left] = struct{}{}
				}
			} else {
				delete(chosenIndices, left)
			}
			left++
		}
	}
}
