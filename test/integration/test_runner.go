package integration

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRunner provides utilities for running integration tests
type TestRunner struct {
	tempDir   string
	testFiles map[string]string
	results   []TestResult
}

// TestResult captures the results of a test run
type TestResult struct {
	Name           string
	Duration       time.Duration
	Success        bool
	Error          error
	RecordsTotal   int
	RecordsValid   int
	RecordsInvalid int
}

// NewTestRunner creates a new test runner instance
func NewTestRunner(tempDir string) *TestRunner {
	return &TestRunner{
		tempDir:   tempDir,
		testFiles: make(map[string]string),
		results:   make([]TestResult, 0),
	}
}

// CreateTestFile creates a test CSV file with the given data
func (tr *TestRunner) CreateTestFile(name string, headers []string, records [][]string) (string, error) {
	filePath := filepath.Join(tr.tempDir, name+".csv")
	
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create test file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if provided
	if headers != nil {
		if err := writer.Write(headers); err != nil {
			return "", fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// Write records
	for i, record := range records {
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}

	tr.testFiles[name] = filePath
	return filePath, nil
}

// GetTestFile returns the path to a test file by name
func (tr *TestRunner) GetTestFile(name string) (string, bool) {
	path, exists := tr.testFiles[name]
	return path, exists
}

// AddResult adds a test result to the runner
func (tr *TestRunner) AddResult(result TestResult) {
	tr.results = append(tr.results, result)
}

// GetResults returns all test results
func (tr *TestRunner) GetResults() []TestResult {
	return tr.results
}

// PrintSummary prints a summary of all test results
func (tr *TestRunner) PrintSummary() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("INTEGRATION TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	totalTests := len(tr.results)
	successCount := 0
	totalDuration := time.Duration(0)

	for _, result := range tr.results {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		} else {
			successCount++
		}

		totalDuration += result.Duration

		fmt.Printf("%-40s %s (%v)\n", result.Name, status, result.Duration)
		if result.RecordsTotal > 0 {
			fmt.Printf("    Records: %d total, %d valid, %d invalid\n", 
				result.RecordsTotal, result.RecordsValid, result.RecordsInvalid)
		}
		if !result.Success && result.Error != nil {
			fmt.Printf("    Error: %v\n", result.Error)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", totalTests-successCount)
	fmt.Printf("Success Rate: %.1f%%\n", float64(successCount)/float64(totalTests)*100)
	fmt.Printf("Total Duration: %v\n", totalDuration)
	fmt.Printf("Average Duration: %v\n", totalDuration/time.Duration(totalTests))
	fmt.Println(strings.Repeat("=", 80))
}

// ValidateOutputFile validates that an output file has the expected structure
func (tr *TestRunner) ValidateOutputFile(t *testing.T, outputPath string, expectedRecords int, hasHeaders bool) error {
	file, err := os.Open(outputPath)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read output CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("output file is empty")
	}

	startIdx := 0
	if hasHeaders {
		// Validate headers
		headers := records[0]
		if len(headers) == 0 {
			return fmt.Errorf("headers row is empty")
		}
		
		// Check that H3 index column is present
		if headers[len(headers)-1] != "h3_index" {
			return fmt.Errorf("expected last header to be 'h3_index', got '%s'", headers[len(headers)-1])
		}
		startIdx = 1
	}

	// Validate record count
	dataRecords := len(records) - startIdx
	if expectedRecords > 0 && dataRecords != expectedRecords {
		return fmt.Errorf("expected %d data records, got %d", expectedRecords, dataRecords)
	}

	// Validate each record has H3 index column
	for i := startIdx; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue // Skip empty rows
		}

		// Check that record has at least one column (the H3 index)
		if len(record) < 1 {
			return fmt.Errorf("record %d has no columns", i)
		}

		// H3 index is the last column
		h3Index := record[len(record)-1]
		
		// H3 index should be either empty (for invalid coordinates) or 15 characters
		if h3Index != "" && len(h3Index) != 15 {
			return fmt.Errorf("record %d: H3 index '%s' has invalid length (expected 15 or empty)", i, h3Index)
		}
	}

	return nil
}

