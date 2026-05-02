package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vlostech/qz/internal/cli/run"
)

var rootCmd = &cobra.Command{
	Use:   "qz",
	Short: "Root command",
}

func init() {
	rootCmd.AddCommand(run.Command)
}

func main() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}
