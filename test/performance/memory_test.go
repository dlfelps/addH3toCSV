package performance

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/service"
)

// MemoryStats captures memory usage statistics
type MemoryStats struct {
	AllocBytes      uint64
	TotalAllocBytes uint64
	SysBytes        uint64
	NumGC           uint32
	HeapObjects     uint64
}

// getMemoryStats returns current memory statistics
func getMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return MemoryStats{
		AllocBytes:      m.Alloc,
		TotalAllocBytes: m.TotalAlloc,
		SysBytes:        m.Sys,
		NumGC:           m.NumGC,
		HeapObjects:     m.HeapObjects,
	}
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// createLargeTestFile creates a CSV file with specified number of records
func createLargeTestFile(t *testing.T, filePath string, numRecords int) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "latitude", "longitude", "name", "category", "timestamp", "value"}); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	// Generate records with realistic data
	for i := 0; i < numRecords; i++ {
		lat := fmt.Sprintf("%.6f", float64((i%180)-90))    // -90 to 89
		lng := fmt.Sprintf("%.6f", float64((i%360)-180))   // -180 to 179
		id := fmt.Sprintf("ID_%08d", i)
		name := fmt.Sprintf("Location_%d_%s", i, generateRandomString(10))
		category := fmt.Sprintf("Category_%d", i%50)
		timestamp := fmt.Sprintf("2024-01-%02d %02d:%02d:%02d", 
			(i%28)+1, (i%24), (i%60), (i%60))
		value := fmt.Sprintf("%.2f", float64(i)*1.23)

		record := []string{id, lat, lng, name, category, timestamp, value}
		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}
}

// generateRandomString generates a pseudo-random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// TestMemoryUsageScaling tests memory usage with increasing file sizes
func TestMemoryUsageScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-memory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with different file sizes
	testSizes := []struct {
		name       string
		numRecords int
		maxMemoryMB float64 // Maximum expected memory usage in MB
	}{
		{"Small_1K", 1000, 50},
		{"Medium_5K", 5000, 100},
		{"Large_10K", 10000, 150},
		{"XLarge_25K", 25000, 250},
		{"XXLarge_50K", 50000, 400},
	}

	for _, test := range testSizes {
		t.Run(test.name, func(t *testing.T) {
			// Force garbage collection before test
			runtime.GC()
			runtime.GC() // Call twice to ensure cleanup
			
			// Get baseline memory stats
			baselineStats := getMemoryStats()
			
			// Create test file
			inputFile := filepath.Join(tempDir, fmt.Sprintf("memory_test_%s.csv", test.name))
			createLargeTestFile(t, inputFile, test.numRecords)
			
			// Configure processing
			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, fmt.Sprintf("memory_output_%s.csv", test.name)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			// Get memory stats before processing
			_ = getMemoryStats()
			
			// Process the file
			start := time.Now()
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			duration := time.Since(start)
			
			if err != nil {
				t.Fatalf("Memory test failed: %v", err)
			}

			// Get memory stats after processing
			postProcessStats := getMemoryStats()
			
			// Force garbage collection to see actual retained memory
			runtime.GC()
			runtime.GC()
			finalStats := getMemoryStats()

			// Calculate memory usage
			peakMemoryBytes := postProcessStats.AllocBytes - baselineStats.AllocBytes
			retainedMemoryBytes := finalStats.AllocBytes - baselineStats.AllocBytes
			peakMemoryMB := float64(peakMemoryBytes) / (1024 * 1024)
			
			// Validate memory usage
			if peakMemoryMB > test.maxMemoryMB {
				t.Errorf("Memory usage too high: %.2f MB (max: %.2f MB)", 
					peakMemoryMB, test.maxMemoryMB)
			}

			// Validate processing results
			if result.TotalRecords != test.numRecords {
				t.Errorf("Expected %d records, got %d", test.numRecords, result.TotalRecords)
			}

			// Calculate performance metrics
			recordsPerSecond := float64(test.numRecords) / duration.Seconds()
			bytesPerRecord := float64(peakMemoryBytes) / float64(test.numRecords)

			t.Logf("Memory Test %s Results:", test.name)
			t.Logf("  Records: %d", test.numRecords)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Records/sec: %.2f", recordsPerSecond)
			t.Logf("  Peak Memory: %s (%.2f MB)", formatBytes(peakMemoryBytes), peakMemoryMB)
			t.Logf("  Retained Memory: %s", formatBytes(retainedMemoryBytes))
			t.Logf("  Memory/Record: %.2f bytes", bytesPerRecord)
			t.Logf("  GC Runs: %d", postProcessStats.NumGC-baselineStats.NumGC)
			t.Logf("  Heap Objects: %d", postProcessStats.HeapObjects)
		})
	}
}

