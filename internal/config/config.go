package config

import (
	"fmt"
	"strings"
	"csv-h3-tool/internal/h3"
	"csv-h3-tool/internal/filehandler"
)

// Config holds all configuration options for the CSV H3 tool
type Config struct {
	// File paths
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
	
	// CSV column configuration
	LatColumn string `json:"lat_column"`
	LngColumn string `json:"lng_column"`
	
	// H3 configuration
	Resolution int `json:"resolution"`
	
	// CSV processing options
	HasHeaders bool `json:"has_headers"`
	Delimiter  rune `json:"delimiter"`
	
	// File handling options
	Overwrite bool `json:"overwrite"`
	
	// Output options
	Verbose bool `json:"verbose"`
	
	// Internal file handler
	fileHandler *filehandler.FileHandler
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		InputFile:   "",
		OutputFile:  "",
		LatColumn:   "latitude",
		LngColumn:   "longitude",
		Resolution:  int(h3.ResolutionStreet), // Default to street level (8)
		HasHeaders:  true,
		Delimiter:   ',',
		Overwrite:   false,
		Verbose:     false,
		fileHandler: filehandler.NewFileHandler(),
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	// Validate input file
	if c.InputFile == "" {
		return fmt.Errorf("input file path is required")
	}
	
	if err := c.validateInputFile(); err != nil {
		return fmt.Errorf("input file validation failed: %w", err)
	}
	
	// Validate column names
	if err := c.validateColumns(); err != nil {
		return fmt.Errorf("column validation failed: %w", err)
	}
	
	// Validate H3 resolution
	if err := c.validateResolution(); err != nil {
		return fmt.Errorf("resolution validation failed: %w", err)
	}
	
	// Validate output file
	if err := c.validateOutputFile(); err != nil {
		return fmt.Errorf("output file validation failed: %w", err)
	}
	
	return nil
}

// validateInputFile checks if the input file exists and is readable
func (c *Config) validateInputFile() error {
	return c.fileHandler.ValidateInputFile(c.InputFile)
}

// validateColumns validates the column configuration
func (c *Config) validateColumns() error {
	if c.LatColumn == "" {
		return fmt.Errorf("latitude column name cannot be empty")
	}
	
	if c.LngColumn == "" {
		return fmt.Errorf("longitude column name cannot be empty")
	}
	
	// Check for common column name patterns
	latColumn := strings.ToLower(strings.TrimSpace(c.LatColumn))
	lngColumn := strings.ToLower(strings.TrimSpace(c.LngColumn))
	
	if latColumn == lngColumn {
		return fmt.Errorf("latitude and longitude columns cannot be the same: %s", c.LatColumn)
	}
	
	return nil
}

// validateResolution validates the H3 resolution level
func (c *Config) validateResolution() error {
	if c.Resolution < 0 || c.Resolution > 15 {
		return fmt.Errorf("H3 resolution %d is out of valid range [0, 15]", c.Resolution)
	}
	return nil
}

// validateOutputFile validates the output file configuration
func (c *Config) validateOutputFile() error {
	// If no output file specified, generate default name
	if c.OutputFile == "" {
		c.OutputFile = c.fileHandler.GenerateOutputPath(c.InputFile, "_with_h3")
	}
	
	return c.fileHandler.ValidateOutputFile(c.OutputFile, c.Overwrite)
}



// GetResolutionDescription returns a human-readable description of the H3 resolution
func (c *Config) GetResolutionDescription() string {
	descriptions := map[int]string{
		0:  "Country level (~1107.71 km)",
		1:  "State level (~418.68 km)",
		2:  "Metro level (~158.24 km)",
		3:  "City level (~59.81 km)",
		4:  "District level (~22.61 km)",
		5:  "Neighborhood level (~8.54 km)",
		6:  "Block level (~3.23 km)",
		7:  "Building level (~1.22 km)",
		8:  "Street level (~461.35 m)",
		9:  "Intersection level (~174.38 m)",
		10: "Property level (~65.91 m)",
		11: "Room level (~24.91 m)",
		12: "Desk level (~9.42 m)",
		13: "Chair level (~3.56 m)",
		14: "Book level (~1.35 m)",
		15: "Page level (~0.51 m)",
	}
	
	if desc, exists := descriptions[c.Resolution]; exists {
		return desc
	}
	return fmt.Sprintf("Resolution %d", c.Resolution)
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf("Config{InputFile: %s, OutputFile: %s, LatColumn: %s, LngColumn: %s, Resolution: %d (%s), HasHeaders: %t, Overwrite: %t, Verbose: %t}",
		c.InputFile, c.OutputFile, c.LatColumn, c.LngColumn, c.Resolution, c.GetResolutionDescription(), c.HasHeaders, c.Overwrite, c.Verbose)
}