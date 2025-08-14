package integration

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// BenchmarkSuite contains benchmark tests for performance evaluation
type BenchmarkSuite struct {
	tempDir string
}

// setupBenchmarkSuite creates temporary directory for benchmark tests
func setupBenchmarkSuite(b *testing.B) *BenchmarkSuite {
	tempDir, err := os.MkdirTemp("", "csv-h3-benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}

	return &BenchmarkSuite{
		tempDir: tempDir,
	}
}

// cleanup removes temporary benchmark files
func (suite *BenchmarkSuite) cleanup() {
	os.RemoveAll(suite.tempDir)
}

// createBenchmarkFile creates a CSV file with specified number of records
func (suite *BenchmarkSuite) createBenchmarkFile(b *testing.B, filename string, numRecords int) string {
	filePath := filepath.Join(suite.tempDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "latitude", "longitude", "name", "category"}); err != nil {
		b.Fatalf("Failed to write header: %v", err)
	}

	// Generate realistic coordinate data
	for i := 0; i < numRecords; i++ {
		// Generate coordinates that cover the globe
		lat := fmt.Sprintf("%.6f", float64((i%180)-90))    // -90 to 89
		lng := fmt.Sprintf("%.6f", float64((i%360)-180))   // -180 to 179
		id := fmt.Sprintf("LOC_%06d", i)
		name := fmt.Sprintf("Location_%d", i)
		category := fmt.Sprintf("Category_%d", i%10)

		record := []string{id, lat, lng, name, category}
		if err := writer.Write(record); err != nil {
			b.Fatalf("Failed to write record %d: %v", i, err)
		}
	}

	return filePath
}

