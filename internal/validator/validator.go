package validator

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   string
	Line    int
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// FileError represents a file operation error
type FileError struct {
	Path      string
	Operation string
	Cause     error
}

func (e FileError) Error() string {
	return e.Cause.Error()
}

// Validator defines the interface for coordinate validation
type Validator interface {
	ValidateCoordinates(lat, lng float64) error
	ParseCoordinate(value string) (float64, error)
}

// CoordinateValidator implements the Validator interface
type CoordinateValidator struct{}

// NewCoordinateValidator creates a new coordinate validator
func NewCoordinateValidator() *CoordinateValidator {
	return &CoordinateValidator{}
}

// ValidateCoordinates validates latitude and longitude values
func (v *CoordinateValidator) ValidateCoordinates(lat, lng float64) error {
	if lat < -90.0 || lat > 90.0 {
		return &ValidationError{
			Field:   "latitude",
			Value:   fmt.Sprintf("%.6f", lat),
			Message: fmt.Sprintf("latitude %.6f is out of range [-90, 90]", lat),
		}
	}
	
	if lng < -180.0 || lng > 180.0 {
		return &ValidationError{
			Field:   "longitude",
			Value:   fmt.Sprintf("%.6f", lng),
			Message: fmt.Sprintf("longitude %.6f is out of range [-180, 180]", lng),
		}
	}
	
	return nil
}

// ParseCoordinate parses a string coordinate value to float64
func (v *CoordinateValidator) ParseCoordinate(value string) (float64, error) {
	// Handle empty or whitespace-only values
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, &ValidationError{
			Field:   "coordinate",
			Value:   value,
			Message: "coordinate value is empty or contains only whitespace",
		}
	}
	
	// Attempt to parse the coordinate
	coord, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, &ValidationError{
			Field:   "coordinate",
			Value:   value,
			Message: fmt.Sprintf("invalid coordinate format: %s", err.Error()),
		}
	}
	
	return coord, nil
}

// ValidateLatitude validates a latitude value specifically
func ValidateLatitude(lat float64) error {
	if lat < -90.0 || lat > 90.0 {
		return &ValidationError{
			Field:   "latitude",
			Value:   fmt.Sprintf("%.6f", lat),
			Message: fmt.Sprintf("latitude %.6f is out of range [-90, 90]", lat),
		}
	}
	return nil
}

// ValidateLongitude validates a longitude value specifically
func ValidateLongitude(lng float64) error {
	if lng < -180.0 || lng > 180.0 {
		return &ValidationError{
			Field:   "longitude",
			Value:   fmt.Sprintf("%.6f", lng),
			Message: fmt.Sprintf("longitude %.6f is out of range [-180, 180]", lng),
		}
	}
	return nil
}

// ParseAndValidateCoordinate combines parsing and validation in one step
func ParseAndValidateCoordinate(value string, field string) (float64, error) {
	validator := NewCoordinateValidator()
	coord, err := validator.ParseCoordinate(value)
	if err != nil {
		return 0, err
	}
	
	// Apply field-specific validation
	switch field {
	case "latitude", "lat":
		if err := ValidateLatitude(coord); err != nil {
			return 0, err
		}
	case "longitude", "lng", "lon":
		if err := ValidateLongitude(coord); err != nil {
			return 0, err
		}
	}
	
	return coord, nil
}