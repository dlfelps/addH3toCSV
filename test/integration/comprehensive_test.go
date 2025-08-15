package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// TestComprehensiveScenarios tests various real-world scenarios
func TestComprehensiveScenarios(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "csv-h3-comprehensive-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	runner := NewTestRunner(tempDir)

	// Test scenarios
	scenarios := []struct {
		name        string
		description string
		setupFunc   func() (string, *config.Config, int, int) // returns inputFile, config, expectedTotal, expectedValid
		validate    func(t *testing.T, outputFile string, result *service.ProcessResult)
	}{
		{
			name:        "RealWorldCities",
			description: "Process real world city coordinates",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"city", "country", "latitude", "longitude", "population"}
				records := [][]string{
					{"New York", "USA", "40.7128", "-74.0060", "8175133"},
					{"London", "UK", "51.5074", "-0.1278", "8982000"},
					{"Tokyo", "Japan", "35.6762", "139.6503", "13929286"},
					{"Sydney", "Australia", "-33.8688", "151.2093", "5312163"},
					{"São Paulo", "Brazil", "-23.5505", "-46.6333", "12325232"},
					{"Cairo", "Egypt", "30.0444", "31.2357", "10230350"},
					{"Mumbai", "India", "19.0760", "72.8777", "20411274"},
					{"Moscow", "Russia", "55.7558", "37.6176", "12506468"},
				}
				
				inputFile, err := runner.CreateTestFile("real_cities", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "real_cities_output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 8,
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 8, 8
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if err := runner.ValidateOutputFile(t, outputFile, 8, true); err != nil {
					t.Errorf("Output validation failed: %v", err)
				}
			},
		},
		{
			name:        "MixedValidInvalid",
			description: "Process mixed valid and invalid coordinates",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"id", "lat", "lng", "status"}
				records := [][]string{
					{"1", "40.7128", "-74.0060", "valid"},
					{"2", "91.0", "0.0", "invalid_lat"},
					{"3", "0.0", "181.0", "invalid_lng"},
					{"4", "", "", "empty"},
					{"5", "abc", "xyz", "malformed"},
					{"6", "51.5074", "-0.1278", "valid"},
					{"7", "35.6762", "139.6503", "valid"},
					{"8", "-91.0", "0.0", "invalid_lat"},
					{"9", "0.0", "-181.0", "invalid_lng"},
					{"10", "48.8566", "2.3522", "valid"},
				}
				
				inputFile, err := runner.CreateTestFile("mixed_data", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "mixed_data_output.csv"),
					LatColumn:  "lat",
					LngColumn:  "lng",
					Resolution: 8,
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    true,
				}
				return inputFile, cfg, 10, 4 // 4 valid out of 10 total
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 4 {
					t.Errorf("Expected 4 valid records, got %d", result.ValidRecords)
				}
				if result.InvalidRecords != 6 {
					t.Errorf("Expected 6 invalid records, got %d", result.InvalidRecords)
				}
			},
		},
		{
			name:        "BoundaryCoordinates",
			description: "Test coordinates at world boundaries",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"location", "latitude", "longitude", "type"}
				records := [][]string{
					{"North Pole", "90.0", "0.0", "pole"},
					{"South Pole", "-90.0", "0.0", "pole"},
					{"Prime Meridian", "0.0", "0.0", "meridian"},
					{"Antimeridian East", "0.0", "180.0", "meridian"},
					{"Antimeridian West", "0.0", "-180.0", "meridian"},
					{"Near North Pole", "89.999", "45.0", "near_pole"},
					{"Near South Pole", "-89.999", "-45.0", "near_pole"},
					{"Max East", "45.0", "179.999", "boundary"},
					{"Max West", "45.0", "-179.999", "boundary"},
				}
				
				inputFile, err := runner.CreateTestFile("boundaries", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "boundaries_output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 8,
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 9, 9
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 9 {
					t.Errorf("Expected 9 valid records, got %d", result.ValidRecords)
				}
			},
		},
		{
			name:        "HighPrecisionCoordinates",
			description: "Test high precision coordinate values",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"name", "latitude", "longitude", "precision"}
				records := [][]string{
					{"High Precision 1", "40.123456789012345", "-74.987654321098765", "15_digits"},
					{"High Precision 2", "51.507400000000000", "-0.127800000000000", "trailing_zeros"},
					{"Scientific 1", "4.07128e1", "-7.40060e1", "scientific"},
					{"Scientific 2", "1.23456e-2", "9.87654e-3", "small_scientific"},
					{"Mixed Precision", "40.7", "-74.006", "mixed"},
					{"Integer Coords", "41", "-74", "integer"},
					{"Leading Zeros", "040.7128", "-074.0060", "leading_zeros"},
				}
				
				inputFile, err := runner.CreateTestFile("precision", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "precision_output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 12, // Higher resolution for precision test
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 7, 7
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 7 {
					t.Errorf("Expected 7 valid records, got %d", result.ValidRecords)
				}
			},
		},
		{
			name:        "NoHeadersWithIndices",
			description: "Process CSV without headers using column indices",
			setupFunc: func() (string, *config.Config, int, int) {
				records := [][]string{
					{"40.7128", "-74.0060", "New York", "USA"},
					{"51.5074", "-0.1278", "London", "UK"},
					{"35.6762", "139.6503", "Tokyo", "Japan"},
					{"48.8566", "2.3522", "Paris", "France"},
				}
				
				inputFile, err := runner.CreateTestFile("no_headers", nil, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "no_headers_output.csv"),
					LatColumn:  "0", // First column
					LngColumn:  "1", // Second column
					Resolution: 8,
					HasHeaders: false,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 4, 4
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if err := runner.ValidateOutputFile(t, outputFile, 4, false); err != nil {
					t.Errorf("Output validation failed: %v", err)
				}
			},
		},
		{
			name:        "CustomColumnNames",
			description: "Process CSV with custom column names",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"location_id", "y_coordinate", "x_coordinate", "place_name", "country_code"}
				records := [][]string{
					{"LOC001", "40.7128", "-74.0060", "New York", "US"},
					{"LOC002", "51.5074", "-0.1278", "London", "GB"},
					{"LOC003", "35.6762", "139.6503", "Tokyo", "JP"},
					{"LOC004", "-33.8688", "151.2093", "Sydney", "AU"},
				}
				
				inputFile, err := runner.CreateTestFile("custom_columns", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "custom_columns_output.csv"),
					LatColumn:  "y_coordinate",
					LngColumn:  "x_coordinate",
					Resolution: 8,
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 4, 4
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 4 {
					t.Errorf("Expected 4 valid records, got %d", result.ValidRecords)
				}
			},
		},
		{
			name:        "DifferentResolutions",
			description: "Test different H3 resolution levels",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"name", "latitude", "longitude"}
				records := [][]string{
					{"Test Location", "40.7128", "-74.0060"},
				}
				
				inputFile, err := runner.CreateTestFile("resolution_test", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "resolution_test_output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 15, // Maximum resolution
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 1, 1
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 1 {
					t.Errorf("Expected 1 valid record, got %d", result.ValidRecords)
				}
			},
		},
		{
			name:        "WhitespaceHandling",
			description: "Test handling of whitespace in coordinates",
			setupFunc: func() (string, *config.Config, int, int) {
				headers := []string{"name", "latitude", "longitude"}
				records := [][]string{
					{"Leading Spaces", "  40.7128", "-74.0060"},
					{"Trailing Spaces", "40.7128  ", "-74.0060"},
					{"Both Spaces", "  40.7128  ", "  -74.0060  "},
					{"Tab Characters", "40.7128\t", "\t-74.0060"},
					{"Mixed Whitespace", " \t40.7128 \t", " \t-74.0060 \t"},
				}
				
				inputFile, err := runner.CreateTestFile("whitespace", headers, records)
				if err != nil {
					panic(fmt.Sprintf("Failed to create test file: %v", err))
				}
				cfg := &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "whitespace_output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 8,
					HasHeaders: true,
					Overwrite:  true,
					Verbose:    false,
				}
				return inputFile, cfg, 5, 5
			},
			validate: func(t *testing.T, outputFile string, result *service.ProcessResult) {
				if result.ValidRecords != 5 {
					t.Errorf("Expected 5 valid records, got %d", result.ValidRecords)
				}
			},
		},
	}

	// Run all scenarios
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			start := time.Now()
			
			inputFile, cfg, expectedTotal, expectedValid := scenario.setupFunc()
			
			// Validate that input file was created
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				t.Fatalf("Input file was not created: %s", inputFile)
			}
			
			// Process the file
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			
			duration := time.Since(start)
			
			// Record result
			testResult := TestResult{
				Name:           scenario.name,
				Duration:       duration,
				Success:        err == nil,
				Error:          err,
				RecordsTotal:   0,
				RecordsValid:   0,
				RecordsInvalid: 0,
			}
			
			if result != nil {
				testResult.RecordsTotal = result.TotalRecords
				testResult.RecordsValid = result.ValidRecords
				testResult.RecordsInvalid = result.InvalidRecords
			}
			
			runner.AddResult(testResult)
			
			if err != nil {
				t.Errorf("Scenario %s failed: %v", scenario.name, err)
				t.Logf("Configuration: %+v", cfg)
				t.Logf("Input file: %s", inputFile)
				return // Skip validation if processing failed
			}
			
			// Validate basic expectations
			if result.TotalRecords != expectedTotal {
				t.Errorf("Expected %d total records, got %d", expectedTotal, result.TotalRecords)
			}
			
			if result.ValidRecords != expectedValid {
				t.Errorf("Expected %d valid records, got %d", expectedValid, result.ValidRecords)
			}
			
			// Run custom validation if provided
			if scenario.validate != nil {
				scenario.validate(t, cfg.OutputFile, result)
			}
			
			t.Logf("Scenario %s completed in %v: %d/%d valid records", 
				scenario.name, duration, result.ValidRecords, result.TotalRecords)
		})
	}
	
	// Print summary
	runner.PrintSummary()
}

