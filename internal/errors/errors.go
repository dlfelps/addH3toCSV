package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeFile        ErrorType = "FILE"
	ErrorTypeCSV         ErrorType = "CSV"
	ErrorTypeCoordinate  ErrorType = "COORDINATE"
	ErrorTypeH3          ErrorType = "H3"
	ErrorTypeConfig      ErrorType = "CONFIG"
	ErrorTypeValidation  ErrorType = "VALIDATION"
	ErrorTypeProcessing  ErrorType = "PROCESSING"
)

// BaseError provides common error functionality
type BaseError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

func (e *BaseError) WithContext(key string, value interface{}) *BaseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *BaseError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// FileError represents file operation errors
type FileError struct {
	*BaseError
	Path      string
	Operation string
}

func NewFileError(path, operation string, cause error) *FileError {
	return &FileError{
		BaseError: &BaseError{
			Type:    ErrorTypeFile,
			Message: fmt.Sprintf("file operation '%s' failed for path '%s'", operation, path),
			Cause:   cause,
		},
		Path:      path,
		Operation: operation,
	}
}

func (e *FileError) Error() string {
	return fmt.Sprintf("FILE: %s operation failed for '%s': %v", e.Operation, e.Path, e.Cause)
}

// CSVError represents CSV parsing and processing errors
type CSVError struct {
	*BaseError
	Line     int
	Column   int
	Field    string
	Value    string
	FileName string
}

func NewCSVError(fileName string, line, column int, field, value, message string, cause error) *CSVError {
	return &CSVError{
		BaseError: &BaseError{
			Type:    ErrorTypeCSV,
			Message: message,
			Cause:   cause,
		},
		Line:     line,
		Column:   column,
		Field:    field,
		Value:    value,
		FileName: fileName,
	}
}

func (e *CSVError) Error() string {
	var parts []string
	parts = append(parts, "CSV")
	
	if e.FileName != "" {
		parts = append(parts, fmt.Sprintf("file '%s'", e.FileName))
	}
	
	if e.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d", e.Line))
	}
	
	if e.Column > 0 {
		parts = append(parts, fmt.Sprintf("column %d", e.Column))
	}
	
	if e.Field != "" {
		parts = append(parts, fmt.Sprintf("field '%s'", e.Field))
	}
	
	if e.Value != "" {
		parts = append(parts, fmt.Sprintf("value '%s'", e.Value))
	}
	
	location := strings.Join(parts, " ")
	
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", location, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", location, e.Message)
}

// CoordinateError represents coordinate validation errors
type CoordinateError struct {
	*BaseError
	Latitude  float64
	Longitude float64
	Line      int
	Field     string
}

func NewCoordinateError(lat, lng float64, line int, field, message string) *CoordinateError {
	return &CoordinateError{
		BaseError: &BaseError{
			Type:    ErrorTypeCoordinate,
			Message: message,
		},
		Latitude:  lat,
		Longitude: lng,
		Line:      line,
		Field:     field,
	}
}

func (e *CoordinateError) Error() string {
	var location []string
	if e.Line > 0 {
		location = append(location, fmt.Sprintf("line %d", e.Line))
	}
	if e.Field != "" {
		location = append(location, fmt.Sprintf("field '%s'", e.Field))
	}
	
	locationStr := ""
	if len(location) > 0 {
		locationStr = fmt.Sprintf(" at %s", strings.Join(location, " "))
	}
	
	return fmt.Sprintf("COORDINATE%s: %s (lat: %.6f, lng: %.6f)", locationStr, e.Message, e.Latitude, e.Longitude)
}

// H3Error represents H3 index generation errors
type H3Error struct {
	*BaseError
	Latitude   float64
	Longitude  float64
	Resolution int
	Line       int
}

func NewH3Error(lat, lng float64, resolution, line int, message string, cause error) *H3Error {
	return &H3Error{
		BaseError: &BaseError{
			Type:    ErrorTypeH3,
			Message: message,
			Cause:   cause,
		},
		Latitude:   lat,
		Longitude:  lng,
		Resolution: resolution,
		Line:       line,
	}
}

func (e *H3Error) Error() string {
	location := ""
	if e.Line > 0 {
		location = fmt.Sprintf(" at line %d", e.Line)
	}
	
	if e.Cause != nil {
		return fmt.Sprintf("H3%s: %s (lat: %.6f, lng: %.6f, resolution: %d) - %v", 
			location, e.Message, e.Latitude, e.Longitude, e.Resolution, e.Cause)
	}
	return fmt.Sprintf("H3%s: %s (lat: %.6f, lng: %.6f, resolution: %d)", 
		location, e.Message, e.Latitude, e.Longitude, e.Resolution)
}

