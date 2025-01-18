package run

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vlostech/qz/internal/model"
	"github.com/vlostech/qz/internal/random"
	"github.com/vlostech/qz/internal/ranges"
	"github.com/vlostech/qz/internal/storage"
	"os"
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
	items, err := storage.GetQuizItems(filePath)

	if err != nil {
		return err
	}

	var indexes []int

	if rangeString != "" {
		indexes, err = ranges.ParseRange(rangeString, len(items))

		if err != nil {
			return err
		}
	} else {
		indexes = make([]int, len(items))

		for i := range len(items) {
			indexes[i] = i
		}
	}

	if count <= 0 {
		count = len(items)
	}

	randomIndexes := random.Randomize(indexes, count)
	randomItems := make([]model.QuizSessionItem, len(randomIndexes))

	for i, idx := range randomIndexes {
		randomItems[i] = items[idx]
	}

	session := model.QuizSession{
		Items: randomItems,
	}

	runFirstPhase(&session)
	runSecondPhase(&session)

	return nil
}

func runFirstPhase(session *model.QuizSession) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("PHASE 1 - QUESTIONS")
	fmt.Println()

	for i := 0; i < len(session.Items); i++ {
		fmt.Printf("%v/%v\n", i+1, len(session.Items))
		fmt.Println()
		fmt.Println(session.Items[i].Question)
		fmt.Println()
		fmt.Println("Write your answer:")

		scanner.Scan()
		session.Items[i].ActualAnswer = scanner.Text()

		fmt.Println()
	}
}

func runSecondPhase(session *model.QuizSession) {
	fmt.Println("PHASE 2 - ANSWERS")
	fmt.Println()

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
		_, _ = fmt.Scanln()
	}
}
