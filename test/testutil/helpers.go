package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T, prefix string) string {
	t.Helper()
	
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Cleanup when test completes
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	
	return dir
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, dir, pattern string) *os.File {
	t.Helper()
	
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	// Cleanup when test completes
	t.Cleanup(func() {
		file.Close()
		os.Remove(file.Name())
	})
	
	return file
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists checks if a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	
	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// WithTimeout runs a function with a timeout
func WithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()
	
	done := make(chan bool, 1)
	
	go func() {
		fn()
		done <- true
	}()
	
	select {
	case <-done:
		// Function completed successfully
	case <-time.After(timeout):
		t.Errorf("Function timed out after %v", timeout)
	}
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	
	if testing.Short() {
		t.Skipf("Skipping in short mode: %s", reason)
	}
}

// RequireEnv checks if required environment variables are set
func RequireEnv(t *testing.T, vars ...string) {
	t.Helper()
	
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Skipf("Required environment variable not set: %s", v)
		}
	}
}