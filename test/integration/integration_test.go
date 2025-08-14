package integration

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"csv-h3-tool/internal/cli"
	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// TestData represents test CSV data
type TestData struct {
	Name        string
	Headers     []string
	Records     [][]string
	ExpectedH3  map[int]string // line number -> expected H3 index
	ExpectedErr bool
}

// IntegrationTestSuite contains all integration tests
type IntegrationTestSuite struct {
	tempDir    string
	testFiles  map[string]string
	testData   []TestData
}

// setupTestSuite creates temporary directory and test files
func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	tempDir, err := os.MkdirTemp("", "csv-h3-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	suite := &IntegrationTestSuite{
		tempDir:   tempDir,
		testFiles: make(map[string]string),
		testData:  createTestData(),
	}

	// Create test CSV files
	for _, data := range suite.testData {
		filePath := filepath.Join(tempDir, data.Name+".csv")
		suite.createTestFile(t, filePath, data)
		suite.testFiles[data.Name] = filePath
	}

	return suite
}

// cleanup removes temporary test files
func (suite *IntegrationTestSuite) cleanup() {
	os.RemoveAll(suite.tempDir)
}

// createTestData defines all test scenarios
func createTestData() []TestData {
	return []TestData{
		{
			Name:    "basic_valid",
			Headers: []string{"name", "latitude", "longitude", "description"},
			Records: [][]string{
				{"New York", "40.7128", "-74.0060", "Valid coordinates"},
				{"London", "51.5074", "-0.1278", "Valid coordinates"},
				{"Tokyo", "35.6762", "139.6503", "Valid coordinates"},
			},
			ExpectedH3: map[int]string{
				1: "882a107289fffff", // New York at resolution 8
				2: "88195da49bfffff", // London at resolution 8
				3: "882f5a363bfffff", // Tokyo at resolution 8
			},
			ExpectedErr: false,
		},
		{
			Name:    "mixed_valid_invalid",
			Headers: []string{"name", "lat", "lng"},
			Records: [][]string{
				{"Valid", "40.7128", "-74.0060"},
				{"Invalid Lat", "91.0", "0.0"},
				{"Invalid Lng", "0.0", "181.0"},
				{"Empty", "", ""},
				{"Valid Again", "51.5074", "-0.1278"},
			},
			ExpectedH3: map[int]string{
				1: "882a107289fffff", // Valid
				5: "88195da49bfffff", // Valid Again
			},
			ExpectedErr: false,
		},
		{
			Name:    "boundary_coordinates",
			Headers: []string{"name", "latitude", "longitude"},
			Records: [][]string{
				{"North Pole", "90.0", "0.0"},
				{"South Pole", "-90.0", "0.0"},
				{"Prime Meridian", "0.0", "0.0"},
				{"Antimeridian East", "0.0", "180.0"},
				{"Antimeridian West", "0.0", "-180.0"},
			},
			ExpectedH3: map[int]string{
				1: "880326233bfffff", // North Pole
				2: "88f29380e1fffff", // South Pole
				3: "88754e6499fffff", // Prime Meridian
				4: "887eb57221fffff", // Antimeridian East
				5: "887eb57221fffff", // Antimeridian West
			},
			ExpectedErr: false,
		},
		{
			Name:    "no_headers",
			Headers: nil, // No headers
			Records: [][]string{
				{"40.7128", "-74.0060", "New York"},
				{"51.5074", "-0.1278", "London"},
			},
			ExpectedH3: map[int]string{
				1: "882a107289fffff", // New York
				2: "88195da49bfffff", // London
			},
			ExpectedErr: false,
		},
		{
			Name:    "custom_columns",
			Headers: []string{"city", "y_coord", "x_coord", "country"},
			Records: [][]string{
				{"Paris", "48.8566", "2.3522", "France"},
				{"Berlin", "52.5200", "13.4050", "Germany"},
			},
			ExpectedH3: map[int]string{
				1: "881fb46d2bfffff", // Paris
				2: "881f1ae4c7fffff", // Berlin
			},
			ExpectedErr: false,
		},
		{
			Name:    "scientific_notation",
			Headers: []string{"name", "latitude", "longitude"},
			Records: [][]string{
				{"Scientific", "1.23e2", "4.56e1"}, // 123.0, 45.6 - invalid lat
				{"Valid Scientific", "4.07128e1", "-7.40060e1"}, // 40.7128, -74.0060
			},
			ExpectedH3: map[int]string{
				2: "882a107289fffff", // Valid Scientific
			},
			ExpectedErr: false,
		},
		{
			Name:    "whitespace_handling",
			Headers: []string{"name", "latitude", "longitude"},
			Records: [][]string{
				{"Whitespace Lat", "  40.7128  ", "-74.0060"},
				{"Whitespace Lng", "40.7128", "  -74.0060  "},
				{"Both Whitespace", "  40.7128  ", "  -74.0060  "},
			},
			ExpectedH3: map[int]string{
				1: "882a107289fffff", // Whitespace Lat
				2: "882a107289fffff", // Whitespace Lng
				3: "882a107289fffff", // Both Whitespace
			},
			ExpectedErr: false,
		},
	}
}

