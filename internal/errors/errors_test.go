package errors

import (
	"fmt"
	"testing"
)

// TestFileError tests FileError creation and formatting
func TestFileError(t *testing.T) {
	cause := fmt.Errorf("permission denied")
	err := NewFileError("/path/to/file.csv", "read", cause)

	if err.Path != "/path/to/file.csv" {
		t.Errorf("Expected path '/path/to/file.csv', got '%s'", err.Path)
	}

	if err.Operation != "read" {
		t.Errorf("Expected operation 'read', got '%s'", err.Operation)
	}

	expectedMsg := "FILE: read operation failed for '/path/to/file.csv': permission denied"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap should return the original cause")
	}
}

// TestCSVError tests CSVError creation and formatting
func TestCSVError(t *testing.T) {
	cause := fmt.Errorf("invalid format")
	err := NewCSVError("test.csv", 5, 2, "latitude", "invalid", "coordinate parsing failed", cause)

	expectedMsg := "CSV file 'test.csv' line 5 column 2 field 'latitude' value 'invalid': coordinate parsing failed (caused by: invalid format)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestCoordinateError tests CoordinateError creation and formatting
func TestCoordinateError(t *testing.T) {
	err := NewCoordinateError(91.0, -74.0, 10, "latitude", "latitude out of range")

	expectedMsg := "COORDINATE at line 10 field 'latitude': latitude out of range (lat: 91.000000, lng: -74.000000)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestH3Error tests H3Error creation and formatting
func TestH3Error(t *testing.T) {
	cause := fmt.Errorf("H3 library error")
	err := NewH3Error(40.7128, -74.0060, 8, 15, "H3 index generation failed", cause)

	expectedMsg := "H3 at line 15: H3 index generation failed (lat: 40.712800, lng: -74.006000, resolution: 8) - H3 library error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestConfigError tests ConfigError creation and formatting
func TestConfigError(t *testing.T) {
	cause := fmt.Errorf("invalid value")
	err := NewConfigError("resolution", "16", "resolution out of range", cause)

	expectedMsg := "CONFIG field 'resolution' (value: '16'): resolution out of range - invalid value"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestValidationError tests ValidationError creation and formatting
func TestValidationError(t *testing.T) {
	cause := fmt.Errorf("validation failed")
	err := NewValidationError("coordinates", "invalid", 20, "coordinate validation failed", cause)

	expectedMsg := "VALIDATION line 20 field 'coordinates' value 'invalid': coordinate validation failed - validation failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestProcessingError tests ProcessingError creation and formatting
func TestProcessingError(t *testing.T) {
	cause := fmt.Errorf("processing failed")
	err := NewProcessingError("csv_parsing", 25, "failed to parse CSV record", cause)

	expectedMsg := "PROCESSING stage 'csv_parsing' line 25: failed to parse CSV record - processing failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestErrorCollector tests ErrorCollector functionality
func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector(3)

	if collector.HasErrors() {
		t.Error("New collector should not have errors")
	}

	if collector.Count() != 0 {
		t.Errorf("Expected count 0, got %d", collector.Count())
	}

	// Add errors
	err1 := NewFileError("file1.csv", "read", fmt.Errorf("error1"))
	err2 := NewFileError("file2.csv", "write", fmt.Errorf("error2"))
	err3 := NewFileError("file3.csv", "delete", fmt.Errorf("error3"))
	err4 := NewFileError("file4.csv", "create", fmt.Errorf("error4"))

	collector.Add(err1)
	collector.Add(err2)
	collector.Add(err3)
	collector.Add(err4) // Should be ignored due to limit

	if !collector.HasErrors() {
		t.Error("Collector should have errors")
	}

	if collector.Count() != 3 {
		t.Errorf("Expected count 3, got %d", collector.Count())
	}

	errors := collector.Errors()
	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errors))
	}
}

// TestIsErrorType tests error type checking
func TestIsErrorType(t *testing.T) {
	fileErr := NewFileError("test.csv", "read", fmt.Errorf("error"))
	csvErr := NewCSVError("test.csv", 1, 1, "field", "value", "message", nil)

	if !IsErrorType(fileErr, ErrorTypeFile) {
		t.Error("Should identify FileError as FILE type")
	}

	if IsErrorType(fileErr, ErrorTypeCSV) {
		t.Error("Should not identify FileError as CSV type")
	}

	if !IsErrorType(csvErr, ErrorTypeCSV) {
		t.Error("Should identify CSVError as CSV type")
	}

	if IsErrorType(nil, ErrorTypeFile) {
		t.Error("Should not identify nil as any error type")
	}
}

// TestBaseErrorContext tests context functionality
func TestBaseErrorContext(t *testing.T) {
	err := &BaseError{
		Type:    ErrorTypeFile,
		Message: "test error",
	}

	// Test adding context
	err.WithContext("key1", "value1")
	err.WithContext("key2", 42)

	value1, exists1 := err.GetContext("key1")
	if !exists1 || value1 != "value1" {
		t.Error("Should retrieve context value1")
	}

	value2, exists2 := err.GetContext("key2")
	if !exists2 || value2 != 42 {
		t.Error("Should retrieve context value2")
	}

	_, exists3 := err.GetContext("nonexistent")
	if exists3 {
		t.Error("Should not find nonexistent context key")
	}
}