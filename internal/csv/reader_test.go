package csv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewReader(t *testing.T) {
	// Create a temporary CSV file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York\n34.0522,-118.2437,Los Angeles"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer reader.Close()

	if reader.latIndex != 0 {
		t.Errorf("Expected latitude index 0, got %d", reader.latIndex)
	}
	if reader.lngIndex != 1 {
		t.Errorf("Expected longitude index 1, got %d", reader.lngIndex)
	}
}

func TestNewReaderWithoutHeaders(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "40.7128,-74.0060,New York\n34.0522,-118.2437,Los Angeles"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "0",
		LngColumn:  "1",
		HasHeaders: false,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer reader.Close()

	if reader.latIndex != 0 {
		t.Errorf("Expected latitude index 0, got %d", reader.latIndex)
	}
	if reader.lngIndex != 1 {
		t.Errorf("Expected longitude index 1, got %d", reader.lngIndex)
	}
}

func TestDetectColumnsByName(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		latColumn   string
		lngColumn   string
		expectedLat int
		expectedLng int
		shouldError bool
	}{
		{
			name:        "exact match",
			headers:     []string{"latitude", "longitude", "name"},
			latColumn:   "latitude",
			lngColumn:   "longitude",
			expectedLat: 0,
			expectedLng: 1,
			shouldError: false,
		},
		{
			name:        "case insensitive",
			headers:     []string{"LATITUDE", "LONGITUDE", "name"},
			latColumn:   "latitude",
			lngColumn:   "longitude",
			expectedLat: 0,
			expectedLng: 1,
			shouldError: false,
		},
		{
			name:        "fallback names",
			headers:     []string{"lat", "lng", "name"},
			latColumn:   "",
			lngColumn:   "",
			expectedLat: 0,
			expectedLng: 1,
			shouldError: false,
		},
		{
			name:        "alternative fallbacks",
			headers:     []string{"y", "x", "name"},
			latColumn:   "",
			lngColumn:   "",
			expectedLat: 0,
			expectedLng: 1,
			shouldError: false,
		},
		{
			name:        "missing latitude",
			headers:     []string{"longitude", "name"},
			latColumn:   "latitude",
			lngColumn:   "longitude",
			expectedLat: -1,
			expectedLng: 0,
			shouldError: true,
		},
		{
			name:        "missing longitude",
			headers:     []string{"latitude", "name"},
			latColumn:   "latitude",
			lngColumn:   "longitude",
			expectedLat: 0,
			expectedLng: -1,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &Reader{
				headers:    tt.headers,
				hasHeaders: true,
			}

			config := Config{
				LatColumn:  tt.latColumn,
				LngColumn:  tt.lngColumn,
				HasHeaders: true,
			}

			err := reader.detectColumns(config)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldError {
				if reader.latIndex != tt.expectedLat {
					t.Errorf("Expected latitude index %d, got %d", tt.expectedLat, reader.latIndex)
				}
				if reader.lngIndex != tt.expectedLng {
					t.Errorf("Expected longitude index %d, got %d", tt.expectedLng, reader.lngIndex)
				}
			}
		})
	}
}

func TestReadRecord(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")
	
	csvContent := "latitude,longitude,name\n40.7128,-74.0060,New York\n34.0522,-118.2437,Los Angeles\n,,-Empty\ninvalid,invalid,Invalid"
	if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	config := Config{
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		HasHeaders: true,
	}

	reader, err := NewReader(testFile, config)
	if err != nil {
		t.Fatalf("NewReader failed: %v", err)
	}
	defer reader.Close()

	// Read first valid record
	record, err := reader.ReadRecord()
	if err != nil {
		t.Fatalf("Failed to read first record: %v", err)
	}
	if !record.IsValid {
		t.Error("Expected first record to be valid")
	}
	if record.Latitude != 40.7128 {
		t.Errorf("Expected latitude 40.7128, got %f", record.Latitude)
	}
	if record.Longitude != -74.0060 {
		t.Errorf("Expected longitude -74.0060, got %f", record.Longitude)
	}
	if len(record.OriginalData) != 3 {
		t.Errorf("Expected 3 original data fields, got %d", len(record.OriginalData))
	}

	// Read second valid record
	record, err = reader.ReadRecord()
	if err != nil {
		t.Fatalf("Failed to read second record: %v", err)
	}
	if !record.IsValid {
		t.Error("Expected second record to be valid")
	}

	// Read empty record
	record, err = reader.ReadRecord()
	if err != nil {
		t.Fatalf("Failed to read empty record: %v", err)
	}
	if record.IsValid {
		t.Error("Expected empty record to be invalid")
	}

	// Read invalid record
	record, err = reader.ReadRecord()
	if err != nil {
		t.Fatalf("Failed to read invalid record: %v", err)
	}
	if record.IsValid {
		t.Error("Expected invalid record to be invalid")
	}
}

func TestValidateColumns(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		config      Config
		shouldError bool
	}{
		{
			name:    "valid headers with column names",
			headers: []string{"latitude", "longitude", "name"},
			config: Config{
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				HasHeaders: true,
			},
			shouldError: false,
		},
		{
			name:    "valid headers with fallback names",
			headers: []string{"lat", "lng", "name"},
			config: Config{
				LatColumn:  "",
				LngColumn:  "",
				HasHeaders: true,
			},
			shouldError: false,
		},
		{
			name:    "missing latitude column",
			headers: []string{"longitude", "name"},
			config: Config{
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				HasHeaders: true,
			},
			shouldError: true,
		},
		{
			name:    "no headers with indices",
			headers: nil,
			config: Config{
				LatColumn:  "0",
				LngColumn:  "1",
				HasHeaders: false,
			},
			shouldError: false,
		},
		{
			name:    "no headers without column specs",
			headers: nil,
			config: Config{
				LatColumn:  "",
				LngColumn:  "",
				HasHeaders: false,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.headers, tt.config)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestReaderEdgeCases(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		config := Config{
			LatColumn:  "latitude",
			LngColumn:  "longitude",
			HasHeaders: true,
		}

		_, err := NewReader("nonexistent.csv", config)
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.csv")
		
		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config := Config{
			LatColumn:  "latitude",
			LngColumn:  "longitude",
			HasHeaders: true,
		}

		_, err := NewReader(testFile, config)
		if err == nil {
			t.Error("Expected error for empty file with headers")
		}
	})

	t.Run("insufficient columns", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "insufficient.csv")
		
		csvContent := "latitude,longitude\n40.7128"
		if err := os.WriteFile(testFile, []byte(csvContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config := Config{
			LatColumn:  "latitude",
			LngColumn:  "longitude",
			HasHeaders: true,
		}

		reader, err := NewReader(testFile, config)
		if err != nil {
			t.Fatalf("NewReader failed: %v", err)
		}
		defer reader.Close()

		_, err = reader.ReadRecord()
		if err == nil {
			t.Error("Expected error for insufficient columns")
		}
	})
}