// BenchmarkH3Generation benchmarks H3 index generation performance
func BenchmarkH3Generation(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	benchmarks := []struct {
		name       string
		numRecords int
		resolution int
	}{
		{"Small_100_Res8", 100, 8},
		{"Medium_1000_Res8", 1000, 8},
		{"Large_10000_Res8", 10000, 8},
		{"Small_100_Res10", 100, 10},
		{"Medium_1000_Res10", 1000, 10},
		{"Small_100_Res12", 100, 12},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			inputFile := suite.createBenchmarkFile(b, fmt.Sprintf("bench_%s.csv", bm.name), bm.numRecords)
			
			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(suite.tempDir, fmt.Sprintf("output_%s.csv", bm.name)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: bm.resolution,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Create fresh output file for each iteration
				cfg.OutputFile = filepath.Join(suite.tempDir, fmt.Sprintf("output_%s_%d.csv", bm.name, i))
				
				orchestrator := service.NewOrchestrator(cfg)
				result, err := orchestrator.ProcessFile()
				if err != nil {
					b.Fatalf("Benchmark processing failed: %v", err)
				}

				if result.ValidRecords != bm.numRecords {
					b.Errorf("Expected %d valid records, got %d", bm.numRecords, result.ValidRecords)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage with large files
func BenchmarkMemoryUsage(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	// Test with progressively larger files to check memory scaling
	sizes := []int{1000, 5000, 10000, 25000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Records_%d", size), func(b *testing.B) {
			inputFile := suite.createBenchmarkFile(b, fmt.Sprintf("memory_test_%d.csv", size), size)
			
			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(suite.tempDir, fmt.Sprintf("memory_output_%d.csv", size)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			b.ResetTimer()
			b.ReportAllocs()

			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			if err != nil {
				b.Fatalf("Memory benchmark failed: %v", err)
			}

			if result.ValidRecords != size {
				b.Errorf("Expected %d valid records, got %d", size, result.ValidRecords)
			}
		})
	}
}

// BenchmarkStreamingEfficiency tests streaming vs batch processing efficiency
func BenchmarkStreamingEfficiency(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	// Create a large file to test streaming efficiency
	numRecords := 50000
	inputFile := suite.createBenchmarkFile(b, "streaming_test.csv", numRecords)

	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(suite.tempDir, "streaming_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cfg.OutputFile = filepath.Join(suite.tempDir, fmt.Sprintf("streaming_output_%d.csv", i))
		
		orchestrator := service.NewOrchestrator(cfg)
		result, err := orchestrator.ProcessFile()
		if err != nil {
			b.Fatalf("Streaming benchmark failed: %v", err)
		}

		if result.ValidRecords != numRecords {
			b.Errorf("Expected %d valid records, got %d", numRecords, result.ValidRecords)
		}
	}
}

// BenchmarkResolutionLevels benchmarks different H3 resolution levels
func BenchmarkResolutionLevels(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	numRecords := 1000
	inputFile := suite.createBenchmarkFile(b, "resolution_test.csv", numRecords)

	resolutions := []int{0, 3, 6, 8, 10, 12, 15}

	for _, resolution := range resolutions {
		b.Run(fmt.Sprintf("Resolution_%d", resolution), func(b *testing.B) {
			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(suite.tempDir, fmt.Sprintf("res_%d_output.csv", resolution)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: resolution,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				cfg.OutputFile = filepath.Join(suite.tempDir, fmt.Sprintf("res_%d_output_%d.csv", resolution, i))
				
				orchestrator := service.NewOrchestrator(cfg)
				result, err := orchestrator.ProcessFile()
				if err != nil {
					b.Fatalf("Resolution benchmark failed: %v", err)
				}

				if result.ValidRecords != numRecords {
					b.Errorf("Expected %d valid records, got %d", numRecords, result.ValidRecords)
				}
			}
		})
	}
}

// BenchmarkErrorHandling benchmarks performance with mixed valid/invalid data
func BenchmarkErrorHandling(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	// Create file with mixed valid/invalid data
	filePath := filepath.Join(suite.tempDir, "mixed_data.csv")
	file, err := os.Create(filePath)
	if err != nil {
		b.Fatalf("Failed to create mixed data file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "latitude", "longitude", "name"}); err != nil {
		b.Fatalf("Failed to write header: %v", err)
	}

	numRecords := 5000
	for i := 0; i < numRecords; i++ {
		var record []string
		
		if i%4 == 0 {
			// Invalid latitude (25% of records)
			record = []string{fmt.Sprintf("ID_%d", i), "91.0", "0.0", fmt.Sprintf("Invalid_Lat_%d", i)}
		} else if i%4 == 1 {
			// Invalid longitude (25% of records)
			record = []string{fmt.Sprintf("ID_%d", i), "0.0", "181.0", fmt.Sprintf("Invalid_Lng_%d", i)}
		} else if i%4 == 2 {
			// Empty coordinates (25% of records)
			record = []string{fmt.Sprintf("ID_%d", i), "", "", fmt.Sprintf("Empty_%d", i)}
		} else {
			// Valid coordinates (25% of records)
			lat := fmt.Sprintf("%.6f", float64((i%180)-90))
			lng := fmt.Sprintf("%.6f", float64((i%360)-180))
			record = []string{fmt.Sprintf("ID_%d", i), lat, lng, fmt.Sprintf("Valid_%d", i)}
		}

		if err := writer.Write(record); err != nil {
			b.Fatalf("Failed to write record %d: %v", i, err)
		}
	}

	cfg := &config.Config{
		InputFile:  filePath,
		OutputFile: filepath.Join(suite.tempDir, "mixed_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cfg.OutputFile = filepath.Join(suite.tempDir, fmt.Sprintf("mixed_output_%d.csv", i))
		
		orchestrator := service.NewOrchestrator(cfg)
		result, err := orchestrator.ProcessFile()
		if err != nil {
			b.Fatalf("Mixed data benchmark failed: %v", err)
		}

		expectedValid := numRecords / 4 // Only 25% should be valid
		if result.ValidRecords != expectedValid {
			b.Errorf("Expected %d valid records, got %d", expectedValid, result.ValidRecords)
		}
	}
}

// TestBenchmarkValidation validates that benchmark results are reasonable
func TestBenchmarkValidation(t *testing.T) {
	suite := setupBenchmarkSuite(&testing.B{})
	defer suite.cleanup()

	// Create a test file
	numRecords := 1000
	inputFile := suite.createBenchmarkFile(&testing.B{}, "validation_test.csv", numRecords)

	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(suite.tempDir, "validation_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    false,
	}

	start := time.Now()
	orchestrator := service.NewOrchestrator(cfg)
	result, err := orchestrator.ProcessFile()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Validation test failed: %v", err)
	}

	// Validate results
	if result.TotalRecords != numRecords {
		t.Errorf("Expected %d total records, got %d", numRecords, result.TotalRecords)
	}

	if result.ValidRecords != numRecords {
		t.Errorf("Expected %d valid records, got %d", numRecords, result.ValidRecords)
	}

	// Performance validation - should process 1000 records quickly
	if duration > 10*time.Second {
		t.Errorf("Processing took too long: %v (expected < 10s for 1000 records)", duration)
	}

	t.Logf("Processed %d records in %v (%.2f records/sec)", 
		numRecords, duration, float64(numRecords)/duration.Seconds())
}

// BenchmarkConcurrentProcessing tests if the tool can handle concurrent usage
func BenchmarkConcurrentProcessing(b *testing.B) {
	suite := setupBenchmarkSuite(b)
	defer suite.cleanup()

	numRecords := 1000
	inputFile := suite.createBenchmarkFile(b, "concurrent_test.csv", numRecords)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(suite.tempDir, fmt.Sprintf("concurrent_output_%d.csv", i)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			if err != nil {
				b.Fatalf("Concurrent processing failed: %v", err)
			}

			if result.ValidRecords != numRecords {
				b.Errorf("Expected %d valid records, got %d", numRecords, result.ValidRecords)
			}
			i++
		}
	})
}