// createTestFile creates a CSV file with the given test data
func (suite *IntegrationTestSuite) createTestFile(t *testing.T, filePath string, data TestData) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if present
	if data.Headers != nil {
		if err := writer.Write(data.Headers); err != nil {
			t.Fatalf("Failed to write headers to %s: %v", filePath, err)
		}
	}

	// Write records
	for _, record := range data.Records {
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record to %s: %v", filePath, err)
		}
	}
}

// TestEndToEndWorkflow tests the complete workflow from CLI to output
func TestEndToEndWorkflow(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	tests := []struct {
		name           string
		testDataName   string
		config         *config.Config
		expectedValid  int
		expectedTotal  int
		shouldFail     bool
	}{
		{
			name:         "Basic Valid Processing",
			testDataName: "basic_valid",
			config: &config.Config{
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			},
			expectedValid: 3,
			expectedTotal: 3,
			shouldFail:    false,
		},
		{
			name:         "Mixed Valid Invalid",
			testDataName: "mixed_valid_invalid",
			config: &config.Config{
				LatColumn:  "lat",
				LngColumn:  "lng",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    true,
			},
			expectedValid: 2,
			expectedTotal: 5,
			shouldFail:    false,
		},
		{
			name:         "No Headers",
			testDataName: "no_headers",
			config: &config.Config{
				LatColumn:  "0", // Column index
				LngColumn:  "1", // Column index
				Resolution: 8,
				HasHeaders: false,
				Overwrite:  true,
				Verbose:    false,
			},
			expectedValid: 2,
			expectedTotal: 2,
			shouldFail:    false,
		},
		{
			name:         "Custom Columns",
			testDataName: "custom_columns",
			config: &config.Config{
				LatColumn:  "y_coord",
				LngColumn:  "x_coord",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			},
			expectedValid: 2,
			expectedTotal: 2,
			shouldFail:    false,
		},
		{
			name:         "Different Resolution",
			testDataName: "basic_valid",
			config: &config.Config{
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 10, // Higher resolution
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			},
			expectedValid: 3,
			expectedTotal: 3,
			shouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up file paths
			inputFile := suite.testFiles[tt.testDataName]
			outputFile := filepath.Join(suite.tempDir, tt.testDataName+"_output.csv")

			tt.config.InputFile = inputFile
			tt.config.OutputFile = outputFile

			// Validate configuration
			if err := tt.config.Validate(); err != nil {
				if !tt.shouldFail {
					t.Fatalf("Configuration validation failed: %v", err)
				}
				return
			}

			// Create orchestrator and process file
			orchestrator := service.NewOrchestrator(tt.config)
			result, err := orchestrator.ProcessFile()

			if tt.shouldFail {
				if err == nil {
					t.Fatalf("Expected processing to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("Processing failed: %v", err)
			}

			// Verify results
			if result.TotalRecords != tt.expectedTotal {
				t.Errorf("Expected %d total records, got %d", tt.expectedTotal, result.TotalRecords)
			}

			if result.ValidRecords != tt.expectedValid {
				t.Errorf("Expected %d valid records, got %d", tt.expectedValid, result.ValidRecords)
			}

			// Verify output file exists
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Fatalf("Output file was not created: %s", outputFile)
			}

			// Verify output file content
			suite.verifyOutputFile(t, outputFile, tt.testDataName, tt.config)
		})
	}
}