// TestErrorRecovery tests the system's ability to recover from various errors
func TestErrorRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "csv-h3-error-recovery-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	runner := NewTestRunner(tempDir)

	errorTests := []struct {
		name        string
		setupFunc   func() *config.Config
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "NonExistentInputFile",
			setupFunc: func() *config.Config {
				return &config.Config{
					InputFile:  filepath.Join(tempDir, "nonexistent.csv"),
					OutputFile: filepath.Join(tempDir, "output.csv"),
					LatColumn:  "latitude",
					LngColumn:  "longitude",
					Resolution: 8,
					HasHeaders: true,
				}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil // Any error is expected
			},
		},
		{
			name: "InvalidResolution",
			setupFunc: func() *config.Config {
				inputFile, _ := runner.CreateTestFile("test", 
					[]string{"lat", "lng"}, 
					[][]string{{"40.7", "-74.0"}})
				
				return &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "output.csv"),
					LatColumn:  "lat",
					LngColumn:  "lng",
					Resolution: 16, // Invalid resolution
					HasHeaders: true,
				}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil
			},
		},
		{
			name: "MissingColumns",
			setupFunc: func() *config.Config {
				inputFile, _ := runner.CreateTestFile("test", 
					[]string{"name", "value"}, 
					[][]string{{"test", "123"}})
				
				return &config.Config{
					InputFile:  inputFile,
					OutputFile: filepath.Join(tempDir, "output.csv"),
					LatColumn:  "latitude", // Column doesn't exist
					LngColumn:  "longitude", // Column doesn't exist
					Resolution: 8,
					HasHeaders: true,
				}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil
			},
		},
		{
			name: "ReadOnlyOutputDirectory",
			setupFunc: func() *config.Config {
				inputFile, _ := runner.CreateTestFile("test", 
					[]string{"lat", "lng"}, 
					[][]string{{"40.7", "-74.0"}})
				
				// Try to write to root directory (should fail on most systems)
				return &config.Config{
					InputFile:  inputFile,
					OutputFile: "/root/output.csv", // Should fail due to permissions
					LatColumn:  "lat",
					LngColumn:  "lng",
					Resolution: 8,
					HasHeaders: true,
				}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil
			},
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			cfg := test.setupFunc()
			
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			
			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for test %s, but got none", test.name)
				} else if !test.errorCheck(err) {
					t.Errorf("Error check failed for test %s: %v", test.name, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for test %s: %v", test.name, err)
				}
				if result == nil {
					t.Errorf("Expected result for test %s, but got nil", test.name)
				}
			}
		})
	}
}

// TestPerformanceCharacteristics tests performance under various conditions
func TestPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-performance-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	runner := NewTestRunner(tempDir)

	performanceTests := []struct {
		name        string
		numRecords  int
		maxDuration time.Duration
	}{
		{"Small_100", 100, 5 * time.Second},
		{"Medium_1000", 1000, 15 * time.Second},
		{"Large_5000", 5000, 60 * time.Second},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			// Create test file
			inputFile, err := runner.CreateLargeTestFile(test.name, test.numRecords)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, test.name+"_output.csv"),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			start := time.Now()
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Performance test %s failed: %v", test.name, err)
			}

			if result.TotalRecords != test.numRecords {
				t.Errorf("Expected %d records, got %d", test.numRecords, result.TotalRecords)
			}

			if duration > test.maxDuration {
				t.Errorf("Performance test %s took too long: %v (max: %v)", 
					test.name, duration, test.maxDuration)
			}

			recordsPerSecond := float64(test.numRecords) / duration.Seconds()
			t.Logf("Performance test %s: %d records in %v (%.2f records/sec)", 
				test.name, test.numRecords, duration, recordsPerSecond)
		})
	}
}