// ConfigError represents configuration validation errors
type ConfigError struct {
	*BaseError
	Field string
	Value string
}

func NewConfigError(field, value, message string, cause error) *ConfigError {
	return &ConfigError{
		BaseError: &BaseError{
			Type:    ErrorTypeConfig,
			Message: message,
			Cause:   cause,
		},
		Field: field,
		Value: value,
	}
}

func (e *ConfigError) Error() string {
	fieldInfo := ""
	if e.Field != "" {
		fieldInfo = fmt.Sprintf(" field '%s'", e.Field)
		if e.Value != "" {
			fieldInfo += fmt.Sprintf(" (value: '%s')", e.Value)
		}
	}
	
	if e.Cause != nil {
		return fmt.Sprintf("CONFIG%s: %s - %v", fieldInfo, e.Message, e.Cause)
	}
	return fmt.Sprintf("CONFIG%s: %s", fieldInfo, e.Message)
}

// ValidationError represents general validation errors
type ValidationError struct {
	*BaseError
	Field string
	Value string
	Line  int
}

func NewValidationError(field, value string, line int, message string, cause error) *ValidationError {
	return &ValidationError{
		BaseError: &BaseError{
			Type:    ErrorTypeValidation,
			Message: message,
			Cause:   cause,
		},
		Field: field,
		Value: value,
		Line:  line,
	}
}

func (e *ValidationError) Error() string {
	var parts []string
	parts = append(parts, "VALIDATION")
	
	if e.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d", e.Line))
	}
	
	if e.Field != "" {
		parts = append(parts, fmt.Sprintf("field '%s'", e.Field))
	}
	
	if e.Value != "" {
		parts = append(parts, fmt.Sprintf("value '%s'", e.Value))
	}
	
	location := strings.Join(parts, " ")
	
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s - %v", location, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", location, e.Message)
}

// ProcessingError represents general processing errors
type ProcessingError struct {
	*BaseError
	Stage string
	Line  int
}

func NewProcessingError(stage string, line int, message string, cause error) *ProcessingError {
	return &ProcessingError{
		BaseError: &BaseError{
			Type:    ErrorTypeProcessing,
			Message: message,
			Cause:   cause,
		},
		Stage: stage,
		Line:  line,
	}
}

func (e *ProcessingError) Error() string {
	var parts []string
	parts = append(parts, "PROCESSING")
	
	if e.Stage != "" {
		parts = append(parts, fmt.Sprintf("stage '%s'", e.Stage))
	}
	
	if e.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d", e.Line))
	}
	
	location := strings.Join(parts, " ")
	
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s - %v", location, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", location, e.Message)
}

// ErrorCollector collects multiple errors during processing
type ErrorCollector struct {
	errors []error
	limit  int
}

func NewErrorCollector(limit int) *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
		limit:  limit,
	}
}

func (ec *ErrorCollector) Add(err error) {
	if ec.limit > 0 && len(ec.errors) >= ec.limit {
		return // Don't collect more errors than the limit
	}
	ec.errors = append(ec.errors, err)
}

func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

func (ec *ErrorCollector) Count() int {
	return len(ec.errors)
}

func (ec *ErrorCollector) Errors() []error {
	return ec.errors
}

func (ec *ErrorCollector) Error() string {
	if len(ec.errors) == 0 {
		return "no errors"
	}
	
	if len(ec.errors) == 1 {
		return ec.errors[0].Error()
	}
	
	var messages []string
	for i, err := range ec.errors {
		if i >= 5 { // Show only first 5 errors in summary
			messages = append(messages, fmt.Sprintf("... and %d more errors", len(ec.errors)-i))
			break
		}
		messages = append(messages, err.Error())
	}
	
	return fmt.Sprintf("multiple errors occurred:\n%s", strings.Join(messages, "\n"))
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	if err == nil {
		return false
	}
	
	switch e := err.(type) {
	case *BaseError:
		return e.Type == errorType
	case *FileError:
		return e.Type == errorType
	case *CSVError:
		return e.Type == errorType
	case *CoordinateError:
		return e.Type == errorType
	case *H3Error:
		return e.Type == errorType
	case *ConfigError:
		return e.Type == errorType
	case *ValidationError:
		return e.Type == errorType
	case *ProcessingError:
		return e.Type == errorType
	}
	
	return false
}

// GetErrorContext extracts context information from an error if available
func GetErrorContext(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	
	switch e := err.(type) {
	case *BaseError:
		return e.Context
	case *FileError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *CSVError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *CoordinateError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *H3Error:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *ConfigError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *ValidationError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	case *ProcessingError:
		if e.BaseError != nil {
			return e.BaseError.Context
		}
	}
	
	return nil
}