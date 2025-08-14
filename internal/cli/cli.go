package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// CLI represents the command line interface
type CLI struct {
	config *config.Config
	rootCmd *cobra.Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{
		config: config.NewConfig(),
	}
	
	cli.rootCmd = &cobra.Command{
		Use:   "csv-h3-tool [input-file]",
		Short: "Add H3 geospatial indexes to CSV files with latitude/longitude coordinates",
		Long: `CSV H3 Tool processes CSV files containing latitude and longitude coordinates
and adds H3 index values as a new column using Uber's H3 geospatial indexing system.

The tool reads the input CSV, calculates H3 indexes for each coordinate pair using
the specified resolution level, and outputs a new CSV file with all original columns
plus a new H3 index column.

Examples:
  csv-h3-tool data.csv
  csv-h3-tool data.csv -o output.csv -r 10
  csv-h3-tool data.csv --lat-column lat --lng-column lon --resolution 8
  csv-h3-tool data.csv --no-headers --delimiter ";"`,
		Args: cobra.ExactArgs(1),
		RunE: cli.run,
	}
	
	cli.setupFlags()
	return cli
}

// setupFlags configures all command line flags
func (c *CLI) setupFlags() {
	flags := c.rootCmd.Flags()
	
	// Output file
	flags.StringVarP(&c.config.OutputFile, "output", "o", "", 
		"Output CSV file path (default: input_with_h3.csv)")
	
	// Column configuration
	flags.StringVar(&c.config.LatColumn, "lat-column", "latitude", 
		"Name of the latitude column")
	flags.StringVar(&c.config.LngColumn, "lng-column", "longitude", 
		"Name of the longitude column")
	
	// H3 resolution
	flags.IntVarP(&c.config.Resolution, "resolution", "r", int(8), 
		"H3 resolution level (0-15, default: 8 - street level)")
	
	// CSV options
	flags.BoolVar(&c.config.HasHeaders, "headers", true, 
		"CSV file has header row")
	
	// We'll handle no-headers in PreRunE since it needs to override the default
	
	// Delimiter option (string that gets converted to rune)
	var delimiterStr string
	flags.StringVar(&delimiterStr, "delimiter", ",", 
		"CSV delimiter character (default: comma)")
	
	// No-headers flag (handled separately)
	var noHeaders bool
	flags.BoolVar(&noHeaders, "no-headers", false, 
		"CSV file does not have header row")
	
	// File handling
	flags.BoolVar(&c.config.Overwrite, "overwrite", false, 
		"Overwrite output file if it exists")
	
	// Verbose output
	flags.BoolVarP(&c.config.Verbose, "verbose", "v", false, 
		"Enable verbose output")
	
	// Custom flag processing for delimiter and no-headers
	c.rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Handle delimiter conversion
		if delimiterStr != "" {
			if len(delimiterStr) != 1 {
				return fmt.Errorf("delimiter must be a single character, got: %s", delimiterStr)
			}
			c.config.Delimiter = rune(delimiterStr[0])
		}
		
		// Handle no-headers flag
		if cmd.Flags().Changed("no-headers") && noHeaders {
			c.config.HasHeaders = false
		}
		
		return nil
	}
}

// run executes the main command logic
func (c *CLI) run(cmd *cobra.Command, args []string) error {
	// Set input file from positional argument
	c.config.InputFile = args[0]
	
	// Validate configuration
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Print configuration if verbose
	if c.config.Verbose {
		fmt.Printf("Configuration: %s\n", c.config.String())
		fmt.Printf("H3 Resolution: %s\n", c.config.GetResolutionDescription())
	}
	
	// Process the file using the orchestrator
	return c.processFile()
}

// Execute runs the CLI application
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

// GetConfig returns the current configuration
func (c *CLI) GetConfig() *config.Config {
	return c.config
}

// AddHelpCommand adds additional help commands for H3 resolutions
func (c *CLI) AddHelpCommand() {
	helpCmd := &cobra.Command{
		Use:   "resolutions",
		Short: "Show H3 resolution levels and their descriptions",
		Long:  "Display all available H3 resolution levels with their approximate edge lengths and use cases",
		Run: func(cmd *cobra.Command, args []string) {
			c.printResolutionHelp()
		},
	}
	
	c.rootCmd.AddCommand(helpCmd)
}

