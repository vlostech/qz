package run

import (
	"github.com/spf13/cobra"
	"github.com/vlostech/qz/internal/adapter"
	"github.com/vlostech/qz/internal/app"
	"github.com/vlostech/qz/internal/service"
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
	storage := adapter.NewFileStorage()
	randomizationService := service.NewRandomizationService()
	useCase := app.NewUseCase(storage, randomizationService)
	runner := NewRunner(useCase)

	args := Args{
		RangeString:   rangeString,
		FilePath:      filePath,
		QuestionCount: count,
	}

	return runner.Run(args)
}