// TestStreamingMemoryEfficiency tests that streaming processing keeps memory usage constant
func TestStreamingMemoryEfficiency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming memory tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-streaming-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a large file to test streaming
	numRecords := 100000 // 100K records
	inputFile := filepath.Join(tempDir, "streaming_test.csv")
	createLargeTestFile(t, inputFile, numRecords)

	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(tempDir, "streaming_output.csv"),
		LatColumn:  "latitude",
		LngColumn:  "longitude",
		Resolution: 8,
		HasHeaders: true,
		Overwrite:  true,
		Verbose:    false,
	}

	// Monitor memory usage during processing
	memorySnapshots := make([]MemoryStats, 0)
	
	// Force GC before starting
	runtime.GC()
	runtime.GC()
	baselineStats := getMemoryStats()
	memorySnapshots = append(memorySnapshots, baselineStats)

	// Process the file
	start := time.Now()
	orchestrator := service.NewOrchestrator(cfg)
	result, err := orchestrator.ProcessFile()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Streaming test failed: %v", err)
	}

	// Get final memory stats
	finalStats := getMemoryStats()
	memorySnapshots = append(memorySnapshots, finalStats)

	// Analyze memory usage
	peakMemoryBytes := finalStats.AllocBytes - baselineStats.AllocBytes
	peakMemoryMB := float64(peakMemoryBytes) / (1024 * 1024)

	// For streaming processing, memory usage should be relatively constant
	// and not scale linearly with file size
	maxExpectedMemoryMB := 200.0 // Should not exceed 200MB for any size file
	
	if peakMemoryMB > maxExpectedMemoryMB {
		t.Errorf("Streaming memory usage too high: %.2f MB (max expected: %.2f MB)", 
			peakMemoryMB, maxExpectedMemoryMB)
	}

	// Validate results
	if result.TotalRecords != numRecords {
		t.Errorf("Expected %d records, got %d", numRecords, result.TotalRecords)
	}

	recordsPerSecond := float64(numRecords) / duration.Seconds()
	
	t.Logf("Streaming Memory Test Results:")
	t.Logf("  Records: %d", numRecords)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Records/sec: %.2f", recordsPerSecond)
	t.Logf("  Peak Memory: %s (%.2f MB)", formatBytes(peakMemoryBytes), peakMemoryMB)
	t.Logf("  Memory/Record: %.2f bytes", float64(peakMemoryBytes)/float64(numRecords))
	t.Logf("  GC Runs: %d", finalStats.NumGC-baselineStats.NumGC)
}

