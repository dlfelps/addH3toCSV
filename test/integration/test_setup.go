package integration

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain provides setup and teardown for integration tests
func TestMain(m *testing.M) {
	// Setup
	setupTestEnvironment()
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	cleanupTestEnvironment()
	
	os.Exit(code)
}

// setupTestEnvironment prepares the test environment
func setupTestEnvironment() {
	// Ensure test data directory exists
	os.MkdirAll("testdata", 0755)
	
	// Create any required test files
	createRequiredTestFiles()
}

// cleanupTestEnvironment cleans up after tests
func cleanupTestEnvironment() {
	// Clean up any temporary files created during tests
	// Note: Individual tests should clean up their own temp files
}

// createRequiredTestFiles creates any test files that are required by multiple tests
func createRequiredTestFiles() {
	// Create basic test data files if they don't exist
	testDataDir := "testdata"
	
	// Create a simple test file for basic validation
	basicTestFile := filepath.Join(testDataDir, "basic_test.csv")
	if _, err := os.Stat(basicTestFile); os.IsNotExist(err) {
		file, err := os.Create(basicTestFile)
		if err == nil {
			file.WriteString("name,latitude,longitude\n")
			file.WriteString("Test Location,40.7128,-74.0060\n")
			file.Close()
		}
	}
}

// validateTestEnvironment checks if the test environment is properly set up
func validateTestEnvironment(t *testing.T) {
	// Check if required directories exist
	if _, err := os.Stat("testdata"); os.IsNotExist(err) {
		// Create it if it doesn't exist
		os.MkdirAll("testdata", 0755)
	}
	
	// Check if Go module is available
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		t.Skip("Go module not found - tests must be run from project root")
	}
}