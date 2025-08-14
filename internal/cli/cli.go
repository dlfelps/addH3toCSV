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
	version string
	buildTime string
	gitCommit string
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

H3 is a hierarchical geospatial indexing system developed by Uber that provides
uniform hexagonal grid cells at multiple resolution levels. Each H3 index represents
a unique location on Earth with consistent spatial relationships.

BASIC USAGE:
  csv-h3-tool data.csv                    # Process with default settings
  csv-h3-tool data.csv -o output.csv      # Specify output file
  csv-h3-tool data.csv -r 10              # Use resolution 10 (property level)

COLUMN CONFIGURATION:
  csv-h3-tool data.csv --lat-column lat --lng-column lon
  csv-h3-tool data.csv --lat-column "Latitude" --lng-column "Longitude"

CSV FORMAT OPTIONS:
  csv-h3-tool data.csv --no-headers       # CSV without header row
  csv-h3-tool data.csv --delimiter ";"    # Semicolon-separated values
  csv-h3-tool data.csv --delimiter "\t"   # Tab-separated values

ADVANCED USAGE:
  csv-h3-tool large_dataset.csv -r 8 -v --overwrite
  csv-h3-tool locations.csv --lat-column "lat_deg" --lng-column "lng_deg" -r 12

RESOLUTION LEVELS:
  Use 'csv-h3-tool resolutions' to see all available H3 resolution levels.
  Common choices:
    - Resolution 6 (3.23 km): City block analysis
    - Resolution 8 (461 m): Street-level analysis (default)
    - Resolution 10 (66 m): Property/lot analysis
    - Resolution 12 (9.4 m): Building/room analysis

