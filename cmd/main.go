package main

import (
	"fmt"
	"os"

	"csv-h3-tool/internal/cli"
)

func main() {
	// Create CLI instance
	cliApp := cli.NewCLI()
	cliApp.AddHelpCommand()

	// Execute the CLI application
	if err := cliApp.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}