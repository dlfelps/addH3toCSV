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
	os.MkdirAll("../testdata", 0755)
	
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
	// This function can be used to create shared test data files
	// Individual tests should create their own specific test data
}

// validateTestEnvironment checks if the test environment is properly set up
func validateTestEnvironment(t *testing.T) {
	// Check if required directories exist
	if _, err := os.Stat("../testdata"); os.IsNotExist(err) {
		t.Fatal("Test data directory does not exist")
	}
	
	// Check if Go module is available
	if _, err := os.Stat("../../go.mod"); os.IsNotExist(err) {
		t.Fatal("Go module not found - tests must be run from project root")
	}
}