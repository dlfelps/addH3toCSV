package integration

import (
	"os"
	"path/filepath"
	"testing"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// TestSimpleProcessing tests the most basic processing functionality
func TestSimpleProcessing(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "csv-h3-simple-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple CSV file manually
	inputFile := filepath.Join(tempDir, "simple.csv")
	file, err := os.Create(inputFile)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}
	
	// Write simple CSV content
	_, err = file.WriteString("name,latitude,longitude\n")
	if err != nil {
		file.Close()
		t.Fatalf("Failed to write header: %v", err)
	}
	
	_, err = file.WriteString("Test,40.7128,-74.0060\n")
	if err != nil {
		file.Close()
		t.Fatalf("Failed to write data: %v", err)
	}
	
	file.Close()

	// Configure processing
	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(tempDir, "simple_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    true, // Enable verbose for debugging
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}

	// Process the file
	orchestrator := service.NewOrchestrator(cfg)
	if orchestrator == nil {
		t.Fatal("Failed to create orchestrator")
	}

	result, err := orchestrator.ProcessFile()
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	// Validate results
	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.TotalRecords != 1 {
		t.Errorf("Expected 1 record, got %d", result.TotalRecords)
	}

	if result.ValidRecords != 1 {
		t.Errorf("Expected 1 valid record, got %d", result.ValidRecords)
	}

	if result.InvalidRecords != 0 {
		t.Errorf("Expected 0 invalid records, got %d", result.InvalidRecords)
	}

	// Check output file exists
	if _, err := os.Stat(cfg.OutputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", cfg.OutputFile)
	}

	t.Logf("Simple processing test completed successfully")
	t.Logf("Processed %d records in %v", result.TotalRecords, result.ProcessingTime)
}