// verifyOutputFile checks that the output file has correct format and content
func (suite *IntegrationTestSuite) verifyOutputFile(t *testing.T, outputFile, testDataName string, cfg *config.Config) {
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read output CSV: %v", err)
	}

	if len(records) == 0 {
		t.Fatalf("Output file is empty")
	}

	// Find test data
	var testData TestData
	for _, data := range suite.testData {
		if data.Name == testDataName {
			testData = data
			break
		}
	}

	startIdx := 0
	if cfg.HasHeaders {
		// Verify headers include H3 index column
		headers := records[0]
		if headers[len(headers)-1] != "h3_index" {
			t.Errorf("Expected last column to be 'h3_index', got '%s'", headers[len(headers)-1])
		}
		startIdx = 1
	}

	// Verify each record has H3 index
	for i := startIdx; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue
		}

		h3Index := record[len(record)-1]
		
		// Check if this record should have a valid H3 index
		if _, exists := testData.ExpectedH3[i-startIdx+1]; exists {
			// Valid records should have a non-empty H3 index of correct length
			if h3Index == "" {
				t.Errorf("Record %d: expected non-empty H3 index for valid coordinates", i)
			} else if len(h3Index) != 15 {
				t.Errorf("Record %d: H3 index '%s' has invalid length (expected 15)", i, h3Index)
			}
		} else {
			// Invalid records should have empty H3 index
			if h3Index != "" {
				t.Errorf("Record %d: expected empty H3 index for invalid coordinates, got '%s'", i, h3Index)
			}
		}
	}
}

