package cli

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	
	if cli == nil {
		t.Fatal("Expected CLI instance, got nil")
	}
	
	if cli.config == nil {
		t.Fatal("Expected config to be initialized")
	}
	
	if cli.rootCmd == nil {
		t.Fatal("Expected rootCmd to be initialized")
	}
	
	// Test default configuration values
	if cli.config.LatColumn != "latitude" {
		t.Errorf("Expected default LatColumn 'latitude', got %s", cli.config.LatColumn)
	}
	
	if cli.config.LngColumn != "longitude" {
		t.Errorf("Expected default LngColumn 'longitude', got %s", cli.config.LngColumn)
	}
	
	if cli.config.Resolution != 8 {
		t.Errorf("Expected default Resolution 8, got %d", cli.config.Resolution)
	}
}

func TestCLI_ValidateArgs(t *testing.T) {
	cli := NewCLI()
	
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid single file",
			args:        []string{tempFile.Name()},
			expectError: false,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "multiple arguments",
			args:        []string{"file1.csv", "file2.csv"},
			expectError: true,
		},
		{
			name:        "empty filename",
			args:        []string{""},
			expectError: true,
		},
		{
			name:        "non-existent file",
			args:        []string{"nonexistent.csv"},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.ValidateArgs(tt.args)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestParseResolution(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:        "valid resolution 0",
			input:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "valid resolution 8",
			input:       "8",
			expected:    8,
			expectError: false,
		},
		{
			name:        "valid resolution 15",
			input:       "15",
			expected:    15,
			expectError: false,
		},
		{
			name:        "resolution with whitespace",
			input:       " 10 ",
			expected:    10,
			expectError: false,
		},
		{
			name:        "invalid negative resolution",
			input:       "-1",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid high resolution",
			input:       "16",
			expected:    0,
			expectError: true,
		},
		{
			name:        "non-numeric resolution",
			input:       "abc",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty resolution",
			input:       "",
			expected:    0,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseResolution(tt.input)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseDelimiter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    rune
		expectError bool
	}{
		{
			name:        "comma delimiter",
			input:       ",",
			expected:    ',',
			expectError: false,
		},
		{
			name:        "semicolon delimiter",
			input:       ";",
			expected:    ';',
			expectError: false,
		},
		{
			name:        "tab delimiter",
			input:       "\\t",
			expected:    '\t',
			expectError: false,
		},
		{
			name:        "delimiter with whitespace",
			input:       " | ",
			expected:    '|',
			expectError: false,
		},
		{
			name:        "empty delimiter",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "multiple character delimiter",
			input:       ",,",
			expected:    0,
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expected:    0,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDelimiter(tt.input)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %c, got %c", tt.expected, result)
			}
		})
	}
}

func TestCLI_FlagParsing(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *CLI)
	}{
		{
			name: "basic flags",
			args: []string{tempFile.Name(), "-o", "output.csv", "-r", "10"},
			validate: func(t *testing.T, cli *CLI) {
				if cli.config.OutputFile != "output.csv" {
					t.Errorf("Expected OutputFile 'output.csv', got %s", cli.config.OutputFile)
				}
				if cli.config.Resolution != 10 {
					t.Errorf("Expected Resolution 10, got %d", cli.config.Resolution)
				}
			},
		},
		{
			name: "column flags",
			args: []string{tempFile.Name(), "--lat-column", "lat", "--lng-column", "lng"},
			validate: func(t *testing.T, cli *CLI) {
				if cli.config.LatColumn != "lat" {
					t.Errorf("Expected LatColumn 'lat', got %s", cli.config.LatColumn)
				}
				if cli.config.LngColumn != "lng" {
					t.Errorf("Expected LngColumn 'lng', got %s", cli.config.LngColumn)
				}
			},
		},
		{
			name: "boolean flags",
			args: []string{tempFile.Name(), "--overwrite", "--verbose", "--no-headers"},
			validate: func(t *testing.T, cli *CLI) {
				if !cli.config.Overwrite {
					t.Error("Expected Overwrite to be true")
				}
				if !cli.config.Verbose {
					t.Error("Expected Verbose to be true")
				}
				if cli.config.HasHeaders {
					t.Error("Expected HasHeaders to be false")
				}
			},
		},
		{
			name: "delimiter flag",
			args: []string{tempFile.Name(), "--delimiter", ";"},
			validate: func(t *testing.T, cli *CLI) {
				if cli.config.Delimiter != ';' {
					t.Errorf("Expected Delimiter ';', got %c", cli.config.Delimiter)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			
			// Capture output to avoid printing during tests
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			
			// Set args and execute
			cli.rootCmd.SetArgs(tt.args)
			err := cli.rootCmd.Execute()
			
			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			
			// Read captured output (we don't need it for these tests)
			buf := make([]byte, 1024)
			r.Read(buf)
			r.Close()
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			tt.validate(t, cli)
		})
	}
}

func TestCLI_HelpOutput(t *testing.T) {
	cli := NewCLI()
	cli.AddHelpCommand()
	
	// Test main help
	cli.rootCmd.SetArgs([]string{"--help"})
	
	// Capture output
	var buf bytes.Buffer
	cli.rootCmd.SetOut(&buf)
	
	err := cli.rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := buf.String()
	
	// Check that help contains key information
	expectedSubstrings := []string{
		"csv-h3-tool",
		"H3 geospatial",
		"latitude",
		"longitude",
		"resolution",
		"BASIC USAGE:",
	}
	
	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected help output to contain %s", expected)
		}
	}
}