// TestMemoryLeaks tests for memory leaks during repeated processing
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-leak-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	numRecords := 5000
	inputFile := filepath.Join(tempDir, "leak_test.csv")
	createLargeTestFile(t, inputFile, numRecords)

	// Force GC and get baseline
	runtime.GC()
	runtime.GC()
	baselineStats := getMemoryStats()

	// Run multiple iterations to detect memory leaks
	iterations := 10
	memoryGrowth := make([]uint64, iterations)

	for i := 0; i < iterations; i++ {
		cfg := &config.Config{
			InputFile:  inputFile,
			OutputFile: filepath.Join(tempDir, fmt.Sprintf("leak_output_%d.csv", i)),
			LatColumn:  "latitude",
			LngColumn:  "longitude",
			Resolution: 8,
			HasHeaders: true,
			Overwrite:  true,
			Verbose:    false,
		}

		// Process file
		orchestrator := service.NewOrchestrator(cfg)
		result, err := orchestrator.ProcessFile()
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", i, err)
		}

		if result.TotalRecords != numRecords {
			t.Errorf("Iteration %d: expected %d records, got %d", i, numRecords, result.TotalRecords)
		}

		// Force GC and measure memory
		runtime.GC()
		runtime.GC()
		currentStats := getMemoryStats()
		memoryGrowth[i] = currentStats.AllocBytes - baselineStats.AllocBytes

		t.Logf("Iteration %d: Memory usage: %s", i+1, formatBytes(memoryGrowth[i]))
	}

	// Analyze memory growth
	firstIterationMemory := memoryGrowth[0]
	lastIterationMemory := memoryGrowth[iterations-1]
	
	// Memory should not grow significantly between iterations
	// Allow for some variance but flag if memory doubles
	maxAllowedGrowth := firstIterationMemory * 2
	
	if lastIterationMemory > maxAllowedGrowth {
		t.Errorf("Potential memory leak detected: memory grew from %s to %s over %d iterations",
			formatBytes(firstIterationMemory), formatBytes(lastIterationMemory), iterations)
	}

	// Calculate average memory usage (excluding first iteration which may be higher due to initialization)
	if iterations > 1 {
		var totalMemory uint64
		for i := 1; i < iterations; i++ {
			totalMemory += memoryGrowth[i]
		}
		avgMemory := totalMemory / uint64(iterations-1)
		
		t.Logf("Memory Leak Test Results:")
		t.Logf("  Iterations: %d", iterations)
		t.Logf("  Records per iteration: %d", numRecords)
		t.Logf("  First iteration memory: %s", formatBytes(firstIterationMemory))
		t.Logf("  Last iteration memory: %s", formatBytes(lastIterationMemory))
		t.Logf("  Average memory (excluding first): %s", formatBytes(avgMemory))
		t.Logf("  Memory growth: %s", formatBytes(lastIterationMemory-firstIterationMemory))
	}
}

// TestGarbageCollectionBehavior tests GC behavior under different loads
func TestGarbageCollectionBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GC behavior tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-gc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name       string
		numRecords int
		maxGCRuns  uint32
	}{
		{"Small_1K", 1000, 5},
		{"Medium_10K", 10000, 15},
		{"Large_50K", 50000, 50},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// Create test file
			inputFile := filepath.Join(tempDir, fmt.Sprintf("gc_test_%s.csv", test.name))
			createLargeTestFile(t, inputFile, test.numRecords)

			// Force GC and get baseline
			runtime.GC()
			runtime.GC()
			baselineStats := getMemoryStats()

			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, fmt.Sprintf("gc_output_%s.csv", test.name)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			// Process file
			start := time.Now()
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("GC test failed: %v", err)
			}

			// Get final stats
			finalStats := getMemoryStats()
			gcRuns := finalStats.NumGC - baselineStats.NumGC

			// Validate GC behavior
			if gcRuns > test.maxGCRuns {
				t.Errorf("Too many GC runs: %d (max expected: %d)", gcRuns, test.maxGCRuns)
			}

			// Validate results
			if result.TotalRecords != test.numRecords {
				t.Errorf("Expected %d records, got %d", test.numRecords, result.TotalRecords)
			}

			memoryUsed := finalStats.AllocBytes - baselineStats.AllocBytes
			recordsPerSecond := float64(test.numRecords) / duration.Seconds()

			t.Logf("GC Test %s Results:", test.name)
			t.Logf("  Records: %d", test.numRecords)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Records/sec: %.2f", recordsPerSecond)
			t.Logf("  Memory used: %s", formatBytes(memoryUsed))
			t.Logf("  GC runs: %d", gcRuns)
			t.Logf("  GC runs per 1K records: %.2f", float64(gcRuns)/float64(test.numRecords)*1000)
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "csv-h3-alloc-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	numRecords := 1000
	inputFile := filepath.Join(tempDir, "alloc_bench.csv")
	createLargeTestFile(&testing.T{}, inputFile, numRecords)

	cfg := &config.Config{
		InputFile:  inputFile,
		OutputFile: filepath.Join(tempDir, "alloc_output.csv"),
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
		cfg.OutputFile = filepath.Join(tempDir, fmt.Sprintf("alloc_output_%d.csv", i))
		
		orchestrator := service.NewOrchestrator(cfg)
		result, err := orchestrator.ProcessFile()
		if err != nil {
			b.Fatalf("Allocation benchmark failed: %v", err)
		}

		if result.TotalRecords != numRecords {
			b.Errorf("Expected %d records, got %d", numRecords, result.TotalRecords)
		}
	}
}