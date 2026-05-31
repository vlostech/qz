package run

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vlostech/qz/internal/adapter"
	"github.com/vlostech/qz/internal/app"
	"github.com/vlostech/qz/internal/domain"
	"github.com/vlostech/qz/internal/helper/ranges"
)

type Runner struct {
	useCase *app.UseCase
}

func NewRunner(useCase *app.UseCase) *Runner {
	return &Runner{
		useCase: useCase,
	}
}

func (r *Runner) Run(args Args) error {
	rangeQuery, err := ranges.Parse(args.RangeString)

	if err != nil {
		return err
	}

	input := app.CreateSessionInput{
		FilePath:   args.FilePath,
		Count:      args.QuestionCount,
		RangeQuery: rangeQuery,
	}

	output, err := r.useCase.CreateSession(input)

	if err != nil {
		return err
	}

	session := output.Session

	state := &sessionState{
		questionIdx:   0,
		quizSession:   session,
		chosenIndices: make(map[int]struct{}, len(session.SessionItems)),
	}

	err = r.runSession(state)

	return err
}

func (r *Runner) runSession(state *sessionState) error {
	var result stepResult
	var err error

	result = stepResult{
		nextFunc:        r.runPhaseWithQuestions,
		isInputRequired: false,
	}

	var inputString string

	for {
		if result.nextFunc == nil {
			return nil
		}

		var args stepArgs

		if result.isInputRequired {
			inputString, err = r.getMultilineString()

			if err != nil {
				return err
			}

			args.input = inputString
		}

		result, err = result.nextFunc(state, args)

		if err != nil {
			return err
		}
	}
}

