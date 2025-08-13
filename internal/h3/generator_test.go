package h3

import (
	"testing"
	"csv-h3-tool/internal/validator"
)

// MockGenerator implements the Generator interface for testing
type MockGenerator struct {
	generateFunc            func(lat, lng float64, resolution H3Resolution) (string, error)
	validateCoordinatesFunc func(lat, lng float64) error
	validateResolutionFunc  func(resolution H3Resolution) error
}

func (m *MockGenerator) Generate(lat, lng float64, resolution H3Resolution) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(lat, lng, resolution)
	}
	return "mock_h3_index", nil
}

func (m *MockGenerator) ValidateCoordinates(lat, lng float64) error {
	if m.validateCoordinatesFunc != nil {
		return m.validateCoordinatesFunc(lat, lng)
	}
	return nil
}

func (m *MockGenerator) ValidateResolution(resolution H3Resolution) error {
	if m.validateResolutionFunc != nil {
		return m.validateResolutionFunc(resolution)
	}
	return nil
}

// TestGeneratorInterface tests that MockGenerator implements Generator interface
func TestGeneratorInterface(t *testing.T) {
	var _ Generator = &MockGenerator{}
}

// TestH3ResolutionConstants tests that all H3 resolution constants are defined correctly
func TestH3ResolutionConstants(t *testing.T) {
	tests := []struct {
		name       string
		resolution H3Resolution
		expected   int
	}{
		{"ResolutionCountry", ResolutionCountry, 0},
		{"ResolutionState", ResolutionState, 1},
		{"ResolutionMetro", ResolutionMetro, 2},
		{"ResolutionCity", ResolutionCity, 3},
		{"ResolutionDistrict", ResolutionDistrict, 4},
		{"ResolutionNeighbor", ResolutionNeighbor, 5},
		{"ResolutionBlock", ResolutionBlock, 6},
		{"ResolutionBuilding", ResolutionBuilding, 7},
		{"ResolutionStreet", ResolutionStreet, 8},
		{"ResolutionIntersect", ResolutionIntersect, 9},
		{"ResolutionProperty", ResolutionProperty, 10},
		{"ResolutionRoom", ResolutionRoom, 11},
		{"ResolutionDesk", ResolutionDesk, 12},
		{"ResolutionChair", ResolutionChair, 13},
		{"ResolutionBook", ResolutionBook, 14},
		{"ResolutionPage", ResolutionPage, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.resolution) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.resolution), tt.expected)
			}
		})
	}
}

