package integration

import (
	"os"
	"path/filepath"
	"testing"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// TestBasicFunctionality tests basic functionality to ensure core components work
func TestBasicFunctionality(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "csv-h3-basic-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test runner
	runner := NewTestRunner(tempDir)

	// Create simple test data
	headers := []string{"name", "latitude", "longitude"}
	records := [][]string{
		{"Test Location", "40.7128", "-74.0060"},
		{"Another Location", "51.5074", "-0.1278"},
	}

	inputFile, err := runner.CreateTestFile("basic_test", headers, records)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Configure processing
	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(tempDir, "basic_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    false,
	}

	// Process the file
	orchestrator := service.NewOrchestrator(cfg)
	result, err := orchestrator.ProcessFile()

	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	// Validate results
	if result.TotalRecords != 2 {
		t.Errorf("Expected 2 records, got %d", result.TotalRecords)
	}

	if result.ValidRecords != 2 {
		t.Errorf("Expected 2 valid records, got %d", result.ValidRecords)
	}

	// Validate output file exists
	if _, err := os.Stat(cfg.OutputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", cfg.OutputFile)
	}
}