func TestCLI_ResolutionHelp(t *testing.T) {
	cli := NewCLI()
	cli.AddHelpCommand()
	
	// Test resolution help command by capturing stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	cli.rootCmd.SetArgs([]string{"resolutions"})
	err := cli.rootCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Read captured output
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	r.Close()
	
	// Check that resolution help contains key information
	expectedSubstrings := []string{
		"H3 Resolution Levels and Use Cases",
		"hierarchical hexagonal grid",
		"Country level (~1107.71 km)",
		"Street level (~461.35 m)",
		"DEFAULT",
		"SELECTION GUIDELINES",
		"HIERARCHICAL RELATIONSHIPS",
		"parent-child relationships",
		"EXAMPLES:",
	}
	
	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected resolution help to contain %s", expected)
		}
	}
	
	// Check that all resolution levels 0-15 are present
	for i := 0; i <= 15; i++ {
		if !strings.Contains(output, fmt.Sprintf("%-4d", i)) {
			t.Errorf("Expected resolution help to contain level %d", i)
		}
	}
}

func TestCLI_ExamplesHelp(t *testing.T) {
	cli := NewCLI()
	cli.AddHelpCommand()
	
	// Test examples help command by capturing stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	cli.rootCmd.SetArgs([]string{"examples"})
	err := cli.rootCmd.Execute()
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Read captured output
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	r.Close()
	
	// Check that examples help contains key information
	expectedSubstrings := []string{
		"CSV H3 Tool - Usage Examples",
		"Basic Usage",
		"Custom Output File",
		"Different Resolution",
		"Custom Column Names",
		"No Header Row",
		"Different Delimiter",
		"Tab-Separated Values",
		"Large File Processing",
		"COMMON CSV FORMATS:",
		"Standard format with headers:",
		"Alternative column names:",
		"No headers (use column indices):",
		"European format (semicolon delimiter):",
		"OUTPUT FORMAT:",
		"h3_index",
	}
	
	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected examples help to contain %s", expected)
		}
	}
	
	// Check that practical commands are included
	expectedCommands := []string{
		"csv-h3-tool locations.csv",
		"csv-h3-tool input.csv -o processed_locations.csv",
		"csv-h3-tool properties.csv -r 10",
		"--lat-column",
		"--lng-column",
		"--no-headers",
		"--delimiter",
	}
	
	for _, expected := range expectedCommands {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected examples help to contain command %s", expected)
		}
	}
}

func TestCLI_HelpCompleteness(t *testing.T) {
	cli := NewCLI()
	cli.AddHelpCommand()
	
	// Test main help
	cli.rootCmd.SetArgs([]string{"--help"})
	
	// Capture output
	var buf bytes.Buffer
	cli.rootCmd.SetOut(&buf)
	
	err := cli.rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := buf.String()
	
	// Check comprehensive help content
	expectedSections := []string{
		"H3 is a hierarchical geospatial indexing system",
		"BASIC USAGE:",
		"COLUMN CONFIGURATION:",
		"CSV FORMAT OPTIONS:",
		"ADVANCED USAGE:",
		"RESOLUTION LEVELS:",
		"OUTPUT FORMAT:",
		"Available Commands:",
		"examples",
		"resolutions",
	}
	
	for _, expected := range expectedSections {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected main help to contain section %s", expected)
		}
	}
	
	// Check that all flags are documented with enhanced descriptions
	expectedFlags := []string{
		"Name or index of the latitude column",
		"Name or index of the longitude column", 
		"H3 resolution level (0-15). Higher = more precise",
		"CSV delimiter character. Use '\\t' for tab",
		"Force processing without header row",
		"Overwrite output file if it already exists",
		"Enable verbose output with processing details",
	}
	
	for _, expected := range expectedFlags {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected main help to contain flag description %s", expected)
		}
	}
}

func TestCLI_HelpCommands(t *testing.T) {
	cli := NewCLI()
	cli.AddHelpCommand()
	
	// Test that help commands are properly registered
	commands := cli.rootCmd.Commands()
	
	var foundResolutions, foundExamples bool
	for _, cmd := range commands {
		if cmd.Use == "resolutions" {
			foundResolutions = true
			if cmd.Short != "Show H3 resolution levels and their descriptions" {
				t.Errorf("Unexpected resolutions command description: %s", cmd.Short)
			}
		}
		if cmd.Use == "examples" {
			foundExamples = true
			if cmd.Short != "Show common usage examples and patterns" {
				t.Errorf("Unexpected examples command description: %s", cmd.Short)
			}
		}
	}
	
	if !foundResolutions {
		t.Error("Expected resolutions command to be registered")
	}
	if !foundExamples {
		t.Error("Expected examples command to be registered")
	}
}

func TestCLI_InvalidFlags(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "invalid resolution",
			args: []string{tempFile.Name(), "-r", "20"},
		},
		{
			name: "invalid delimiter",
			args: []string{tempFile.Name(), "--delimiter", ",,"},
		},
		{
			name: "missing input file",
			args: []string{"-o", "output.csv"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			
			// Capture output to avoid printing during tests
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w
			
			cli.rootCmd.SetArgs(tt.args)
			err := cli.rootCmd.Execute()
			
			// Restore stderr
			w.Close()
			os.Stderr = oldStderr
			
			// Read captured output
			buf := make([]byte, 1024)
			r.Read(buf)
			r.Close()
			
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestCLI_GetConfig(t *testing.T) {
	cli := NewCLI()
	
	config := cli.GetConfig()
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	
	if config != cli.config {
		t.Error("Expected GetConfig to return the same config instance")
	}
}