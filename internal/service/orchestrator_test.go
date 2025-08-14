package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"csv-h3-tool/internal/config"
)

// TestOrchestrator_ProcessFile tests the complete workflow integration
func TestOrchestrator_ProcessFile(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "orchestrator_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test CSV file
	inputFile := filepath.Join(tempDir, "test_input.csv")
	testCSV := `latitude,longitude,name
40.7128,-74.0060,New York
34.0522,-118.2437,Los Angeles
41.8781,-87.6298,Chicago
29.7604,-95.3698,Houston
33.4484,-112.0740,Phoenix
`
	if err := os.WriteFile(inputFile, []byte(testCSV), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test_output.csv")

	// Create configuration
	cfg := config.NewConfig()
	cfg.InputFile = inputFile
	cfg.OutputFile = outputFile
	cfg.LatColumn = "latitude"
	cfg.LngColumn = "longitude"
	cfg.Resolution = 8
	cfg.HasHeaders = true
	cfg.Overwrite = true
	cfg.Verbose = false

	// Create orchestrator
	orchestrator := NewOrchestrator(cfg)

	// Test component validation
	if err := orchestrator.ValidateComponents(); err != nil {
		t.Fatalf("Component validation failed: %v", err)
	}

	// Process the file
	result, err := orchestrator.ProcessFile()
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Validate results
	if result == nil {
		t.Fatal("ProcessResult is nil")
	}

	if result.TotalRecords != 5 {
		t.Errorf("Expected 5 total records, got %d", result.TotalRecords)
	}

	if result.ValidRecords != 5 {
		t.Errorf("Expected 5 valid records, got %d", result.ValidRecords)
	}

	if result.InvalidRecords != 0 {
		t.Errorf("Expected 0 invalid records, got %d", result.InvalidRecords)
	}

	if result.ProcessingTime <= 0 {
		t.Error("Processing time should be greater than 0")
	}

	if result.OutputFile != outputFile {
		t.Errorf("Expected output file %s, got %s", outputFile, result.OutputFile)
	}

	// Verify output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and validate output file content
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Should have header + 5 data rows
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines in output (header + 5 data), got %d", len(lines))
	}

	// Check header includes h3_index column
	header := lines[0]
	if !strings.Contains(header, "h3_index") {
		t.Error("Output header should contain h3_index column")
	}

	// Check that each data row has an H3 index
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		if len(fields) < 4 {
			t.Errorf("Line %d should have at least 4 fields (original 3 + h3_index)", i)
		}
		
		h3Index := fields[len(fields)-1]
		if h3Index == "" {
			t.Errorf("Line %d should have a non-empty H3 index", i)
		}
		
		// H3 indexes at resolution 8 should be 15 characters long
		if len(h3Index) != 15 {
			t.Errorf("Line %d H3 index should be 15 characters, got %d: %s", i, len(h3Index), h3Index)
		}
	}
}

// TestOrchestrator_ProcessFileWithInvalidData tests handling of invalid coordinates
func TestOrchestrator_ProcessFileWithInvalidData(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "orchestrator_invalid_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test CSV file with some invalid data
	inputFile := filepath.Join(tempDir, "test_invalid.csv")
	testCSV := `latitude,longitude,name
40.7128,-74.0060,New York
999.0,-74.0060,Invalid Lat
34.0522,-999.0,Invalid Lng
invalid,invalid,Invalid Both
41.8781,-87.6298,Chicago
`
	if err := os.WriteFile(inputFile, []byte(testCSV), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test_invalid_output.csv")

	// Create configuration
	cfg := config.NewConfig()
	cfg.InputFile = inputFile
	cfg.OutputFile = outputFile
	cfg.LatColumn = "latitude"
	cfg.LngColumn = "longitude"
	cfg.Resolution = 8
	cfg.HasHeaders = true
	cfg.Overwrite = true
	cfg.Verbose = true

	// Create orchestrator
	orchestrator := NewOrchestrator(cfg)

	// Process the file
	result, err := orchestrator.ProcessFile()
	if err != nil {
		t.Fatalf("ProcessFile failed: %v", err)
	}

	// Validate results - should handle invalid data gracefully
	if result.TotalRecords != 5 {
		t.Errorf("Expected 5 total records, got %d", result.TotalRecords)
	}

	if result.ValidRecords != 2 {
		t.Errorf("Expected 2 valid records, got %d", result.ValidRecords)
	}

	if result.InvalidRecords != 3 {
		t.Errorf("Expected 3 invalid records, got %d", result.InvalidRecords)
	}

	// Verify output file exists and has correct structure
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(outputContent)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Should still have all rows, but invalid ones should have empty H3 index
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines in output, got %d", len(lines))
	}
}