// CompareFiles compares two CSV files and returns differences
func (tr *TestRunner) CompareFiles(file1, file2 string) ([]string, error) {
	records1, err := tr.readCSVFile(file1)
	if err != nil {
		return nil, fmt.Errorf("failed to read file1: %w", err)
	}

	records2, err := tr.readCSVFile(file2)
	if err != nil {
		return nil, fmt.Errorf("failed to read file2: %w", err)
	}

	var differences []string

	// Compare record counts
	if len(records1) != len(records2) {
		differences = append(differences, 
			fmt.Sprintf("Record count differs: file1=%d, file2=%d", len(records1), len(records2)))
	}

	// Compare records
	maxRecords := len(records1)
	if len(records2) > maxRecords {
		maxRecords = len(records2)
	}

	for i := 0; i < maxRecords; i++ {
		if i >= len(records1) {
			differences = append(differences, fmt.Sprintf("Record %d: missing in file1", i))
			continue
		}
		if i >= len(records2) {
			differences = append(differences, fmt.Sprintf("Record %d: missing in file2", i))
			continue
		}

		record1 := records1[i]
		record2 := records2[i]

		if len(record1) != len(record2) {
			differences = append(differences, 
				fmt.Sprintf("Record %d: column count differs: file1=%d, file2=%d", 
					i, len(record1), len(record2)))
			continue
		}

		for j := 0; j < len(record1); j++ {
			if record1[j] != record2[j] {
				differences = append(differences, 
					fmt.Sprintf("Record %d, Column %d: '%s' != '%s'", 
						i, j, record1[j], record2[j]))
			}
		}
	}

	return differences, nil
}

// readCSVFile reads a CSV file and returns all records
func (tr *TestRunner) readCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// CreateLargeTestFile creates a large CSV file for performance testing
func (tr *TestRunner) CreateLargeTestFile(name string, numRecords int) (string, error) {
	filePath := filepath.Join(tr.tempDir, name+".csv")
	
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create large test file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"id", "latitude", "longitude", "name", "category", "timestamp"}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %w", err)
	}

	// Generate records
	for i := 0; i < numRecords; i++ {
		lat := fmt.Sprintf("%.6f", float64((i%180)-90))    // -90 to 89
		lng := fmt.Sprintf("%.6f", float64((i%360)-180))   // -180 to 179
		id := fmt.Sprintf("LOC_%08d", i)
		name := fmt.Sprintf("Location_%d", i)
		category := fmt.Sprintf("Cat_%d", i%20)
		timestamp := fmt.Sprintf("2024-01-%02d", (i%28)+1)

		record := []string{id, lat, lng, name, category, timestamp}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}

	tr.testFiles[name] = filePath
	return filePath, nil
}

// CreateMalformedTestFile creates a CSV file with various malformed data
func (tr *TestRunner) CreateMalformedTestFile(name string) (string, error) {
	filePath := filepath.Join(tr.tempDir, name+".csv")
	
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create malformed test file: %w", err)
	}
	defer file.Close()

	// Write raw CSV content with various malformations
	content := `name,latitude,longitude,description
Valid,40.7128,-74.0060,Valid coordinates
Invalid Lat,91.0,0.0,Invalid latitude
Invalid Lng,0.0,181.0,Invalid longitude
Empty Lat,,0.0,Empty latitude
Empty Lng,0.0,,Empty longitude
Both Empty,,,Both coordinates empty
Non-numeric Lat,abc,0.0,Non-numeric latitude
Non-numeric Lng,0.0,xyz,Non-numeric longitude
Mixed Format,40.7abc,-74.0060,Mixed alphanumeric
Multiple Dots,40.71.28,-74.0060,Multiple decimal points
Scientific,1.23e2,4.56e1,Scientific notation
Whitespace,"  40.7128  ","-74.0060",Whitespace in coordinates
Quoted,"40.7128","-74.0060",Quoted coordinates
Extra Comma,40.7128,,-74.0060,Extra comma
Missing Column,40.7128,Missing longitude column
`

	if _, err := file.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write malformed content: %w", err)
	}

	tr.testFiles[name] = filePath
	return filePath, nil
}

// Cleanup removes all temporary test files
func (tr *TestRunner) Cleanup() error {
	return os.RemoveAll(tr.tempDir)
}

// GetTempDir returns the temporary directory path
func (tr *TestRunner) GetTempDir() string {
	return tr.tempDir
}

// ListTestFiles returns a list of all created test files
func (tr *TestRunner) ListTestFiles() map[string]string {
	result := make(map[string]string)
	for name, path := range tr.testFiles {
		result[name] = path
	}
	return result
}