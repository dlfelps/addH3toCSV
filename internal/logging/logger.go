package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"csv-h3-tool/internal/errors"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level      LogLevel
	output     io.Writer
	prefix     string
	verbose    bool
	errorCount int
	warnCount  int
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel, output io.Writer, verbose bool) *Logger {
	if output == nil {
		output = os.Stderr
	}
	
	return &Logger{
		level:   level,
		output:  output,
		verbose: verbose,
	}
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger(verbose bool) *Logger {
	level := LogLevelInfo
	if verbose {
		level = LogLevelDebug
	}
	return NewLogger(level, os.Stderr, verbose)
}

// SetLevel sets the minimum logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetPrefix sets a prefix for all log messages
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}
	return fmt.Sprintf("%s %s%s: %s", timestamp, prefix, level.String(), message)
}

// log writes a message at the specified level
func (l *Logger) log(level LogLevel, message string) {
	if !l.shouldLog(level) {
		return
	}
	
	formatted := l.formatMessage(level, message)
	fmt.Fprintln(l.output, formatted)
	
	// Update counters
	switch level {
	case LogLevelError, LogLevelFatal:
		l.errorCount++
	case LogLevelWarn:
		l.warnCount++
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(LogLevelDebug, message)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(LogLevelInfo, message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(LogLevelWarn, message)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(LogLevelError, message)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	l.log(LogLevelFatal, message)
	os.Exit(1)
}

// LogError logs an error with detailed context
func (l *Logger) LogError(err error) {
	if err == nil {
		return
	}
	
	// Extract additional context for structured errors
	switch e := err.(type) {
	case *errors.CSVError:
		l.Error("CSV processing error: %s", e.Error())
		if l.verbose && e.Context != nil {
			for key, value := range e.Context {
				l.Debug("  %s: %v", key, value)
			}
		}
	case *errors.CoordinateError:
		l.Error("Coordinate validation error: %s", e.Error())
	case *errors.H3Error:
		l.Error("H3 generation error: %s", e.Error())
	case *errors.FileError:
		l.Error("File operation error: %s", e.Error())
	case *errors.ConfigError:
		l.Error("Configuration error: %s", e.Error())
	case *errors.ValidationError:
		l.Error("Validation error: %s", e.Error())
	case *errors.ProcessingError:
		l.Error("Processing error: %s", e.Error())
	default:
		l.Error("Error: %v", err)
	}
}

// LogSkippedRecord logs information about a skipped record
func (l *Logger) LogSkippedRecord(line int, reason string, details ...string) {
	message := fmt.Sprintf("Skipping record at line %d: %s", line, reason)
	if len(details) > 0 && l.verbose {
		message += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
	}
	l.Warn(message)
}

// LogProcessingProgress logs processing progress
func (l *Logger) LogProcessingProgress(processed, total int, stage string) {
	if !l.verbose {
		return
	}
	
	percentage := float64(processed) / float64(total) * 100
	l.Info("Processing progress: %d/%d (%.1f%%) - %s", processed, total, percentage, stage)
}

// LogProcessingSummary logs a summary of processing results
func (l *Logger) LogProcessingSummary(total, valid, invalid int, duration time.Duration) {
	l.Info("Processing completed in %v", duration)
	l.Info("Total records: %d", total)
	l.Info("Valid records: %d", valid)
	if invalid > 0 {
		l.Warn("Invalid records: %d", invalid)
	}
	
	if l.errorCount > 0 || l.warnCount > 0 {
		l.Info("Summary: %d errors, %d warnings", l.errorCount, l.warnCount)
	}
}

// GetErrorCount returns the number of errors logged
func (l *Logger) GetErrorCount() int {
	return l.errorCount
}

// GetWarnCount returns the number of warnings logged
func (l *Logger) GetWarnCount() int {
	return l.warnCount
}

// Reset resets the error and warning counters
func (l *Logger) Reset() {
	l.errorCount = 0
	l.warnCount = 0
}

// WithPrefix creates a new logger with the specified prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	newLogger := *l
	newLogger.prefix = prefix
	return &newLogger
}

// ProcessingLogger provides specialized logging for CSV processing
type ProcessingLogger struct {
	*Logger
	fileName     string
	totalRecords int
	processed    int
	lastReported time.Time
	reportInterval time.Duration
}

// NewProcessingLogger creates a logger specialized for processing operations
func NewProcessingLogger(logger *Logger, fileName string, totalRecords int) *ProcessingLogger {
	return &ProcessingLogger{
		Logger:         logger,
		fileName:       fileName,
		totalRecords:   totalRecords,
		processed:      0,
		lastReported:   time.Now(),
		reportInterval: 2 * time.Second,
	}
}

// LogRecordProcessed logs that a record has been processed
func (pl *ProcessingLogger) LogRecordProcessed(line int, valid bool, h3Index string) {
	pl.processed++
	
	if pl.verbose {
		if valid {
			pl.Debug("Line %d: Generated H3 index %s", line, h3Index)
		} else {
			pl.Debug("Line %d: Skipped invalid record", line)
		}
	}
	
	// Report progress periodically
	now := time.Now()
	if now.Sub(pl.lastReported) >= pl.reportInterval {
		pl.LogProcessingProgress(pl.processed, pl.totalRecords, "processing records")
		pl.lastReported = now
	}
}

// LogCoordinateError logs a coordinate validation error with context
func (pl *ProcessingLogger) LogCoordinateError(line int, lat, lng float64, field, reason string) {
	err := errors.NewCoordinateError(lat, lng, line, field, reason)
	pl.LogError(err)
}

// LogH3Error logs an H3 generation error with context
func (pl *ProcessingLogger) LogH3Error(line int, lat, lng float64, resolution int, reason string, cause error) {
	err := errors.NewH3Error(lat, lng, resolution, line, reason, cause)
	pl.LogError(err)
}

// LogCSVError logs a CSV parsing error with context
func (pl *ProcessingLogger) LogCSVError(line, column int, field, value, reason string, cause error) {
	err := errors.NewCSVError(pl.fileName, line, column, field, value, reason, cause)
	pl.LogError(err)
}

// Complete logs completion of processing
func (pl *ProcessingLogger) Complete(duration time.Duration, validCount, invalidCount int) {
	pl.LogProcessingSummary(pl.processed, validCount, invalidCount, duration)
}

// Global logger instance
var defaultLogger *Logger

// InitDefaultLogger initializes the global logger
func InitDefaultLogger(verbose bool) {
	defaultLogger = NewDefaultLogger(verbose)
}

// GetDefaultLogger returns the global logger instance
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewDefaultLogger(false)
	}
	return defaultLogger
}

// Convenience functions for global logger
func Debug(message string, args ...interface{}) {
	GetDefaultLogger().Debug(message, args...)
}

func Info(message string, args ...interface{}) {
	GetDefaultLogger().Info(message, args...)
}

func Warn(message string, args ...interface{}) {
	GetDefaultLogger().Warn(message, args...)
}

func Error(message string, args ...interface{}) {
	GetDefaultLogger().Error(message, args...)
}

func Fatal(message string, args ...interface{}) {
	GetDefaultLogger().Fatal(message, args...)
}

func LogError(err error) {
	GetDefaultLogger().LogError(err)
}