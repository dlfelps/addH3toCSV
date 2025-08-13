package validator

import (
	"math"
	"testing"
)

func TestCoordinateValidator_ValidateCoordinates(t *testing.T) {
	validator := NewCoordinateValidator()

	tests := []struct {
		name      string
		lat       float64
		lng       float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid coordinates",
			lat:       40.7128,
			lng:       -74.0060,
			wantError: false,
		},
		{
			name:      "valid boundary coordinates - north pole",
			lat:       90.0,
			lng:       0.0,
			wantError: false,
		},
		{
			name:      "valid boundary coordinates - south pole",
			lat:       -90.0,
			lng:       0.0,
			wantError: false,
		},
		{
			name:      "valid boundary coordinates - antimeridian east",
			lat:       0.0,
			lng:       180.0,
			wantError: false,
		},
		{
			name:      "valid boundary coordinates - antimeridian west",
			lat:       0.0,
			lng:       -180.0,
			wantError: false,
		},
		{
			name:      "invalid latitude - too high",
			lat:       90.1,
			lng:       0.0,
			wantError: true,
			errorMsg:  "latitude 90.100000 is out of range [-90, 90]",
		},
		{
			name:      "invalid latitude - too low",
			lat:       -90.1,
			lng:       0.0,
			wantError: true,
			errorMsg:  "latitude -90.100000 is out of range [-90, 90]",
		},
		{
			name:      "invalid longitude - too high",
			lat:       0.0,
			lng:       180.1,
			wantError: true,
			errorMsg:  "longitude 180.100000 is out of range [-180, 180]",
		},
		{
			name:      "invalid longitude - too low",
			lat:       0.0,
			lng:       -180.1,
			wantError: true,
			errorMsg:  "longitude -180.100000 is out of range [-180, 180]",
		},
		{
			name:      "invalid both coordinates",
			lat:       91.0,
			lng:       181.0,
			wantError: true,
			errorMsg:  "latitude 91.000000 is out of range [-90, 90]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCoordinates(tt.lat, tt.lng)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateCoordinates() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateCoordinates() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCoordinates() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestCoordinateValidator_ParseCoordinate(t *testing.T) {
	validator := NewCoordinateValidator()

	tests := []struct {
		name      string
		value     string
		want      float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid positive decimal",
			value:     "40.7128",
			want:      40.7128,
			wantError: false,
		},
		{
			name:      "valid negative decimal",
			value:     "-74.0060",
			want:      -74.0060,
			wantError: false,
		},
		{
			name:      "valid integer",
			value:     "90",
			want:      90.0,
			wantError: false,
		},
		{
			name:      "valid zero",
			value:     "0",
			want:      0.0,
			wantError: false,
		},
		{
			name:      "valid with leading/trailing spaces",
			value:     "  40.7128  ",
			want:      40.7128,
			wantError: false,
		},
		{
			name:      "valid scientific notation",
			value:     "1.23e2",
			want:      123.0,
			wantError: false,
		},
		{
			name:      "empty string",
			value:     "",
			want:      0,
			wantError: true,
			errorMsg:  "coordinate value is empty or contains only whitespace",
		},
		{
			name:      "whitespace only",
			value:     "   ",
			want:      0,
			wantError: true,
			errorMsg:  "coordinate value is empty or contains only whitespace",
		},
		{
			name:      "invalid format - letters",
			value:     "abc",
			want:      0,
			wantError: true,
			errorMsg:  "invalid coordinate format: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			name:      "invalid format - mixed",
			value:     "40.7abc",
			want:      0,
			wantError: true,
			errorMsg:  "invalid coordinate format: strconv.ParseFloat: parsing \"40.7abc\": invalid syntax",
		},
		{
			name:      "invalid format - multiple dots",
			value:     "40.71.28",
			want:      0,
			wantError: true,
			errorMsg:  "invalid coordinate format: strconv.ParseFloat: parsing \"40.71.28\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.ParseCoordinate(tt.value)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseCoordinate() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ParseCoordinate() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseCoordinate() unexpected error = %v", err)
					return
				}
				if math.Abs(got-tt.want) > 1e-9 {
					t.Errorf("ParseCoordinate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestValidateLatitude(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid latitude - equator",
			lat:       0.0,
			wantError: false,
		},
		{
			name:      "valid latitude - north pole",
			lat:       90.0,
			wantError: false,
		},
		{
			name:      "valid latitude - south pole",
			lat:       -90.0,
			wantError: false,
		},
		{
			name:      "valid latitude - positive",
			lat:       45.5,
			wantError: false,
		},
		{
			name:      "valid latitude - negative",
			lat:       -45.5,
			wantError: false,
		},
		{
			name:      "invalid latitude - too high",
			lat:       90.000001,
			wantError: true,
			errorMsg:  "latitude 90.000001 is out of range [-90, 90]",
		},
		{
			name:      "invalid latitude - too low",
			lat:       -90.000001,
			wantError: true,
			errorMsg:  "latitude -90.000001 is out of range [-90, 90]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLatitude(tt.lat)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateLatitude() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateLatitude() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateLatitude() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateLongitude(t *testing.T) {
	tests := []struct {
		name      string
		lng       float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid longitude - prime meridian",
			lng:       0.0,
			wantError: false,
		},
		{
			name:      "valid longitude - antimeridian east",
			lng:       180.0,
			wantError: false,
		},
		{
			name:      "valid longitude - antimeridian west",
			lng:       -180.0,
			wantError: false,
		},
		{
			name:      "valid longitude - positive",
			lng:       120.5,
			wantError: false,
		},
		{
			name:      "valid longitude - negative",
			lng:       -120.5,
			wantError: false,
		},
		{
			name:      "invalid longitude - too high",
			lng:       180.000001,
			wantError: true,
			errorMsg:  "longitude 180.000001 is out of range [-180, 180]",
		},
		{
			name:      "invalid longitude - too low",
			lng:       -180.000001,
			wantError: true,
			errorMsg:  "longitude -180.000001 is out of range [-180, 180]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLongitude(tt.lng)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateLongitude() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ValidateLongitude() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateLongitude() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestParseAndValidateCoordinate(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		field     string
		want      float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid latitude",
			value:     "45.5",
			field:     "latitude",
			want:      45.5,
			wantError: false,
		},
		{
			name:      "valid longitude",
			value:     "-120.5",
			field:     "longitude",
			want:      -120.5,
			wantError: false,
		},
		{
			name:      "valid latitude - boundary",
			value:     "90",
			field:     "lat",
			want:      90.0,
			wantError: false,
		},
		{
			name:      "valid longitude - boundary",
			value:     "-180",
			field:     "lng",
			want:      -180.0,
			wantError: false,
		},
		{
			name:      "invalid latitude - out of range",
			value:     "91",
			field:     "latitude",
			want:      0,
			wantError: true,
			errorMsg:  "latitude 91.000000 is out of range [-90, 90]",
		},
		{
			name:      "invalid longitude - out of range",
			value:     "181",
			field:     "longitude",
			want:      0,
			wantError: true,
			errorMsg:  "longitude 181.000000 is out of range [-180, 180]",
		},
		{
			name:      "invalid format",
			value:     "abc",
			field:     "latitude",
			want:      0,
			wantError: true,
			errorMsg:  "invalid coordinate format: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			name:      "empty value",
			value:     "",
			field:     "longitude",
			want:      0,
			wantError: true,
			errorMsg:  "coordinate value is empty or contains only whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndValidateCoordinate(tt.value, tt.field)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseAndValidateCoordinate() expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("ParseAndValidateCoordinate() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseAndValidateCoordinate() unexpected error = %v", err)
					return
				}
				if math.Abs(got-tt.want) > 1e-9 {
					t.Errorf("ParseAndValidateCoordinate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "latitude",
		Value:   "91.0",
		Line:    5,
		Message: "latitude 91.0 is out of range [-90, 90]",
	}
	
	expected := "latitude 91.0 is out of range [-90, 90]"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestFileError(t *testing.T) {
	originalErr := &ValidationError{Message: "test error"}
	err := &FileError{
		Path:      "/path/to/file.csv",
		Operation: "read",
		Cause:     originalErr,
	}
	
	expected := "test error"
	if err.Error() != expected {
		t.Errorf("FileError.Error() = %v, want %v", err.Error(), expected)
	}
}