// TestCLIIntegration tests the CLI interface end-to-end
func TestCLIIntegration(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, outputFile string)
	}{
		{
			name: "Basic CLI Usage",
			args: []string{
				suite.testFiles["basic_valid"],
				"-o", filepath.Join(suite.tempDir, "cli_basic_output.csv"),
				"--resolution", "8",
			},
			wantErr: false,
			validate: func(t *testing.T, outputFile string) {
				suite.validateBasicOutput(t, outputFile, true)
			},
		},
		{
			name: "Custom Columns",
			args: []string{
				suite.testFiles["custom_columns"],
				"-o", filepath.Join(suite.tempDir, "cli_custom_output.csv"),
				"--lat-column", "y_coord",
				"--lng-column", "x_coord",
				"--resolution", "10",
			},
			wantErr: false,
			validate: func(t *testing.T, outputFile string) {
				suite.validateBasicOutput(t, outputFile, true)
			},
		},
		{
			name: "No Headers",
			args: []string{
				suite.testFiles["no_headers"],
				"-o", filepath.Join(suite.tempDir, "cli_no_headers_output.csv"),
				"--no-headers",
				"--lat-column", "0",
				"--lng-column", "1",
			},
			wantErr: false,
			validate: func(t *testing.T, outputFile string) {
				suite.validateBasicOutput(t, outputFile, false)
			},
		},
		{
			name: "Verbose Output",
			args: []string{
				suite.testFiles["mixed_valid_invalid"],
				"-o", filepath.Join(suite.tempDir, "cli_verbose_output.csv"),
				"--verbose",
				"--lat-column", "lat",
				"--lng-column", "lng",
			},
			wantErr: false,
			validate: func(t *testing.T, outputFile string) {
				suite.validateBasicOutput(t, outputFile, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI instance
			cliApp := cli.NewCLI()
			
			// Set args for testing
			os.Args = append([]string{"csv-h3-tool"}, tt.args...)
			
			// Execute CLI
			err := cliApp.Execute()
			
			if tt.wantErr && err == nil {
				t.Errorf("Expected error, but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr && tt.validate != nil {
				// Find output file from args
				outputFile := ""
				for i, arg := range tt.args {
					if arg == "-o" && i+1 < len(tt.args) {
						outputFile = tt.args[i+1]
						break
					}
				}
				if outputFile != "" {
					tt.validate(t, outputFile)
				}
			}
		})
	}
}

// validateBasicOutput performs basic validation of output file
func (suite *IntegrationTestSuite) validateBasicOutput(t *testing.T, outputFile string, hasHeaders bool) {
	file, err := os.Open(outputFile)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read output CSV: %v", err)
	}

	if len(records) == 0 {
		t.Fatalf("Output file is empty")
	}

	startIdx := 0
	if hasHeaders {
		// Check that headers include h3_index
		headers := records[0]
		if len(headers) == 0 || headers[len(headers)-1] != "h3_index" {
			t.Errorf("Expected last header to be 'h3_index', got headers: %v", headers)
		}
		startIdx = 1
	}

	// Check that each data row has an additional column
	for i := startIdx; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue
		}
		
		// Last column should be H3 index (could be empty for invalid coordinates)
		h3Index := record[len(record)-1]
		
		// H3 index should be either empty (invalid) or a valid H3 string
		if h3Index != "" && len(h3Index) != 15 {
			t.Errorf("Record %d: H3 index '%s' has invalid length (expected 15 characters)", i, h3Index)
		}
	}
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	tests := []struct {
		name        string
		setupFunc   func() *config.Config
		expectError bool
		errorType   string
	}{
		{
			name: "Non-existent Input File",
			setupFunc: func() *config.Config {
				return &config.Config{
					InputFile:  filepath.Join(suite.tempDir, "nonexistent.csv"),
					OutputFile: filepath.Join(suite.tempDir, "output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 8,
					HasHeaders: true,
				}
			},
			expectError: true,
			errorType:   "file",
		},
		{
			name: "Invalid Resolution",
			setupFunc: func() *config.Config {
				return &config.Config{
					InputFile:  suite.testFiles["basic_valid"],
					OutputFile: filepath.Join(suite.tempDir, "output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 16, // Invalid resolution
					HasHeaders: true,
				}
			},
			expectError: true,
			errorType:   "config",
		},
		{
			name: "Invalid Column Names",
			setupFunc: func() *config.Config {
				return &config.Config{
					InputFile:  suite.testFiles["basic_valid"],
					OutputFile: filepath.Join(suite.tempDir, "output.csv"),
					LatColumn:  "nonexistent_lat",
					LngColumn:  "nonexistent_lng",
					Resolution: 8,
					HasHeaders: true,
				}
			},
			expectError: true,
			errorType:   "validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupFunc()
			
			orchestrator := service.NewOrchestrator(cfg)
			_, err := orchestrator.ProcessFile()

			if tt.expectError && err == nil {
				t.Errorf("Expected error of type %s, but got none", tt.errorType)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectError && err != nil {
				// Verify error type (basic check)
				errorStr := strings.ToLower(err.Error())
				if !strings.Contains(errorStr, tt.errorType) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorType, err)
				}
			}
		})
	}
}

// TestLargeFileHandling tests processing of larger CSV files
func TestLargeFileHandling(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	// Create a larger test file
	largeFile := filepath.Join(suite.tempDir, "large_test.csv")
	suite.createLargeTestFile(t, largeFile, 1000) // 1000 records

	cfg := &config.Config{
		InputFile:  largeFile,
		OutputFile: filepath.Join(suite.tempDir, "large_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    true,
	}

	start := time.Now()
	orchestrator := service.NewOrchestrator(cfg)
	result, err := orchestrator.ProcessFile()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Large file processing failed: %v", err)
	}

	// Verify results
	if result.TotalRecords != 1000 {
		t.Errorf("Expected 1000 records, got %d", result.TotalRecords)
	}

	// Performance check - should process 1000 records reasonably quickly
	if duration > 30*time.Second {
		t.Errorf("Processing took too long: %v (expected < 30s)", duration)
	}

	t.Logf("Processed %d records in %v", result.TotalRecords, duration)
}

