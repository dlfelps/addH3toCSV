package config

import (
	"os"
	"path/filepath"
	"testing"
	"csv-h3-tool/internal/h3"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	
	// Test default values
	if config.InputFile != "" {
		t.Errorf("Expected empty InputFile, got %s", config.InputFile)
	}
	
	if config.OutputFile != "" {
		t.Errorf("Expected empty OutputFile, got %s", config.OutputFile)
	}
	
	if config.LatColumn != "latitude" {
		t.Errorf("Expected LatColumn 'latitude', got %s", config.LatColumn)
	}
	
	if config.LngColumn != "longitude" {
		t.Errorf("Expected LngColumn 'longitude', got %s", config.LngColumn)
	}
	
	if config.Resolution != int(h3.ResolutionStreet) {
		t.Errorf("Expected Resolution %d, got %d", int(h3.ResolutionStreet), config.Resolution)
	}
	
	if !config.HasHeaders {
		t.Error("Expected HasHeaders to be true")
	}
	
	if config.Delimiter != ',' {
		t.Errorf("Expected Delimiter ',', got %c", config.Delimiter)
	}
	
	if config.Overwrite {
		t.Error("Expected Overwrite to be false")
	}
	
	if config.Verbose {
		t.Error("Expected Verbose to be false")
	}
}

func TestConfig_ValidateInputFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	tests := []struct {
		name        string
		inputFile   string
		expectError bool
	}{
		{
			name:        "empty input file",
			inputFile:   "",
			expectError: true,
		},
		{
			name:        "non-existent file",
			inputFile:   "/path/to/nonexistent/file.csv",
			expectError: true,
		},
		{
			name:        "valid existing file",
			inputFile:   tempFile.Name(),
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.InputFile = tt.inputFile
			
			err := config.validateInputFile()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfig_ValidateColumns(t *testing.T) {
	tests := []struct {
		name        string
		latColumn   string
		lngColumn   string
		expectError bool
	}{
		{
			name:        "valid columns",
			latColumn:   "latitude",
			lngColumn:   "longitude",
			expectError: false,
		},
		{
			name:        "empty latitude column",
			latColumn:   "",
			lngColumn:   "longitude",
			expectError: true,
		},
		{
			name:        "empty longitude column",
			latColumn:   "latitude",
			lngColumn:   "",
			expectError: true,
		},
		{
			name:        "same column names",
			latColumn:   "coord",
			lngColumn:   "coord",
			expectError: true,
		},
		{
			name:        "same column names with different case",
			latColumn:   "Latitude",
			lngColumn:   "latitude",
			expectError: true,
		},
		{
			name:        "columns with whitespace",
			latColumn:   " lat ",
			lngColumn:   " lng ",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.LatColumn = tt.latColumn
			config.LngColumn = tt.lngColumn
			
			err := config.validateColumns()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfig_ValidateResolution(t *testing.T) {
	tests := []struct {
		name        string
		resolution  int
		expectError bool
	}{
		{
			name:        "valid resolution 0",
			resolution:  0,
			expectError: false,
		},
		{
			name:        "valid resolution 8",
			resolution:  8,
			expectError: false,
		},
		{
			name:        "valid resolution 15",
			resolution:  15,
			expectError: false,
		},
		{
			name:        "invalid negative resolution",
			resolution:  -1,
			expectError: true,
		},
		{
			name:        "invalid high resolution",
			resolution:  16,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Resolution = tt.resolution
			
			err := config.validateResolution()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfig_ValidateOutputFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create an existing file
	existingFile := filepath.Join(tempDir, "existing.csv")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}
	
	tests := []struct {
		name        string
		outputFile  string
		overwrite   bool
		expectError bool
	}{
		{
			name:        "empty output file generates default",
			outputFile:  "",
			overwrite:   false,
			expectError: false,
		},
		{
			name:        "new file in existing directory",
			outputFile:  filepath.Join(tempDir, "new.csv"),
			overwrite:   false,
			expectError: false,
		},
		{
			name:        "existing file without overwrite",
			outputFile:  existingFile,
			overwrite:   false,
			expectError: true,
		},
		{
			name:        "existing file with overwrite",
			outputFile:  existingFile,
			overwrite:   true,
			expectError: false,
		},
		{
			name:        "file in non-existent directory",
			outputFile:  "/path/to/nonexistent/dir/output.csv",
			overwrite:   false,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.OutputFile = tt.outputFile
			config.Overwrite = tt.overwrite
			config.InputFile = "test.csv" // Set for default output generation
			
			err := config.validateOutputFile()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfig_GenerateDefaultOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		expected   string
	}{
		{
			name:      "empty input file",
			inputFile: "",
			expected:  "output_with_h3.csv",
		},
		{
			name:      "simple filename",
			inputFile: "data.csv",
			expected:  "data_with_h3.csv",
		},
		{
			name:      "filename with path",
			inputFile: filepath.Join("path", "to", "data.csv"),
			expected:  filepath.Join("path", "to", "data_with_h3.csv"),
		},
		{
			name:      "filename without extension",
			inputFile: "data",
			expected:  "data_with_h3",
		},
		{
			name:      "filename with different extension",
			inputFile: "data.txt",
			expected:  "data_with_h3.txt",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.InputFile = tt.inputFile
			
			result := config.fileHandler.GenerateOutputPath(tt.inputFile, "_with_h3")
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConfig_GetResolutionDescription(t *testing.T) {
	tests := []struct {
		resolution int
		expected   string
	}{
		{0, "Country level (~1107.71 km)"},
		{8, "Street level (~461.35 m)"},
		{15, "Page level (~0.51 m)"},
		{99, "Resolution 99"}, // Invalid resolution
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			config := NewConfig()
			config.Resolution = tt.resolution
			
			result := config.GetResolutionDescription()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_validate_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	tests := []struct {
		name        string
		setupConfig func(*Config)
		expectError bool
	}{
		{
			name: "valid configuration",
			setupConfig: func(c *Config) {
				c.InputFile = tempFile.Name()
				c.LatColumn = "lat"
				c.LngColumn = "lng"
				c.Resolution = 8
			},
			expectError: false,
		},
		{
			name: "missing input file",
			setupConfig: func(c *Config) {
				c.InputFile = ""
			},
			expectError: true,
		},
		{
			name: "invalid resolution",
			setupConfig: func(c *Config) {
				c.InputFile = tempFile.Name()
				c.Resolution = -1
			},
			expectError: true,
		},
		{
			name: "same column names",
			setupConfig: func(c *Config) {
				c.InputFile = tempFile.Name()
				c.LatColumn = "coord"
				c.LngColumn = "coord"
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			tt.setupConfig(config)
			
			err := config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	config := NewConfig()
	config.InputFile = "input.csv"
	config.OutputFile = "output.csv"
	config.LatColumn = "lat"
	config.LngColumn = "lng"
	config.Resolution = 8
	
	result := config.String()
	
	// Check that the string contains key information
	expectedSubstrings := []string{
		"input.csv",
		"output.csv",
		"lat",
		"lng",
		"Resolution: 8",
		"Street level",
	}
	
	for _, expected := range expectedSubstrings {
		if !contains(result, expected) {
			t.Errorf("Expected string to contain %s, got: %s", expected, result)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}