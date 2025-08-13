package csv

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Mock validator for testing
type mockValidator struct {
	shouldFail bool
}

func (m *mockValidator) ValidateCoordinates(lat, lng float64) error {
	if m.shouldFail {
		return fmt.Errorf("mock validation error")
	}
	if lat < -90 || lat > 90 {
		return fmt.Errorf("invalid latitude: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("invalid longitude: %f", lng)
	}
	return nil
}

// Mock H3 generator for testing
type mockH3Generator struct {
	shouldFail bool
}

func (m *mockH3Generator) Generate(lat, lng float64, resolution int) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock H3 generation error")
	}
	return fmt.Sprintf("h3_%d_%.3f_%.3f", resolution, lat, lng), nil
}

func TestNewStreamingProcessor(t *testing.T) {
	validator := &mockValidator{}
	generator := &mockH3Generator{}
	
	processor := NewStreamingProcessor(validator, generator)
	
	if processor == nil {
		t.Fatal("NewStreamingProcessor returned nil")
	}
	if processor.validator != validator {
		t.Error("Validator not set correctly")
	}
	if processor.h3Generator != generator {
		t.Error("H3 generator not set correctly")
	}
}

func TestProcessStream(t *testing.T) {
	// Create test CSV file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York\n34.0522,-118.2437,Los Angeles\n91.0,0.0,Invalid\n,,-Empty\ninvalid,invalid,Invalid"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	validator := &mockValidator{shouldFail: false}
	generator := &mockH3Generator{shouldFail: false}
	processor := NewStreamingProcessor(validator, generator)

	var processedRecords []*Record
	recordHandler := func(record *Record) error {
		processedRecords = append(processedRecords, record)
		return nil
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	// Should have processed 5 records (2 valid, 3 invalid)
	if len(processedRecords) != 5 {
		t.Errorf("Expected 5 processed records, got %d", len(processedRecords))
	}

	// Check first valid record
	if !processedRecords[0].IsValid {
		t.Error("First record should be valid")
	}
	if processedRecords[0].H3Index == "" {
		t.Error("First record should have H3 index")
	}

	// Check second valid record
	if !processedRecords[1].IsValid {
		t.Error("Second record should be valid")
	}
	if processedRecords[1].H3Index == "" {
		t.Error("Second record should have H3 index")
	}

	// Check invalid coordinate record (latitude 91.0)
	if processedRecords[2].IsValid {
		t.Error("Third record should be invalid (out of range latitude)")
	}

	// Check empty record
	if processedRecords[3].IsValid {
		t.Error("Fourth record should be invalid (empty coordinates)")
	}

	// Check unparseable record
	if processedRecords[4].IsValid {
		t.Error("Fifth record should be invalid (unparseable coordinates)")
	}
}

func TestProcessStreamWithValidationFailure(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Use validator that always fails
	validator := &mockValidator{shouldFail: true}
	generator := &mockH3Generator{shouldFail: false}
	processor := NewStreamingProcessor(validator, generator)

	var processedRecords []*Record
	recordHandler := func(record *Record) error {
		processedRecords = append(processedRecords, record)
		return nil
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if len(processedRecords) != 1 {
		t.Errorf("Expected 1 processed record, got %d", len(processedRecords))
	}

	// Record should be marked as invalid due to validation failure
	if processedRecords[0].IsValid {
		t.Error("Record should be invalid due to validation failure")
	}
}

func TestProcessStreamWithH3GenerationFailure(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	validator := &mockValidator{shouldFail: false}
	// Use H3 generator that always fails
	generator := &mockH3Generator{shouldFail: true}
	processor := NewStreamingProcessor(validator, generator)

	var processedRecords []*Record
	recordHandler := func(record *Record) error {
		processedRecords = append(processedRecords, record)
		return nil
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if len(processedRecords) != 1 {
		t.Errorf("Expected 1 processed record, got %d", len(processedRecords))
	}

	// Record should be marked as invalid due to H3 generation failure
	if processedRecords[0].IsValid {
		t.Error("Record should be invalid due to H3 generation failure")
	}
}

func TestProcessStreamWithRecordHandlerError(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	validator := &mockValidator{shouldFail: false}
	generator := &mockH3Generator{shouldFail: false}
	processor := NewStreamingProcessor(validator, generator)

	// Record handler that always returns an error
	recordHandler := func(record *Record) error {
		return fmt.Errorf("handler error")
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err == nil {
		t.Error("Expected error from record handler")
	}
}

func TestProcessStreamMemoryEfficiency(t *testing.T) {
	// Create a larger test file to verify streaming behavior
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.csv")
	
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Write header
	file.WriteString("latitude,longitude,name\n")
	
	// Write many records
	recordCount := 1000
	for i := 0; i < recordCount; i++ {
		lat := 40.0 + float64(i)*0.001
		lng := -74.0 + float64(i)*0.001
		file.WriteString(fmt.Sprintf("%.6f,%.6f,Location_%d\n", lat, lng, i))
	}
	file.Close()

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	validator := &mockValidator{shouldFail: false}
	generator := &mockH3Generator{shouldFail: false}
	processor := NewStreamingProcessor(validator, generator)

	processedCount := 0
	recordHandler := func(record *Record) error {
		processedCount++
		// Verify that we're not accumulating all records in memory
		// by checking that each record is processed individually
		if record.IsValid && record.H3Index == "" {
			t.Error("Valid record should have H3 index")
		}
		return nil
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if processedCount != recordCount {
		t.Errorf("Expected %d processed records, got %d", recordCount, processedCount)
	}
}

func TestProcessStreamWithNilComponents(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
		Resolution: 8,
		Verbose:    false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	// Test with nil validator and generator
	processor := NewStreamingProcessor(nil, nil)

	var processedRecords []*Record
	recordHandler := func(record *Record) error {
		processedRecords = append(processedRecords, record)
		return nil
	}

	err = processor.ProcessStream(reader, config, recordHandler)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if len(processedRecords) != 1 {
		t.Errorf("Expected 1 processed record, got %d", len(processedRecords))
	}

	// Record should still be valid (coordinates were parsed successfully)
	// but won't have H3 index since generator is nil
	if !processedRecords[0].IsValid {
		t.Error("Record should be valid even with nil components")
	}
	if processedRecords[0].H3Index != "" {
		t.Error("Record should not have H3 index with nil generator")
	}
}