package main

import (
	"fmt"
	"os"

	"csv-h3-tool/internal/cli"
)

// Version information (set by build flags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Create CLI instance with version information
	cliApp := cli.NewCLI()
	cliApp.SetVersionInfo(Version, BuildTime, GitCommit)
	cliApp.AddHelpCommand()

	// Execute the CLI application
	if err := cliApp.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}