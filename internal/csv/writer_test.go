package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	inputHeaders := []string{"latitude", "longitude", "name"}
	config := Config{
		HasHeaders: true,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, inputHeaders, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	defer writer.Close()

	expectedHeaders := []string{"latitude", "longitude", "name", "h3_index"}
	if len(writer.headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(writer.headers))
	}
	for i, expected := range expectedHeaders {
		if writer.headers[i] != expected {
			t.Errorf("Expected header %d to be %s, got %s", i, expected, writer.headers[i])
		}
	}
}

func TestNewWriterWithoutHeaders(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	config := Config{
		HasHeaders: false,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, nil, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	defer writer.Close()

	if writer.headers != nil {
		t.Error("Expected nil headers for file without headers")
	}
}

func TestNewWriterOverwriteProtection(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "existing.csv")

	// Create existing file
	if err := os.WriteFile(outputFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	config := Config{
		HasHeaders: true,
		Overwrite:  false, // Don't allow overwrite
	}

	_, err := NewWriter(outputFile, []string{"test"}, config)
	if err == nil {
		t.Error("Expected error when trying to overwrite existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestNewWriterWithOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "existing.csv")

	// Create existing file
	if err := os.WriteFile(outputFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	config := Config{
		HasHeaders: true,
		Overwrite:  true, // Allow overwrite
	}

	writer, err := NewWriter(outputFile, []string{"test"}, config)
	if err != nil {
		t.Fatalf("NewWriter failed with overwrite enabled: %v", err)
	}
	defer writer.Close()
}

func TestWriteRecord(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	inputHeaders := []string{"latitude", "longitude", "name"}
	config := Config{
		HasHeaders: true,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, inputHeaders, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	// Write valid record
	validRecord := &Record{
		OriginalData: []string{"40.7128", "-74.0060", "New York"},
		Latitude:     40.7128,
		Longitude:    -74.0060,
		H3Index:      "8a2a1072b59ffff",
		IsValid:      true,
	}

	if err := writer.WriteRecord(validRecord); err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Write invalid record
	invalidRecord := &Record{
		OriginalData: []string{"invalid", "invalid", "Invalid"},
		IsValid:      false,
	}

	if err := writer.WriteRecord(invalidRecord); err != nil {
		t.Fatalf("WriteRecord failed for invalid record: %v", err)
	}

	writer.Close()

	// Read and verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 lines in output, got %d", len(lines))
	}

	// Check header
	expectedHeader := "latitude,longitude,name,h3_index"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header %s, got %s", expectedHeader, lines[0])
	}

	// Check valid record
	expectedValidRow := "40.7128,-74.0060,New York,8a2a1072b59ffff"
	if lines[1] != expectedValidRow {
		t.Errorf("Expected valid row %s, got %s", expectedValidRow, lines[1])
	}

	// Check invalid record (should have empty H3 index)
	expectedInvalidRow := "invalid,invalid,Invalid,"
	if lines[2] != expectedInvalidRow {
		t.Errorf("Expected invalid row %s, got %s", expectedInvalidRow, lines[2])
	}
}

func TestWriteRecords(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	inputHeaders := []string{"lat", "lng"}
	config := Config{
		HasHeaders: true,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, inputHeaders, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	records := []*Record{
		{
			OriginalData: []string{"40.7128", "-74.0060"},
			H3Index:      "8a2a1072b59ffff",
			IsValid:      true,
		},
		{
			OriginalData: []string{"34.0522", "-118.2437"},
			H3Index:      "8a2a100725bffff",
			IsValid:      true,
		},
	}

	if err := writer.WriteRecords(records); err != nil {
		t.Fatalf("WriteRecords failed: %v", err)
	}

	writer.Close()

	// Verify output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 lines in output, got %d", len(lines))
	}
}

func TestWriteRecordNil(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	config := Config{
		HasHeaders: false,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, nil, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	defer writer.Close()

	err = writer.WriteRecord(nil)
	if err == nil {
		t.Error("Expected error when writing nil record")
	}
	if !strings.Contains(err.Error(), "record is nil") {
		t.Errorf("Expected 'record is nil' error, got: %v", err)
	}
}

func TestWriterFlush(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.csv")

	config := Config{
		HasHeaders: false,
		Overwrite:  true,
	}

	writer, err := NewWriter(outputFile, nil, config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	defer writer.Close()

	record := &Record{
		OriginalData: []string{"40.7128", "-74.0060"},
		H3Index:      "8a2a1072b59ffff",
		IsValid:      true,
	}

	if err := writer.WriteRecord(record); err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Flush explicitly
	if err := writer.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Verify data was written
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected content in output file after flush")
	}
}

func TestWriterIntegrationWithReader(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.csv")
	outputFile := filepath.Join(tempDir, "output.csv")

	// Create input CSV
	inputContent := "latitude,longitude,name\n40.7128,-74.0060,New York\n34.0522,-118.2437,Los Angeles"
	if err := os.WriteFile(inputFile, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Overwrite:  true,
	}

	// Read from input
	reader, err := NewReader(inputFile, config)
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer reader.Close()

	// Create writer
	writer, err := NewWriter(outputFile, reader.GetHeaders(), config)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	defer writer.Close()

	// Process records
	for {
		record, err := reader.ReadRecord()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("ReadRecord failed: %v", err)
		}

		// Add mock H3 index for valid records
		if record.IsValid {
			record.H3Index = "mock_h3_index"
		}

		if err := writer.WriteRecord(record); err != nil {
			t.Fatalf("WriteRecord failed: %v", err)
		}
	}

	writer.Close()

	// Verify output
	outputContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(outputContent)), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Errorf("Expected 3 lines in output, got %d", len(lines))
	}

	// Check that H3 index column was added
	expectedHeader := "latitude,longitude,name,h3_index"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header %s, got %s", expectedHeader, lines[0])
	}

	// Check that H3 index values were added
	if !strings.Contains(lines[1], "mock_h3_index") {
		t.Error("Expected H3 index in first data row")
	}
	if !strings.Contains(lines[2], "mock_h3_index") {
		t.Error("Expected H3 index in second data row")
	}
}

func TestWriterFileCreationError(t *testing.T) {
	// Try to create writer in non-existent directory
	invalidPath := "/nonexistent/directory/output.csv"
	
	config := Config{
		HasHeaders: true,
		Overwrite:  true,
	}

	_, err := NewWriter(invalidPath, []string{"test"}, config)
	if err == nil {
		t.Error("Expected error when creating file in non-existent directory")
	}
}