package logging

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"csv-h3-tool/internal/errors"
)

// TestLogger_BasicLogging tests basic logging functionality
func TestLogger_BasicLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf, true)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	// Check that each line contains the expected level
	if !strings.Contains(lines[0], "DEBUG") {
		t.Error("First line should contain DEBUG")
	}
	if !strings.Contains(lines[1], "INFO") {
		t.Error("Second line should contain INFO")
	}
	if !strings.Contains(lines[2], "WARN") {
		t.Error("Third line should contain WARN")
	}
	if !strings.Contains(lines[3], "ERROR") {
		t.Error("Fourth line should contain ERROR")
	}
}

// TestLogger_LogLevel tests log level filtering
func TestLogger_LogLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelWarn, &buf, false)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have WARN and ERROR messages
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}

	if !strings.Contains(output, "WARN") {
		t.Error("Output should contain WARN message")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Output should contain ERROR message")
	}
	if strings.Contains(output, "DEBUG") || strings.Contains(output, "INFO") {
		t.Error("Output should not contain DEBUG or INFO messages")
	}
}

// TestLogger_Formatting tests message formatting
func TestLogger_Formatting(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf, true)

	logger.Info("formatted message: %s %d", "test", 42)

	output := buf.String()
	if !strings.Contains(output, "formatted message: test 42") {
		t.Error("Message should be properly formatted")
	}
}

// TestLogger_Prefix tests prefix functionality
func TestLogger_Prefix(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf, true)
	logger.SetPrefix("TEST")

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[TEST]") {
		t.Error("Output should contain prefix")
	}
}

// TestLogger_ErrorCounting tests error and warning counting
func TestLogger_ErrorCounting(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf, true)

	logger.Info("info message")
	logger.Warn("warning 1")
	logger.Warn("warning 2")
	logger.Error("error 1")
	logger.Error("error 2")
	logger.Error("error 3")

	if logger.GetWarnCount() != 2 {
		t.Errorf("Expected 2 warnings, got %d", logger.GetWarnCount())
	}

	if logger.GetErrorCount() != 3 {
		t.Errorf("Expected 3 errors, got %d", logger.GetErrorCount())
	}

	// Test reset
	logger.Reset()
	if logger.GetWarnCount() != 0 || logger.GetErrorCount() != 0 {
		t.Error("Reset should clear counters")
	}
}

// TestLogger_LogError tests structured error logging
func TestLogger_LogError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelDebug, &buf, true)

	// Test with different error types
	fileErr := errors.NewFileError("test.csv", "read", nil)
	csvErr := errors.NewCSVError("test.csv", 5, 2, "lat", "invalid", "parsing failed", nil)
	coordErr := errors.NewCoordinateError(91.0, -74.0, 10, "latitude", "out of range")

	logger.LogError(fileErr)
	logger.LogError(csvErr)
	logger.LogError(coordErr)

	output := buf.String()

	if !strings.Contains(output, "File operation error") {
		t.Error("Should log file error")
	}
	if !strings.Contains(output, "CSV processing error") {
		t.Error("Should log CSV error")
	}
	if !strings.Contains(output, "Coordinate validation error") {
		t.Error("Should log coordinate error")
	}
}

// TestLogger_ProcessingSummary tests processing summary logging
func TestLogger_ProcessingSummary(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LogLevelInfo, &buf, true)

	duration := 100 * time.Millisecond
	logger.LogProcessingSummary(100, 85, 15, duration)

	output := buf.String()

	if !strings.Contains(output, "Processing completed") {
		t.Error("Should log processing completion")
	}
	if !strings.Contains(output, "Total records: 100") {
		t.Error("Should log total records")
	}
	if !strings.Contains(output, "Valid records: 85") {
		t.Error("Should log valid records")
	}
	if !strings.Contains(output, "Invalid records: 15") {
		t.Error("Should log invalid records")
	}
}

// TestProcessingLogger tests specialized processing logger
func TestProcessingLogger(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := NewLogger(LogLevelDebug, &buf, true)
	processLogger := NewProcessingLogger(baseLogger, "test.csv", 100)

	// Test record processing logging
	processLogger.LogRecordProcessed(1, true, "882a107289fffff")
	processLogger.LogRecordProcessed(2, false, "")

	output := buf.String()

	if !strings.Contains(output, "Generated H3 index 882a107289fffff") {
		t.Error("Should log successful H3 generation")
	}
	if !strings.Contains(output, "Skipped invalid record") {
		t.Error("Should log skipped record")
	}
}

// TestProcessingLogger_ErrorLogging tests error logging in processing logger
func TestProcessingLogger_ErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := NewLogger(LogLevelDebug, &buf, true)
	processLogger := NewProcessingLogger(baseLogger, "test.csv", 100)

	// Test coordinate error logging
	processLogger.LogCoordinateError(5, 91.0, -74.0, "latitude", "out of range")

	// Test H3 error logging
	h3Err := errors.NewH3Error(40.7, -74.0, 8, 10, "generation failed", nil)
	processLogger.LogH3Error(10, 40.7, -74.0, 8, "generation failed", h3Err)

	// Test CSV error logging
	processLogger.LogCSVError(15, 2, "latitude", "invalid", "parsing failed", nil)

	output := buf.String()

	if !strings.Contains(output, "Coordinate validation error") {
		t.Error("Should log coordinate error")
	}
	if !strings.Contains(output, "H3 generation error") {
		t.Error("Should log H3 error")
	}
	if !strings.Contains(output, "CSV processing error") {
		t.Error("Should log CSV error")
	}
}

// TestDefaultLogger tests global logger functionality
func TestDefaultLogger(t *testing.T) {
	// Initialize default logger
	InitDefaultLogger(true)

	logger := GetDefaultLogger()
	if logger == nil {
		t.Fatal("Default logger should not be nil")
	}

	if !logger.verbose {
		t.Error("Default logger should be verbose when initialized with true")
	}

	// Test global functions
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// These should not panic
}

// TestLogLevel_String tests LogLevel string representation
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}