// TestOrchestrator_ValidateComponents tests component validation
func TestOrchestrator_ValidateComponents(t *testing.T) {
	cfg := config.NewConfig()
	orchestrator := NewOrchestrator(cfg)

	// Should pass validation with properly initialized components
	if err := orchestrator.ValidateComponents(); err != nil {
		t.Errorf("ValidateComponents should pass with properly initialized orchestrator: %v", err)
	}

	// Test with nil components
	orchestrator.validator = nil
	if err := orchestrator.ValidateComponents(); err == nil {
		t.Error("ValidateComponents should fail with nil validator")
	}

	orchestrator = NewOrchestrator(cfg)
	orchestrator.h3Generator = nil
	if err := orchestrator.ValidateComponents(); err == nil {
		t.Error("ValidateComponents should fail with nil h3Generator")
	}

	orchestrator = NewOrchestrator(cfg)
	orchestrator.processor = nil
	if err := orchestrator.ValidateComponents(); err == nil {
		t.Error("ValidateComponents should fail with nil processor")
	}

	orchestrator = NewOrchestrator(cfg)
	orchestrator.config = nil
	if err := orchestrator.ValidateComponents(); err == nil {
		t.Error("ValidateComponents should fail with nil config")
	}
}

// TestOrchestrator_ConfigurationHandling tests configuration management
func TestOrchestrator_ConfigurationHandling(t *testing.T) {
	cfg := config.NewConfig()
	orchestrator := NewOrchestrator(cfg)

	// Test GetConfig
	retrievedConfig := orchestrator.GetConfig()
	if retrievedConfig != cfg {
		t.Error("GetConfig should return the same config instance")
	}

	// Test SetConfig
	newConfig := config.NewConfig()
	newConfig.Resolution = 10
	orchestrator.SetConfig(newConfig)

	if orchestrator.GetConfig().Resolution != 10 {
		t.Error("SetConfig should update the configuration")
	}
}

// TestProgressReporter tests progress reporting functionality
func TestProgressReporter(t *testing.T) {
	reporter := NewProgressReporter(1000, true)

	if reporter.fileSize != 1000 {
		t.Errorf("Expected file size 1000, got %d", reporter.fileSize)
	}

	if !reporter.verbose {
		t.Error("Expected verbose to be true")
	}

	// Test progress updates (should not panic)
	reporter.UpdateProgress(100)
	reporter.UpdateProgress(500)
	reporter.UpdateProgress(1000)
	reporter.Complete()

	// Test non-verbose mode
	quietReporter := NewProgressReporter(1000, false)
	quietReporter.UpdateProgress(100)
	quietReporter.Complete()
}

// BenchmarkOrchestrator_ProcessFile benchmarks the complete processing workflow
func BenchmarkOrchestrator_ProcessFile(b *testing.B) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "orchestrator_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test CSV file with more data for meaningful benchmark
	inputFile := filepath.Join(tempDir, "benchmark_input.csv")
	var csvBuilder strings.Builder
	csvBuilder.WriteString("latitude,longitude,name\n")
	
	// Generate 1000 test records
	for i := 0; i < 1000; i++ {
		lat := 40.0 + float64(i%90)/100.0  // Vary latitude
		lng := -74.0 + float64(i%180)/100.0 // Vary longitude
		csvBuilder.WriteString(fmt.Sprintf("%.4f,%.4f,Location_%d\n", lat, lng, i))
	}
	
	if err := os.WriteFile(inputFile, []byte(csvBuilder.String()), 0644); err != nil {
		b.Fatalf("Failed to create benchmark CSV file: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(tempDir, fmt.Sprintf("benchmark_output_%d.csv", i))

		// Create configuration
		cfg := config.NewConfig()
		cfg.InputFile = inputFile
		cfg.OutputFile = outputFile
		cfg.LatColumn = "latitude"
		cfg.LngColumn = "longitude"
		cfg.Resolution = 8
		cfg.HasHeaders = true
		cfg.Overwrite = true
		cfg.Verbose = false

		// Create orchestrator and process
		orchestrator := NewOrchestrator(cfg)
		_, err := orchestrator.ProcessFile()
		if err != nil {
			b.Fatalf("ProcessFile failed: %v", err)
		}
	}
}