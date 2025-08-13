package filehandler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileHandler(t *testing.T) {
	fh := NewFileHandler()
	if fh == nil {
		t.Fatal("Expected FileHandler instance, got nil")
	}
}

func TestFileHandler_ValidateInputFile(t *testing.T) {
	fh := NewFileHandler()
	
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_input_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "non-existent file",
			path:        "/path/to/nonexistent/file.csv",
			expectError: true,
		},
		{
			name:        "valid existing file",
			path:        tempFile.Name(),
			expectError: false,
		},
		{
			name:        "directory instead of file",
			path:        tempDir,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fh.ValidateInputFile(tt.path)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestFileHandler_ValidateOutputFile(t *testing.T) {
	fh := NewFileHandler()
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create an existing file
	existingFile := filepath.Join(tempDir, "existing.csv")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}
	
	tests := []struct {
		name        string
		path        string
		overwrite   bool
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			overwrite:   false,
			expectError: true,
		},
		{
			name:        "new file in existing directory",
			path:        filepath.Join(tempDir, "new.csv"),
			overwrite:   false,
			expectError: false,
		},
		{
			name:        "existing file without overwrite",
			path:        existingFile,
			overwrite:   false,
			expectError: true,
		},
		{
			name:        "existing file with overwrite",
			path:        existingFile,
			overwrite:   true,
			expectError: false,
		},
		{
			name:        "file in non-existent directory",
			path:        "/path/to/nonexistent/dir/output.csv",
			overwrite:   false,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fh.ValidateOutputFile(tt.path, tt.overwrite)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestFileHandler_ValidateOutputDirectory(t *testing.T) {
	fh := NewFileHandler()
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a file (not directory) for testing
	tempFile := filepath.Join(tempDir, "notdir.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	tests := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{
			name:        "current directory",
			dir:         ".",
			expectError: false,
		},
		{
			name:        "empty directory",
			dir:         "",
			expectError: false,
		},
		{
			name:        "valid existing directory",
			dir:         tempDir,
			expectError: false,
		},
		{
			name:        "non-existent directory",
			dir:         "/path/to/nonexistent/dir",
			expectError: true,
		},
		{
			name:        "file instead of directory",
			dir:         tempFile,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fh.ValidateOutputDirectory(tt.dir)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestFileHandler_GenerateOutputPath(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name      string
		inputFile string
		suffix    string
		expected  string
	}{
		{
			name:      "empty input file",
			inputFile: "",
			suffix:    "_with_h3",
			expected:  "output_with_h3.csv",
		},
		{
			name:      "simple filename",
			inputFile: "data.csv",
			suffix:    "_with_h3",
			expected:  "data_with_h3.csv",
		},
		{
			name:      "filename with path",
			inputFile: filepath.Join("path", "to", "data.csv"),
			suffix:    "_with_h3",
			expected:  filepath.Join("path", "to", "data_with_h3.csv"),
		},
		{
			name:      "filename without extension",
			inputFile: "data",
			suffix:    "_with_h3",
			expected:  "data_with_h3",
		},
		{
			name:      "filename with different extension",
			inputFile: "data.txt",
			suffix:    "_processed",
			expected:  "data_processed.txt",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.GenerateOutputPath(tt.inputFile, tt.suffix)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFileHandler_EnsureCSVExtension(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "already has csv extension",
			path:     "data.csv",
			expected: "data.csv",
		},
		{
			name:     "has CSV extension uppercase",
			path:     "data.CSV",
			expected: "data.CSV",
		},
		{
			name:     "no extension",
			path:     "data",
			expected: "data.csv",
		},
		{
			name:     "different extension",
			path:     "data.txt",
			expected: "data.txt.csv",
		},
		{
			name:     "path with directory",
			path:     filepath.Join("path", "to", "data"),
			expected: filepath.Join("path", "to", "data.csv"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.EnsureCSVExtension(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFileHandler_GetAbsolutePath(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "relative path",
			path:        "data.csv",
			expectError: false,
		},
		{
			name:        "current directory",
			path:        ".",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fh.GetAbsolutePath(tt.path)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && !filepath.IsAbs(result) {
				t.Errorf("Expected absolute path, got: %s", result)
			}
		})
	}
}

func TestFileHandler_GetFileSize(t *testing.T) {
	fh := NewFileHandler()
	
	// Create a temporary file with known content
	tempFile, err := os.CreateTemp("", "test_size_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	testContent := "test content"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()
	
	tests := []struct {
		name        string
		path        string
		expectedSize int64
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			expectedSize: 0,
			expectError: true,
		},
		{
			name:        "non-existent file",
			path:        "/path/to/nonexistent.csv",
			expectedSize: 0,
			expectError: true,
		},
		{
			name:        "valid file",
			path:        tempFile.Name(),
			expectedSize: int64(len(testContent)),
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := fh.GetFileSize(tt.path)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && size != tt.expectedSize {
				t.Errorf("Expected size %d, got %d", tt.expectedSize, size)
			}
		})
	}
}

func TestFileHandler_IsCSVFile(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "csv file",
			path:     "data.csv",
			expected: true,
		},
		{
			name:     "CSV file uppercase",
			path:     "data.CSV",
			expected: true,
		},
		{
			name:     "txt file",
			path:     "data.txt",
			expected: false,
		},
		{
			name:     "no extension",
			path:     "data",
			expected: false,
		},
		{
			name:     "csv in path but not extension",
			path:     "csv/data.txt",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.IsCSVFile(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestFileHandler_CreateBackup(t *testing.T) {
	fh := NewFileHandler()
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_backup_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a test file
	testFile := filepath.Join(tempDir, "test.csv")
	testContent := "test,content\n1,2\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "non-existent file",
			path:        "/path/to/nonexistent.csv",
			expectError: true,
		},
		{
			name:        "valid file",
			path:        testFile,
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupPath, err := fh.CreateBackup(tt.path)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError {
				// Check that backup file exists
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("Backup file was not created: %s", backupPath)
				}
				
				// Check that backup has same content
				backupContent, err := os.ReadFile(backupPath)
				if err != nil {
					t.Errorf("Failed to read backup file: %v", err)
				} else if string(backupContent) != testContent {
					t.Errorf("Backup content doesn't match original")
				}
				
				// Clean up backup file
				os.Remove(backupPath)
			}
		})
	}
}

func TestFileHandler_CleanPath(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "simple path",
			path:     "data.csv",
			expected: "data.csv",
		},
		{
			name:     "path with double slashes",
			path:     "path//to//data.csv",
			expected: filepath.Join("path", "to", "data.csv"),
		},
		{
			name:     "path with current directory",
			path:     "./data.csv",
			expected: "data.csv",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.CleanPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFileHandler_SanitizeFilename(t *testing.T) {
	fh := NewFileHandler()
	
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "valid filename",
			filename: "data.csv",
			expected: "data.csv",
		},
		{
			name:     "filename with invalid characters",
			filename: "data<>:\"|?*.csv",
			expected: "data_______.csv",
		},
		{
			name:     "filename with spaces and dots",
			filename: " .data.csv. ",
			expected: "data.csv",
		},
		{
			name:     "only invalid characters",
			filename: "<>:\"|?*",
			expected: "output",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fh.SanitizeFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}