// printResolutionHelp prints detailed information about H3 resolution levels
func (c *CLI) printResolutionHelp() {
	fmt.Println("H3 Resolution Levels:")
	fmt.Println("====================")
	fmt.Println()
	
	resolutions := []struct {
		level       int
		description string
		useCase     string
	}{
		{0, "Country level (~1107.71 km)", "Continental/country-wide analysis"},
		{1, "State level (~418.68 km)", "State/province-wide analysis"},
		{2, "Metro level (~158.24 km)", "Metropolitan area analysis"},
		{3, "City level (~59.81 km)", "City-wide analysis"},
		{4, "District level (~22.61 km)", "District/county analysis"},
		{5, "Neighborhood level (~8.54 km)", "Neighborhood analysis"},
		{6, "Block level (~3.23 km)", "City block analysis"},
		{7, "Building level (~1.22 km)", "Building cluster analysis"},
		{8, "Street level (~461.35 m)", "Street-level analysis (default)"},
		{9, "Intersection level (~174.38 m)", "Street intersection analysis"},
		{10, "Property level (~65.91 m)", "Property/lot analysis"},
		{11, "Room level (~24.91 m)", "Room-level analysis"},
		{12, "Desk level (~9.42 m)", "Desk/workspace analysis"},
		{13, "Chair level (~3.56 m)", "Chair/seat analysis"},
		{14, "Book level (~1.35 m)", "Book/object analysis"},
		{15, "Page level (~0.51 m)", "Page/fine-detail analysis"},
	}
	
	for _, res := range resolutions {
		fmt.Printf("  %2d: %-30s %s\n", res.level, res.description, res.useCase)
	}
	
	fmt.Println()
	fmt.Println("Higher resolution levels provide more precise spatial indexing but")
	fmt.Println("result in more unique indexes and larger datasets.")
	fmt.Println()
	fmt.Println("Default resolution (8) provides street-level precision suitable for")
	fmt.Println("most location-based applications.")
}

// ValidateArgs validates command line arguments before execution
func (c *CLI) ValidateArgs(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one input file must be specified")
	}
	
	inputFile := args[0]
	if inputFile == "" {
		return fmt.Errorf("input file cannot be empty")
	}
	
	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}
	
	return nil
}

// ParseResolution parses and validates a resolution string
func ParseResolution(resStr string) (int, error) {
	res, err := strconv.Atoi(strings.TrimSpace(resStr))
	if err != nil {
		return 0, fmt.Errorf("invalid resolution format: %s", resStr)
	}
	
	if res < 0 || res > 15 {
		return 0, fmt.Errorf("resolution %d is out of valid range [0, 15]", res)
	}
	
	return res, nil
}

// ParseDelimiter parses and validates a delimiter string
func ParseDelimiter(delimStr string) (rune, error) {
	delimStr = strings.TrimSpace(delimStr)
	if delimStr == "" {
		return 0, fmt.Errorf("delimiter cannot be empty")
	}
	
	// Handle special escape sequences
	switch delimStr {
	case "\\t":
		return '\t', nil
	case "\\n":
		return '\n', nil
	case "\\r":
		return '\r', nil
	}
	
	if len(delimStr) != 1 {
		return 0, fmt.Errorf("delimiter must be a single character, got: %s", delimStr)
	}
	
	return rune(delimStr[0]), nil
}

// processFile processes the CSV file using the orchestrator
func (c *CLI) processFile() error {
	// Create orchestrator with the configuration
	orchestrator := service.NewOrchestrator(c.config)

	// Validate all components are properly wired
	if err := orchestrator.ValidateComponents(); err != nil {
		return fmt.Errorf("component validation failed: %w", err)
	}

	// Process the file
	result, err := orchestrator.ProcessFile()
	if err != nil {
		return fmt.Errorf("file processing failed: %w", err)
	}

	// Display results
	fmt.Printf("Processing completed successfully!\n")
	fmt.Printf("Output file: %s\n", result.OutputFile)
	fmt.Printf("Total records: %d\n", result.TotalRecords)
	fmt.Printf("Valid records: %d\n", result.ValidRecords)
	fmt.Printf("Invalid records: %d\n", result.InvalidRecords)
	fmt.Printf("Processing time: %v\n", result.ProcessingTime)

	if result.InvalidRecords > 0 {
		fmt.Printf("\nWarning: %d records were skipped due to invalid coordinates.\n", result.InvalidRecords)
		fmt.Printf("Use --verbose flag to see detailed error messages.\n")
	}

	return nil
}