OUTPUT FORMAT:
  The output CSV will contain all original columns plus a new 'h3_index' column
  with the calculated H3 index values. Invalid coordinates will have empty H3 values.`,
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
		"Name or index of the latitude column (e.g., 'latitude', 'lat', '0')")
	flags.StringVar(&c.config.LngColumn, "lng-column", "longitude", 
		"Name or index of the longitude column (e.g., 'longitude', 'lng', '1')")
	
	// H3 resolution
	flags.IntVarP(&c.config.Resolution, "resolution", "r", int(8), 
		"H3 resolution level (0-15). Higher = more precise. Default: 8 (street level)")
	
	// CSV options
	flags.BoolVar(&c.config.HasHeaders, "headers", true, 
		"CSV file has header row (automatically detected)")
	
	// We'll handle no-headers in PreRunE since it needs to override the default
	
	// Delimiter option (string that gets converted to rune)
	var delimiterStr string
	flags.StringVar(&delimiterStr, "delimiter", ",", 
		"CSV delimiter character. Use '\\t' for tab, ';' for semicolon")
	
	// No-headers flag (handled separately)
	var noHeaders bool
	flags.BoolVar(&noHeaders, "no-headers", false, 
		"Force processing without header row (overrides --headers)")
	
	// File handling
	flags.BoolVar(&c.config.Overwrite, "overwrite", false, 
		"Overwrite output file if it already exists")
	
	// Verbose output
	flags.BoolVarP(&c.config.Verbose, "verbose", "v", false, 
		"Enable verbose output with processing details and error messages")
	
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

// SetVersionInfo sets version information for the CLI
func (c *CLI) SetVersionInfo(version, buildTime, gitCommit string) {
	c.version = version
	c.buildTime = buildTime
	c.gitCommit = gitCommit
	
	// Update the root command with version information
	c.rootCmd.Version = fmt.Sprintf("%s (built %s, commit %s)", version, buildTime, gitCommit)
}

// AddHelpCommand adds additional help commands for H3 resolutions and examples
func (c *CLI) AddHelpCommand() {
	// H3 resolutions help command
	resolutionsCmd := &cobra.Command{
		Use:   "resolutions",
		Short: "Show H3 resolution levels and their descriptions",
		Long:  "Display all available H3 resolution levels with their approximate edge lengths and use cases",
		Run: func(cmd *cobra.Command, args []string) {
			c.printResolutionHelp()
		},
	}
	
	// Examples help command
	examplesCmd := &cobra.Command{
		Use:   "examples",
		Short: "Show common usage examples and patterns",
		Long:  "Display practical examples of how to use the CSV H3 tool with different data formats and scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			c.printExamplesHelp()
		},
	}
	
	c.rootCmd.AddCommand(resolutionsCmd)
	c.rootCmd.AddCommand(examplesCmd)
}

// printResolutionHelp prints detailed information about H3 resolution levels
func (c *CLI) printResolutionHelp() {
	fmt.Println("H3 Resolution Levels and Use Cases")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Println("H3 uses a hierarchical hexagonal grid system where each resolution level")
	fmt.Println("provides increasingly precise spatial indexing. Choose the resolution that")
	fmt.Println("matches your analysis requirements:")
	fmt.Println()
	
	resolutions := []struct {
		level       int
		description string
		useCase     string
		examples    string
	}{
		{0, "Country level (~1107.71 km)", "Continental/country-wide analysis", "Global logistics, climate zones"},
		{1, "State level (~418.68 km)", "State/province-wide analysis", "Regional planning, weather patterns"},
		{2, "Metro level (~158.24 km)", "Metropolitan area analysis", "Urban planning, service areas"},
		{3, "City level (~59.81 km)", "City-wide analysis", "Municipal services, demographics"},
		{4, "District level (~22.61 km)", "District/county analysis", "School districts, postal zones"},
		{5, "Neighborhood level (~8.54 km)", "Neighborhood analysis", "Community planning, local services"},
		{6, "Block level (~3.23 km)", "City block analysis", "Traffic analysis, retail catchment"},
		{7, "Building level (~1.22 km)", "Building cluster analysis", "Campus planning, facility management"},
		{8, "Street level (~461.35 m)", "Street-level analysis (DEFAULT)", "Address geocoding, delivery routes"},
		{9, "Intersection level (~174.38 m)", "Street intersection analysis", "Traffic lights, crosswalk planning"},
		{10, "Property level (~65.91 m)", "Property/lot analysis", "Real estate, land parcels"},
		{11, "Room level (~24.91 m)", "Room-level analysis", "Indoor positioning, floor plans"},
		{12, "Desk level (~9.42 m)", "Desk/workspace analysis", "Office layouts, seating charts"},
		{13, "Chair level (~3.56 m)", "Chair/seat analysis", "Precise indoor positioning"},
		{14, "Book level (~1.35 m)", "Book/object analysis", "Inventory tracking, asset management"},
		{15, "Page level (~0.51 m)", "Page/fine-detail analysis", "High-precision measurements"},
	}
	
	fmt.Printf("%-4s %-32s %-35s %s\n", "Res", "Scale & Edge Length", "Primary Use Case", "Example Applications")
	fmt.Printf("%-4s %-32s %-35s %s\n", "---", "--------------------------------", "-----------------------------------", "--------------------")
	
	for _, res := range resolutions {
		marker := ""
		if res.level == 8 {
			marker = " *"
		}
		fmt.Printf("%-4d %-32s %-35s %s%s\n", res.level, res.description, res.useCase, res.examples, marker)
	}
	
	fmt.Println()
	fmt.Println("SELECTION GUIDELINES:")
	fmt.Println("* Higher resolution = more precise indexing but larger datasets")
	fmt.Println("* Lower resolution = broader spatial grouping, smaller datasets")
	fmt.Println("* Resolution 8 (default) works well for most location-based applications")
	fmt.Println("* Consider your data size and analysis requirements when choosing")
	fmt.Println()
	fmt.Println("HIERARCHICAL RELATIONSHIPS:")
	fmt.Println("H3 indexes have parent-child relationships across resolution levels.")
	fmt.Println("A single resolution 7 cell contains exactly 7 resolution 8 cells.")
	fmt.Println("This enables efficient spatial aggregation and drill-down analysis.")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  csv-h3-tool data.csv -r 6   # City block level analysis")
	fmt.Println("  csv-h3-tool data.csv -r 8   # Street level (default)")
	fmt.Println("  csv-h3-tool data.csv -r 10  # Property/lot level analysis")
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

// printExamplesHelp prints practical usage examples
func (c *CLI) printExamplesHelp() {
	fmt.Println("CSV H3 Tool - Usage Examples")
	fmt.Println("============================")
	fmt.Println()
	
	examples := []struct {
		title       string
		description string
		command     string
		notes       string
	}{
		{
			"Basic Usage",
			"Process a CSV file with default settings",
			"csv-h3-tool locations.csv",
			"Uses default column names 'latitude' and 'longitude', resolution 8",
		},
		{
			"Custom Output File",
			"Specify where to save the results",
			"csv-h3-tool input.csv -o processed_locations.csv",
			"Creates processed_locations.csv with H3 indexes added",
		},
		{
			"Different Resolution",
			"Use property-level precision for real estate data",
			"csv-h3-tool properties.csv -r 10",
			"Resolution 10 provides ~66m precision, good for property analysis",
		},
		{
			"Custom Column Names",
			"Handle CSV files with different column headers",
			"csv-h3-tool data.csv --lat-column \"Lat\" --lng-column \"Long\"",
			"Specify exact column names as they appear in your CSV header",
		},
		{
			"No Header Row",
			"Process CSV files without headers",
			"csv-h3-tool raw_data.csv --no-headers --lat-column \"0\" --lng-column \"1\"",
			"Use column indices (0-based) when there are no headers",
		},
		{
			"Different Delimiter",
			"Handle semicolon-separated or tab-separated files",
			"csv-h3-tool european_data.csv --delimiter \";\"",
			"Common in European CSV files",
		},
		{
			"Tab-Separated Values",
			"Process TSV files",
			"csv-h3-tool data.tsv --delimiter \"\\t\"",
			"Use \\t for tab character",
		},
		{
			"Large File Processing",
			"Process large datasets with verbose output",
			"csv-h3-tool big_dataset.csv -v --overwrite",
			"Verbose mode shows progress and detailed error messages",
		},
		{
			"City-Level Analysis",
			"Group locations by city blocks",
			"csv-h3-tool stores.csv -r 6 -o city_blocks.csv",
			"Resolution 6 (~3.2km) good for city-wide spatial analysis",
		},
		{
			"High-Precision Indoor",
			"Indoor positioning with room-level precision",
			"csv-h3-tool indoor_sensors.csv -r 11 --lat-column \"sensor_lat\" --lng-column \"sensor_lng\"",
			"Resolution 11 (~25m) suitable for building/campus analysis",
		},
	}
	
	for i, example := range examples {
		fmt.Printf("%d. %s\n", i+1, example.title)
		fmt.Printf("   %s\n", example.description)
		fmt.Printf("   Command: %s\n", example.command)
		fmt.Printf("   Notes: %s\n", example.notes)
		fmt.Println()
	}
	
	fmt.Println("COMMON CSV FORMATS:")
	fmt.Println("===================")
	fmt.Println()
	fmt.Println("Standard format with headers:")
	fmt.Println("  latitude,longitude,name,category")
	fmt.Println("  40.7128,-74.0060,New York,City")
	fmt.Println("  34.0522,-118.2437,Los Angeles,City")
	fmt.Println()
	fmt.Println("Alternative column names:")
	fmt.Println("  lat,lng,location_name")
	fmt.Println("  40.7128,-74.0060,NYC")
	fmt.Println("  Command: csv-h3-tool data.csv --lat-column lat --lng-column lng")
	fmt.Println()
	fmt.Println("No headers (use column indices):")
	fmt.Println("  40.7128,-74.0060,New York")
	fmt.Println("  34.0522,-118.2437,Los Angeles")
	fmt.Println("  Command: csv-h3-tool data.csv --no-headers --lat-column 0 --lng-column 1")
	fmt.Println()
	fmt.Println("European format (semicolon delimiter):")
	fmt.Println("  latitude;longitude;city")
	fmt.Println("  48,8567;2,3508;Paris")
	fmt.Println("  Command: csv-h3-tool data.csv --delimiter \";\"")
	fmt.Println()
	fmt.Println("OUTPUT FORMAT:")
	fmt.Println("==============")
	fmt.Println("The tool adds an 'h3_index' column to your original data:")
	fmt.Println("  latitude,longitude,name,category,h3_index")
	fmt.Println("  40.7128,-74.0060,New York,City,882a100d2ffffff")
	fmt.Println("  34.0522,-118.2437,Los Angeles,City,882ad0682ffffff")
	fmt.Println()
	fmt.Println("For more information about H3 resolution levels:")
	fmt.Println("  csv-h3-tool resolutions")
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