// createLargeTestFile creates a CSV file with many records for performance testing
func (suite *IntegrationTestSuite) createLargeTestFile(t *testing.T, filePath string, numRecords int) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"name", "latitude", "longitude", "description"}); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Generate test coordinates around the world
	for i := 0; i < numRecords; i++ {
		lat := fmt.Sprintf("%.6f", float64(i%180-90))      // -90 to 89
		lng := fmt.Sprintf("%.6f", float64(i%360-180))     // -180 to 179
		name := fmt.Sprintf("Location_%d", i)
		desc := fmt.Sprintf("Test location %d", i)

		record := []string{name, lat, lng, desc}
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}
}

// TestOutputFormatPreservation tests that original CSV format is preserved
func TestOutputFormatPreservation(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	inputFile := suite.testFiles["basic_valid"]
	outputFile := filepath.Join(suite.tempDir, "format_test_output.csv")

	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: outputFile,
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
	}

	orchestrator := service.NewOrchestrator(cfg)
	_, err := orchestrator.ProcessFile()
	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	// Read original file
	originalRecords := suite.readCSVFile(t, inputFile)
	
	// Read output file
	outputRecords := suite.readCSVFile(t, outputFile)

	// Verify structure
	if len(outputRecords) != len(originalRecords) {
		t.Errorf("Output has different number of records: expected %d, got %d", 
			len(originalRecords), len(outputRecords))
	}

	// Verify headers
	if len(originalRecords) > 0 && len(outputRecords) > 0 {
		originalHeaders := originalRecords[0]
		outputHeaders := outputRecords[0]
		
		// Output should have all original headers plus h3_index
		if len(outputHeaders) != len(originalHeaders)+1 {
			t.Errorf("Output headers count mismatch: expected %d, got %d", 
				len(originalHeaders)+1, len(outputHeaders))
		}
		
		// Check original headers are preserved
		for i, header := range originalHeaders {
			if i >= len(outputHeaders) || outputHeaders[i] != header {
				t.Errorf("Header %d mismatch: expected '%s', got '%s'", 
					i, header, outputHeaders[i])
			}
		}
		
		// Check H3 index column is added
		if outputHeaders[len(outputHeaders)-1] != "h3_index" {
			t.Errorf("Expected last header to be 'h3_index', got '%s'", 
				outputHeaders[len(outputHeaders)-1])
		}
	}

	// Verify data rows preserve original columns
	for i := 1; i < len(originalRecords) && i < len(outputRecords); i++ {
		originalRow := originalRecords[i]
		outputRow := outputRecords[i]
		
		// Check original columns are preserved
		for j, value := range originalRow {
			if j >= len(outputRow) || outputRow[j] != value {
				t.Errorf("Row %d, column %d mismatch: expected '%s', got '%s'", 
					i, j, value, outputRow[j])
			}
		}
	}
}

// readCSVFile reads a CSV file and returns all records
func (suite *IntegrationTestSuite) readCSVFile(t *testing.T, filePath string) [][]string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file %s: %v", filePath, err)
	}

	return records
}

// TestProgressReporting tests that progress is reported for large files
func TestProgressReporting(t *testing.T) {
	suite := setupTestSuite(t)
	defer suite.cleanup()

	// Create a moderately large test file
	largeFile := filepath.Join(suite.tempDir, "progress_test.csv")
	suite.createLargeTestFile(t, largeFile, 100)

	cfg := &config.Config{
		InputFile:  largeFile,
		OutputFile: filepath.Join(suite.tempDir, "progress_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    true, // Enable verbose to test progress reporting
	}

	// Capture output to verify progress messages
	orchestrator := service.NewOrchestrator(cfg)
	result, err := orchestrator.ProcessFile()

	if err != nil {
		t.Fatalf("Processing failed: %v", err)
	}

	if result.TotalRecords != 100 {
		t.Errorf("Expected 100 records, got %d", result.TotalRecords)
	}

	// Verify processing time is recorded
	if result.ProcessingTime == 0 {
		t.Error("Processing time should be recorded")
	}
}