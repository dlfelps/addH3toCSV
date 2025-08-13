package filehandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler provides file path handling and validation functionality
type FileHandler struct{}

// NewFileHandler creates a new file handler instance
func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// ValidateInputFile checks if the input file exists and is readable
func (fh *FileHandler) ValidateInputFile(path string) error {
	if path == "" {
		return fmt.Errorf("input file path cannot be empty")
	}
	
	// Clean the path
	cleanPath := filepath.Clean(path)
	
	// Check if file exists
	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", cleanPath)
	}
	if err != nil {
		return fmt.Errorf("cannot access input file %s: %w", cleanPath, err)
	}
	
	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("input path is not a regular file: %s", cleanPath)
	}
	
	// Check if file is readable
	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("cannot read input file %s: %w", cleanPath, err)
	}
	file.Close()
	
	return nil
}

// ValidateOutputFile checks if the output file can be created or overwritten
func (fh *FileHandler) ValidateOutputFile(path string, overwrite bool) error {
	if path == "" {
		return fmt.Errorf("output file path cannot be empty")
	}
	
	// Clean the path
	cleanPath := filepath.Clean(path)
	
	// Check if output file already exists
	if _, err := os.Stat(cleanPath); err == nil {
		if !overwrite {
			return fmt.Errorf("output file already exists: %s (use --overwrite to overwrite)", cleanPath)
		}
	}
	
	// Check if output directory exists and is writable
	outputDir := filepath.Dir(cleanPath)
	if err := fh.ValidateOutputDirectory(outputDir); err != nil {
		return fmt.Errorf("output directory validation failed: %w", err)
	}
	
	return nil
}

// ValidateOutputDirectory checks if the output directory exists and is writable
func (fh *FileHandler) ValidateOutputDirectory(dir string) error {
	if dir == "" || dir == "." {
		// Current directory - check if writable
		return fh.testWritePermission(".")
	}
	
	// Clean the directory path
	cleanDir := filepath.Clean(dir)
	
	info, err := os.Stat(cleanDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %s", cleanDir)
	}
	if err != nil {
		return fmt.Errorf("cannot access output directory %s: %w", cleanDir, err)
	}
	
	if !info.IsDir() {
		return fmt.Errorf("output path is not a directory: %s", cleanDir)
	}
	
	// Test write permissions
	return fh.testWritePermission(cleanDir)
}

// testWritePermission tests if we can write to a directory
func (fh *FileHandler) testWritePermission(dir string) error {
	// Create a temporary file to test write permissions
	tempFile := filepath.Join(dir, ".csv-h3-tool-write-test")
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("cannot write to directory %s: %w", dir, err)
	}
	file.Close()
	
	// Clean up the test file
	if err := os.Remove(tempFile); err != nil {
		// Log warning but don't fail - the main operation succeeded
		fmt.Fprintf(os.Stderr, "Warning: could not remove test file %s: %v\n", tempFile, err)
	}
	
	return nil
}

// GenerateOutputPath creates a default output file path based on input file
func (fh *FileHandler) GenerateOutputPath(inputFile string, suffix string) string {
	if inputFile == "" {
		return fmt.Sprintf("output%s.csv", suffix)
	}
	
	// Clean the input path
	cleanInput := filepath.Clean(inputFile)
	
	ext := filepath.Ext(cleanInput)
	base := strings.TrimSuffix(filepath.Base(cleanInput), ext)
	dir := filepath.Dir(cleanInput)
	
	return filepath.Join(dir, fmt.Sprintf("%s%s%s", base, suffix, ext))
}

// EnsureCSVExtension ensures the file has a .csv extension
func (fh *FileHandler) EnsureCSVExtension(path string) string {
	if path == "" {
		return path
	}
	
	cleanPath := filepath.Clean(path)
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".csv") {
		return cleanPath + ".csv"
	}
	
	return cleanPath
}

// GetAbsolutePath returns the absolute path for a given file path
func (fh *FileHandler) GetAbsolutePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot get absolute path for %s: %w", path, err)
	}
	
	return absPath, nil
}

// GetFileSize returns the size of a file in bytes
func (fh *FileHandler) GetFileSize(path string) (int64, error) {
	if path == "" {
		return 0, fmt.Errorf("path cannot be empty")
	}
	
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("cannot get file info for %s: %w", path, err)
	}
	
	return info.Size(), nil
}

// IsCSVFile checks if a file has a CSV extension
func (fh *FileHandler) IsCSVFile(path string) bool {
	if path == "" {
		return false
	}
	
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".csv"
}

// CreateBackup creates a backup of an existing file
func (fh *FileHandler) CreateBackup(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	
	// Check if original file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("original file does not exist: %s", path)
	}
	
	// Generate backup filename
	backupPath := path + ".backup"
	counter := 1
	
	// Find an available backup filename
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		}
		backupPath = fmt.Sprintf("%s.backup.%d", path, counter)
		counter++
	}
	
	// Copy the file
	if err := fh.copyFile(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	
	return backupPath, nil
}

// copyFile copies a file from src to dst
func (fh *FileHandler) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	// Copy file contents
	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := sourceFile.Read(buffer)
		if n > 0 {
			if _, writeErr := destFile.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
	}
	
	return nil
}

// CleanPath cleans and normalizes a file path
func (fh *FileHandler) CleanPath(path string) string {
	if path == "" {
		return path
	}
	
	return filepath.Clean(path)
}

// SanitizeFilename removes or replaces invalid characters in a filename
func (fh *FileHandler) SanitizeFilename(filename string) string {
	if filename == "" {
		return filename
	}
	
	// Replace invalid characters with underscores
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	sanitized := filename
	
	for _, char := range invalidChars {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure filename is not empty or only underscores after sanitization
	if sanitized == "" || strings.Trim(sanitized, "_") == "" {
		sanitized = "output"
	}
	
	return sanitized
}