// TestMockGeneratorGenerate tests the mock generator's Generate method
func TestMockGeneratorGenerate(t *testing.T) {
	tests := []struct {
		name       string
		lat        float64
		lng        float64
		resolution H3Resolution
		mockFunc   func(lat, lng float64, resolution H3Resolution) (string, error)
		expected   string
		expectErr  bool
	}{
		{
			name:       "default mock behavior",
			lat:        37.7749,
			lng:        -122.4194,
			resolution: ResolutionStreet,
			mockFunc:   nil,
			expected:   "mock_h3_index",
			expectErr:  false,
		},
		{
			name:       "custom mock behavior",
			lat:        40.7128,
			lng:        -74.0060,
			resolution: ResolutionCity,
			mockFunc: func(lat, lng float64, resolution H3Resolution) (string, error) {
				return "custom_h3_index", nil
			},
			expected:  "custom_h3_index",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockGenerator{
				generateFunc: tt.mockFunc,
			}

			result, err := mock.Generate(tt.lat, tt.lng, tt.resolution)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Generate() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestMockGeneratorValidateCoordinates tests the mock generator's ValidateCoordinates method
func TestMockGeneratorValidateCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		lat       float64
		lng       float64
		mockFunc  func(lat, lng float64) error
		expectErr bool
	}{
		{
			name:      "default mock behavior - valid coordinates",
			lat:       37.7749,
			lng:       -122.4194,
			mockFunc:  nil,
			expectErr: false,
		},
		{
			name:     "custom mock behavior - invalid coordinates",
			lat:      91.0,
			lng:      -122.4194,
			mockFunc: func(lat, lng float64) error { return &ValidationError{Field: "latitude", Message: "invalid latitude"} },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockGenerator{
				validateCoordinatesFunc: tt.mockFunc,
			}

			err := mock.ValidateCoordinates(tt.lat, tt.lng)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestMockGeneratorValidateResolution tests the mock generator's ValidateResolution method
func TestMockGeneratorValidateResolution(t *testing.T) {
	tests := []struct {
		name       string
		resolution H3Resolution
		mockFunc   func(resolution H3Resolution) error
		expectErr  bool
	}{
		{
			name:       "default mock behavior - valid resolution",
			resolution: ResolutionStreet,
			mockFunc:   nil,
			expectErr:  false,
		},
		{
			name:       "custom mock behavior - invalid resolution",
			resolution: H3Resolution(16),
			mockFunc:   func(resolution H3Resolution) error { return &ValidationError{Field: "resolution", Message: "invalid resolution"} },
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockGenerator{
				validateResolutionFunc: tt.mockFunc,
			}

			err := mock.ValidateResolution(tt.resolution)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ValidationError for testing purposes
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// TestBaseGeneratorValidateCoordinates tests the BaseGenerator's coordinate validation
func TestBaseGeneratorValidateCoordinates(t *testing.T) {
	generator := &BaseGenerator{
		validator: validator.NewCoordinateValidator(),
	}

	tests := []struct {
		name      string
		lat       float64
		lng       float64
		expectErr bool
	}{
		{
			name:      "valid coordinates - San Francisco",
			lat:       37.7749,
			lng:       -122.4194,
			expectErr: false,
		},
		{
			name:      "valid coordinates - New York",
			lat:       40.7128,
			lng:       -74.0060,
			expectErr: false,
		},
		{
			name:      "valid coordinates - North Pole",
			lat:       90.0,
			lng:       0.0,
			expectErr: false,
		},
		{
			name:      "valid coordinates - South Pole",
			lat:       -90.0,
			lng:       0.0,
			expectErr: false,
		},
		{
			name:      "valid coordinates - International Date Line",
			lat:       0.0,
			lng:       180.0,
			expectErr: false,
		},
		{
			name:      "valid coordinates - Prime Meridian",
			lat:       0.0,
			lng:       -180.0,
			expectErr: false,
		},
		{
			name:      "invalid latitude - too high",
			lat:       91.0,
			lng:       0.0,
			expectErr: true,
		},
		{
			name:      "invalid latitude - too low",
			lat:       -91.0,
			lng:       0.0,
			expectErr: true,
		},
		{
			name:      "invalid longitude - too high",
			lat:       0.0,
			lng:       181.0,
			expectErr: true,
		},
		{
			name:      "invalid longitude - too low",
			lat:       0.0,
			lng:       -181.0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.ValidateCoordinates(tt.lat, tt.lng)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestBaseGeneratorValidateResolution tests the BaseGenerator's resolution validation
func TestBaseGeneratorValidateResolution(t *testing.T) {
	generator := &BaseGenerator{
		validator: validator.NewCoordinateValidator(),
	}

	tests := []struct {
		name       string
		resolution H3Resolution
		expectErr  bool
	}{
		{
			name:       "valid resolution - 0",
			resolution: H3Resolution(0),
			expectErr:  false,
		},
		{
			name:       "valid resolution - 8 (default)",
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "valid resolution - 15 (maximum)",
			resolution: H3Resolution(15),
			expectErr:  false,
		},
		{
			name:       "invalid resolution - negative",
			resolution: H3Resolution(-1),
			expectErr:  true,
		},
		{
			name:       "invalid resolution - too high",
			resolution: H3Resolution(16),
			expectErr:  true,
		},
		{
			name:       "invalid resolution - way too high",
			resolution: H3Resolution(100),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.ValidateResolution(tt.resolution)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestBaseGeneratorConstruction tests the BaseGenerator construction
func TestBaseGeneratorConstruction(t *testing.T) {
	generator := &BaseGenerator{
		validator: validator.NewCoordinateValidator(),
	}
	
	if generator == nil {
		t.Error("BaseGenerator construction failed")
	}
	
	if generator.validator == nil {
		t.Error("BaseGenerator validator is nil")
	}
}

// TestNewH3Generator tests the H3Generator constructor
func TestNewH3Generator(t *testing.T) {
	generator := NewH3Generator()
	
	if generator == nil {
		t.Error("NewH3Generator() returned nil")
	}
	
	// With value embedding, we can test that the validator is properly initialized
	if generator.validator == nil {
		t.Error("H3Generator validator is nil")
	}
}

// TestH3GeneratorInterface tests that H3Generator implements Generator interface
func TestH3GeneratorInterface(t *testing.T) {
	var _ Generator = &H3Generator{}
}

// TestH3GeneratorGenerate tests the H3 index generation functionality
func TestH3GeneratorGenerate(t *testing.T) {
	generator := NewH3Generator()

	tests := []struct {
		name       string
		lat        float64
		lng        float64
		resolution H3Resolution
		expectErr  bool
		checkIndex bool
	}{
		{
			name:       "San Francisco - Resolution 8",
			lat:        37.7749,
			lng:        -122.4194,
			resolution: ResolutionStreet,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "New York - Resolution 8",
			lat:        40.7128,
			lng:        -74.0060,
			resolution: ResolutionStreet,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "London - Resolution 10",
			lat:        51.5074,
			lng:        -0.1278,
			resolution: ResolutionProperty,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "Tokyo - Resolution 5",
			lat:        35.6762,
			lng:        139.6503,
			resolution: ResolutionNeighbor,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "North Pole - Resolution 0",
			lat:        90.0,
			lng:        0.0,
			resolution: ResolutionCountry,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "South Pole - Resolution 15",
			lat:        -90.0,
			lng:        0.0,
			resolution: ResolutionPage,
			expectErr:  false,
			checkIndex: true,
		},
		{
			name:       "Invalid latitude - too high",
			lat:        91.0,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  true,
			checkIndex: false,
		},
		{
			name:       "Invalid longitude - too low",
			lat:        0.0,
			lng:        -181.0,
			resolution: ResolutionStreet,
			expectErr:  true,
			checkIndex: false,
		},
		{
			name:       "Invalid resolution - negative",
			lat:        37.7749,
			lng:        -122.4194,
			resolution: H3Resolution(-1),
			expectErr:  true,
			checkIndex: false,
		},
		{
			name:       "Invalid resolution - too high",
			lat:        37.7749,
			lng:        -122.4194,
			resolution: H3Resolution(16),
			expectErr:  true,
			checkIndex: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, err := generator.Generate(tt.lat, tt.lng, tt.resolution)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.checkIndex && !tt.expectErr {
				if index == "" {
					t.Error("expected non-empty H3 index")
				}
				// H3 indexes should be hexadecimal strings
				if len(index) == 0 {
					t.Error("H3 index should not be empty")
				}
				t.Logf("Generated H3 index: %s", index)
			}
		})
	}
}

// TestH3GeneratorConsistency tests that the same coordinates always produce the same H3 index
func TestH3GeneratorConsistency(t *testing.T) {
	generator := NewH3Generator()

	lat := 37.7749
	lng := -122.4194
	resolution := ResolutionStreet

	// Generate the same index multiple times
	index1, err1 := generator.Generate(lat, lng, resolution)
	index2, err2 := generator.Generate(lat, lng, resolution)
	index3, err3 := generator.Generate(lat, lng, resolution)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("unexpected errors: %v, %v, %v", err1, err2, err3)
	}

	if index1 != index2 || index2 != index3 {
		t.Errorf("H3 indexes should be consistent: %s, %s, %s", index1, index2, index3)
	}
}

// TestH3GeneratorResolutionLevels tests different resolution levels produce different indexes
func TestH3GeneratorResolutionLevels(t *testing.T) {
	generator := NewH3Generator()

	lat := 37.7749
	lng := -122.4194

	resolutions := []H3Resolution{
		ResolutionCountry,
		ResolutionState,
		ResolutionCity,
		ResolutionStreet,
		ResolutionProperty,
		ResolutionPage,
	}

	indexes := make(map[H3Resolution]string)

	// Generate indexes for different resolutions
	for _, resolution := range resolutions {
		index, err := generator.Generate(lat, lng, resolution)
		if err != nil {
			t.Fatalf("unexpected error for resolution %d: %v", resolution, err)
		}
		indexes[resolution] = index
		t.Logf("Resolution %d: %s", resolution, index)
	}

	// Verify that different resolutions produce different indexes
	for i, res1 := range resolutions {
		for j, res2 := range resolutions {
			if i != j && indexes[res1] == indexes[res2] {
				t.Errorf("Different resolutions %d and %d produced same index: %s", res1, res2, indexes[res1])
			}
		}
	}
}

// TestH3GeneratorEdgeCases tests edge cases for coordinate boundaries
func TestH3GeneratorEdgeCases(t *testing.T) {
	generator := NewH3Generator()

	tests := []struct {
		name       string
		lat        float64
		lng        float64
		resolution H3Resolution
		expectErr  bool
	}{
		{
			name:       "Equator and Prime Meridian",
			lat:        0.0,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "International Date Line East",
			lat:        0.0,
			lng:        180.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "International Date Line West",
			lat:        0.0,
			lng:        -180.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "Arctic Circle",
			lat:        66.5622,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "Antarctic Circle",
			lat:        -66.5622,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "Tropic of Cancer",
			lat:        23.4372,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
		{
			name:       "Tropic of Capricorn",
			lat:        -23.4372,
			lng:        0.0,
			resolution: ResolutionStreet,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, err := generator.Generate(tt.lat, tt.lng, tt.resolution)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && index == "" {
				t.Error("expected non-empty H3 index")
			}
		})
	}
}