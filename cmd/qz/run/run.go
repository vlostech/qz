package run

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vlostech/qz/internal/model"
	"github.com/vlostech/qz/internal/ranges"
	"github.com/vlostech/qz/internal/session"
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
	rangeQuery, err := ranges.ParseRange(rangeString)

	if err != nil {
		return err
	}

	s, err := session.CreateSession(filePath, count, rangeQuery)

	if err != nil {
		return err
	}

	runFirstPhase(s)
	runSecondPhase(s)

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