func (r *Runner) runPhaseWithQuestions(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println("PHASE 1 - QUESTIONS")
	fmt.Println()

	return stepResult{
		nextFunc:        r.showNextQuestion,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showNextQuestion(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Printf("%v/%v\n", state.questionIdx+1, len(state.quizSession.SessionItems))
	fmt.Println()
	fmt.Println(state.quizSession.SessionItems[state.questionIdx].CollectionItem.Question)
	fmt.Println()

	// Show current answer if exists.
	if !r.isStringEmpty(state.quizSession.SessionItems[state.questionIdx].ActualAnswer) {
		fmt.Println("Current answer:")
		fmt.Println(state.quizSession.SessionItems[state.questionIdx].ActualAnswer)
		fmt.Println()
	}

	fmt.Println("Write your answer or enter a command (/g, /f, /q, /?):")

	return stepResult{
		nextFunc:        r.handleAnswerToQuestion,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleAnswerToQuestion(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println()

	if strings.HasPrefix(args.input, "/") {
		commandWithArgs, err := r.getSubcommandArgs(args.input)

		if err != nil {
			fmt.Println("ERROR: Provide a valid command.")
			fmt.Println()

			return stepResult{
				nextFunc:        r.showNextQuestion,
				isInputRequired: false,
			}, nil
		}

		switch commandWithArgs[0] {
		case "/g", "/go":
			if len(commandWithArgs) == 1 {
				return stepResult{
					nextFunc:        r.showQuestionsToNavigate,
					isInputRequired: false,
				}, nil
			}

			var questionNumber int
			questionNumber, err = strconv.Atoi(commandWithArgs[1])

			if err != nil {
				fmt.Println("ERROR: /go command supports only integer arguments.")
				fmt.Println()

				return stepResult{
					nextFunc:        r.showNextQuestion,
					isInputRequired: false,
				}, nil
			}

			state.questionIdx = questionNumber - 1

			return stepResult{
				nextFunc:        r.showNextQuestion,
				isInputRequired: false,
			}, nil

		case "/f", "/finish":
			state.questionIdx = 0

			return stepResult{
				nextFunc:        r.runPhaseWithAnswers,
				isInputRequired: false,
			}, nil

		case "/q", "/quit":
			return stepResult{
				nextFunc:        nil,
				isInputRequired: false,
			}, nil

		case "/?", "/help":
			return stepResult{
				nextFunc:        r.showHelpForAnsweringToQuestion,
				isInputRequired: false,
			}, nil

		default:
			fmt.Println("ERROR: Provide a valid command.")
			fmt.Println()
			return stepResult{
				nextFunc:        r.showNextQuestion,
				isInputRequired: false,
			}, nil
		}
	}

	state.quizSession.SessionItems[state.questionIdx].ActualAnswer = args.input

	// When the current question is answered, we should move to the next
	// unanswered question. First, we should show the next unanswered question
	// after the current position. If there is no such question, we should show
	// the next unanswered question before the current position.
	//
	// Example 1:
	// [ ?, ?, ? ] -> [ ?, "foo", ? ]
	//      ^                     ^
	//
	// Example 2:
	// [ ?, ?, ?, "bar" ] -> [ ?, ?, "foo", "bar" ]
	//         ^               ^

	// Search for unanswered question after the current question.
	for i := state.questionIdx + 1; i < len(state.quizSession.SessionItems); i++ {
		if r.isStringEmpty(state.quizSession.SessionItems[i].ActualAnswer) {
			state.questionIdx = i
			return stepResult{
				nextFunc:        r.showNextQuestion,
				isInputRequired: false,
			}, nil
		}
	}

	// Search for unanswered question before the current question.
	for i := 0; i < state.questionIdx; i++ {
		if r.isStringEmpty(state.quizSession.SessionItems[i].ActualAnswer) {
			state.questionIdx = i
			return stepResult{
				nextFunc:        r.showNextQuestion,
				isInputRequired: false,
			}, nil
		}
	}

	return stepResult{
		nextFunc:        r.runPhaseWithAnswers,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showQuestionsToNavigate(state *sessionState, args stepArgs) (stepResult, error) {
	builder := strings.Builder{}

	for i, item := range state.quizSession.SessionItems {
		answeredMark := " "

		if item.ActualAnswer != "" {
			answeredMark = "+"
		}

		question := r.getTrimmedString(item.CollectionItem.Question, 50)
		builder.WriteString(fmt.Sprintf("%v [%v] %v\n", answeredMark, i+1, question))
	}

	fmt.Println(builder.String())
	fmt.Println("Enter a question number:")

	return stepResult{
		nextFunc:        r.handleProvidedQuestionToNavigate,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleProvidedQuestionToNavigate(state *sessionState, args stepArgs) (stepResult, error) {
	idx, err := strconv.Atoi(args.input)

	fmt.Println()

	if err != nil {
		fmt.Println("ERROR: Provide a valid question number.")
		fmt.Println()
		fmt.Println("Enter a question number:")
		return stepResult{
			nextFunc:        r.handleProvidedQuestionToNavigate,
			isInputRequired: true,
		}, nil
	}

	if idx > len(state.quizSession.SessionItems) {
		fmt.Println("ERROR: Question number out of range.")
		fmt.Println()
		fmt.Println("Enter a question number:")
		return stepResult{
			nextFunc:        r.handleProvidedQuestionToNavigate,
			isInputRequired: true,
		}, nil
	}

	state.questionIdx = idx - 1

	return stepResult{
		nextFunc:        r.showNextQuestion,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showHelpForAnsweringToQuestion(state *sessionState, args stepArgs) (stepResult, error) {
	const str = "Available commands:\n\n" +
		"/go, /g       Go to question choosing it from the list.\n" +
		"/go, /g {num} Go to question with the specified number 'num'.\n" +
		"/finish, /f   Finish the current phase and go to answers.\n" +
		"/quit, /q     Quit the program.\n" +
		"/help, /?     Show this help."

	fmt.Println(str)
	fmt.Println()

	return stepResult{
		nextFunc:        r.showNextQuestion,
		isInputRequired: false,
	}, nil
}

func (r *Runner) runPhaseWithAnswers(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println("PHASE 2 - ANSWERS")
	fmt.Println()

	state.questionIdx = 0

	return stepResult{
		nextFunc:        r.showNextAnswer,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showNextAnswer(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Printf("%v/%v\n", state.questionIdx+1, len(state.quizSession.SessionItems))
	fmt.Println()
	fmt.Println("Question:")
	fmt.Println(state.quizSession.SessionItems[state.questionIdx].CollectionItem.Question)
	fmt.Println()
	fmt.Println("Expected answer:")
	fmt.Println(state.quizSession.SessionItems[state.questionIdx].CollectionItem.Answer)
	fmt.Println()
	fmt.Println("Actual answer:")
	fmt.Println(state.quizSession.SessionItems[state.questionIdx].ActualAnswer)
	fmt.Println()
	fmt.Println("Press Enter to continue...")

	return stepResult{
		nextFunc:        r.handleShownAnswer,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleShownAnswer(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println()

	state.questionIdx++

	if state.questionIdx < len(state.quizSession.SessionItems) {
		return stepResult{
			nextFunc:        r.showNextAnswer,
			isInputRequired: false,
		}, nil
	}

	return stepResult{
		nextFunc:        r.askIfQuestionsShouldBeSavedToFile,
		isInputRequired: false,
	}, nil
}

func (r *Runner) askIfQuestionsShouldBeSavedToFile(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println("Would you like to save questions to a file (y/n)?")

	return stepResult{
		nextFunc:        r.handleAnswerAboutSavingFile,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleAnswerAboutSavingFile(state *sessionState, args stepArgs) (stepResult, error) {
	answer := strings.TrimSpace(args.input)

	switch strings.ToLower(answer) {
	case "y":
		fmt.Println()

		return stepResult{
			nextFunc:        r.askForQuestionsForSaving,
			isInputRequired: false,
		}, nil
	case "n":
		return stepResult{
			nextFunc:        nil,
			isInputRequired: false,
		}, nil
	}

	fmt.Println("ERROR: Provide a valid answer.")

	return stepResult{
		nextFunc:        r.askIfQuestionsShouldBeSavedToFile,
		isInputRequired: false,
	}, nil
}

func (r *Runner) askForQuestionsForSaving(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println("Questions:")
	fmt.Println()

	questions := r.showChosenQuestions(state)

	fmt.Println(questions)
	fmt.Println("Choose questions or enter command (f, q, ?):")

	return stepResult{
		nextFunc:        r.handleInputWithChosenQuestions,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleInputWithChosenQuestions(state *sessionState, args stepArgs) (stepResult, error) {
	answer := strings.TrimSpace(args.input)

	if answer == "" {
		fmt.Println("ERROR: Provide a valid request.")

		return stepResult{
			nextFunc:        r.askForQuestionsForSaving,
			isInputRequired: false,
		}, nil
	}

	switch strings.ToLower(answer) {
	case "q":
		return stepResult{
			nextFunc:        nil,
			isInputRequired: false,
		}, nil
	case "f":
		return stepResult{
			nextFunc:        r.askForFilePath,
			isInputRequired: false,
		}, nil
	case "?":
		return stepResult{
			nextFunc:        r.showHelpForRangeInput,
			isInputRequired: false,
		}, nil
	}

	parsedRange, err := ranges.Parse(answer)

	if err != nil {
		fmt.Println("ERROR: Provide a valid range.")

		return stepResult{
			nextFunc:        r.askForQuestionsForSaving,
			isInputRequired: false,
		}, nil
	}

	r.toggleSelectedQuestions(state, parsedRange)

	return stepResult{
		nextFunc:        r.askForQuestionsForSaving,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showHelpForRangeInput(state *sessionState, args stepArgs) (stepResult, error) {
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

	return stepResult{
		nextFunc:        r.handleInputWithChosenQuestions,
		isInputRequired: true,
	}, nil
}

func (r *Runner) askForFilePath(state *sessionState, args stepArgs) (stepResult, error) {
	fmt.Println("Enter a file path:")

	return stepResult{
		nextFunc:        r.handleInputWithFilePath,
		isInputRequired: true,
	}, nil
}

func (r *Runner) handleInputWithFilePath(state *sessionState, args stepArgs) (stepResult, error) {
	path := strings.TrimSpace(args.input)

	if path == "" {
		fmt.Println("ERROR: Provide a valid file path.")

		return stepResult{
			nextFunc:        r.askForFilePath,
			isInputRequired: false,
		}, nil
	}

	items := make([]domain.CollectionItem, 0, len(state.chosenIndices))

	for idx := range state.chosenIndices {
		items = append(items, state.quizSession.SessionItems[idx].CollectionItem)
	}

	saveQuestionsInput := app.SaveQuestionsInput{
		FilePath: path,
		Items:    items,
	}

	output, err := r.useCase.SaveQuestions(saveQuestionsInput)

	if err != nil {
		if errors.Is(err, adapter.ErrInvalidFilePath) {
			fmt.Println("ERROR: Provide a valid file path.")

			return stepResult{
				nextFunc:        r.askForFilePath,
				isInputRequired: false,
			}, nil
		}

		return stepResult{}, err
	}

	fmt.Println()
	fmt.Println("Questions are saved in the following file:")
	fmt.Println(output.AbsoluteFilePath)

	return stepResult{
		nextFunc:        nil,
		isInputRequired: false,
	}, nil
}

func (r *Runner) showChosenQuestions(state *sessionState) string {
	builder := strings.Builder{}

	for i, item := range state.quizSession.SessionItems {
		var selected string

		if _, ok := state.chosenIndices[i]; ok {
			selected = "+"
		} else {
			selected = " "
		}

		question := r.getTrimmedString(item.CollectionItem.Question, 50)

		builder.WriteString(fmt.Sprintf("%v [%v] %v\n", selected, i, question))
	}

	return builder.String()
}

func (r *Runner) toggleSelectedQuestions(state *sessionState, selectionRange domain.RangeGroup) {
	for _, rangePart := range selectionRange.Ranges {
		left := rangePart.FirstIndex

		var right int

		if rangePart.LastIndex == -1 {
			right = len(state.quizSession.SessionItems)
		} else {
			right = rangePart.LastIndex
		}

		for left <= right {
			if _, ok := state.chosenIndices[left]; !ok {
				if left < len(state.quizSession.SessionItems) {
					state.chosenIndices[left] = struct{}{}
				}
			} else {
				delete(state.chosenIndices, left)
			}
			left++
		}
	}
}

func (r *Runner) getMultilineString() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	var strList []string

	for {
		if !scanner.Scan() {
			return "", scanner.Err()
		}

		line := scanner.Text()

		if line == "\\end" {
			break
		}

		line, hasSeparator := strings.CutSuffix(line, "\\")

		strList = append(strList, line)

		if !hasSeparator {
			break
		}
	}

	return strings.Join(strList, "\n"), nil
}

// getSubcommandArgs extracts arguments from input. Supports quoted arguments.
//
//	foo bar baz -> [ "foo", "bar", "baz" ]
//	foo "bar baz" -> [ "foo", "bar baz" ]
func (r *Runner) getSubcommandArgs(input string) ([]string, error) {
	// foo bar fiz baz -> [ "foo", "bar", "fiz", "baz" ]
	// "foo bar" fiz baz -> [ "foo bar", "fiz", "baz" ]
	// foo    bar "fiz    baz" -> [ "foo", "bar", "fiz    baz" ]
	// "foo bar fiz baz -> error

	var args []string

	builder := strings.Builder{}

	isArgumentActive := false
	isStringArgument := false

	for _, char := range input {
		if char == ' ' {
			if isArgumentActive {
				if isStringArgument {
					builder.WriteByte(' ')
					continue
				} else {
					args = append(args, builder.String())
					builder.Reset()
					isArgumentActive = false
					continue
				}
			} else {
				continue
			}
		}

		if char == '"' {
			if isArgumentActive {
				if isStringArgument {
					args = append(args, builder.String())
					builder.Reset()
					isArgumentActive = false
					isStringArgument = false
					continue
				} else {
					return nil, errors.New("syntax error")
				}
			} else {
				isArgumentActive = true
				isStringArgument = true
				continue
			}
		}

		if !isArgumentActive {
			isArgumentActive = true
		}

		builder.WriteRune(char)
	}

	if isArgumentActive {
		if isStringArgument {
			return nil, errors.New("syntax error")
		}

		args = append(args, builder.String())
	}

	return args, nil
}

// isStringEmpty determines if the given string is empty or consists only of
// whitespace characters.
func (r *Runner) isStringEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// getTrimmedString trims the input string to the specified maxLen. Appends
// "..." to the end of the string if it exceeds the max length. The length of
// the output string does not exceed maxLen.
//
// Note that "..." requires 3 characters. If an input string is less or equal
// 3, the string will be trimmed without an ellipsis.
func (r *Runner) getTrimmedString(str string, maxLen int) string {
	var output string

	for line := range strings.Lines(str) {
		output = strings.TrimSuffix(line, "\n")
		break
	}

	var suffix string

	if maxLen <= 3 {
		suffix = ""
	} else {
		suffix = "..."
	}

	characters := []rune(output)

	if len(characters) > maxLen {
		characters = characters[:maxLen-len(suffix)]
		output = string(characters) + suffix
	}

	return output
}
