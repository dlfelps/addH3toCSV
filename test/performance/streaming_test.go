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

// StreamingTestResult captures streaming performance metrics
type StreamingTestResult struct {
	Name               string
	RecordCount        int
	Duration           time.Duration
	ThroughputRPS      float64
	PeakMemoryBytes    uint64
	AverageMemoryBytes uint64
	GCRuns             uint32
}

// createStreamingTestFile creates a CSV file optimized for streaming tests
func createStreamingTestFile(t *testing.T, filePath string, numRecords int, includeInvalid bool) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create streaming test file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	headers := []string{"record_id", "latitude", "longitude", "location_name", "category", "timestamp", "metadata"}
	if err := writer.Write(headers); err != nil {
		t.Fatalf("Failed to write headers: %v", err)
	}

	// Generate records with realistic data distribution
	for i := 0; i < numRecords; i++ {
		var lat, lng string
		
		if includeInvalid && i%10 == 0 {
			// 10% invalid records
			if i%20 == 0 {
				lat = "91.0" // Invalid latitude
				lng = "0.0"
			} else {
				lat = "0.0"
				lng = "181.0" // Invalid longitude
			}
		} else {
			// Valid coordinates distributed globally
			latVal := float64((i%180)-90) + (float64(i%1000)/1000.0)
			lngVal := float64((i%360)-180) + (float64(i%1000)/1000.0)
			lat = fmt.Sprintf("%.6f", latVal)
			lng = fmt.Sprintf("%.6f", lngVal)
		}

		record := []string{
			fmt.Sprintf("REC_%08d", i),
			lat,
			lng,
			fmt.Sprintf("Location_%d", i),
			fmt.Sprintf("Category_%d", i%20),
			fmt.Sprintf("2024-01-%02d %02d:%02d:%02d", 
				(i%28)+1, (i%24), (i%60), (i%60)),
			fmt.Sprintf("metadata_value_%d", i),
		}

		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}
}

// monitorMemoryUsage monitors memory usage during processing
func monitorMemoryUsage(done chan bool, results chan uint64) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			results <- m.Alloc
		}
	}
}

// TestStreamingPerformance tests streaming processing performance
func TestStreamingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming performance tests in short mode")
	}
	
	// Set a reasonable timeout for the test
	if deadline, ok := t.Deadline(); ok {
		remaining := deadline.Sub(time.Now())
		if remaining < 60*time.Second {
			t.Skip("Insufficient time remaining for streaming performance test")
		}
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-streaming-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	streamingTests := []struct {
		name           string
		numRecords     int
		includeInvalid bool
		maxDuration    time.Duration
		minThroughput  float64 // records per second
		maxMemoryMB    float64
	}{
		{"Small_1K", 1000, false, 10 * time.Second, 50, 100},
		{"Medium_10K", 10000, false, 60 * time.Second, 100, 200},
		{"Large_50K", 50000, false, 5 * time.Minute, 100, 400},
		{"XLarge_100K", 100000, false, 10 * time.Minute, 100, 600},
		{"Mixed_10K", 10000, true, 60 * time.Second, 50, 200},
		{"Mixed_50K", 50000, true, 5 * time.Minute, 50, 400},
	}

	results := make([]StreamingTestResult, 0, len(streamingTests))

	for _, test := range streamingTests {
		t.Run(test.name, func(t *testing.T) {
			// Create test file
			inputFile := filepath.Join(tempDir, fmt.Sprintf("streaming_%s.csv", test.name))
			createStreamingTestFile(t, inputFile, test.numRecords, test.includeInvalid)

			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, fmt.Sprintf("streaming_output_%s.csv", test.name)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			// Start memory monitoring
			memoryReadings := make(chan uint64, 1000)
			monitorDone := make(chan bool)
			go monitorMemoryUsage(monitorDone, memoryReadings)

			// Force GC before test
			runtime.GC()
			runtime.GC()
			baselineStats := getMemoryStats()

			// Process the file
			start := time.Now()
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			duration := time.Since(start)

			// Stop memory monitoring
			close(monitorDone)
			close(memoryReadings)

			if err != nil {
				t.Fatalf("Streaming test %s failed: %v", test.name, err)
			}

			// Collect memory readings
			var memoryValues []uint64
			for reading := range memoryReadings {
				memoryValues = append(memoryValues, reading)
			}

			// Calculate memory statistics
			var peakMemory, totalMemory uint64
			for _, mem := range memoryValues {
				if mem > peakMemory {
					peakMemory = mem
				}
				totalMemory += mem
			}

			var avgMemory uint64
			if len(memoryValues) > 0 {
				avgMemory = totalMemory / uint64(len(memoryValues))
			}

			finalStats := getMemoryStats()
			gcRuns := finalStats.NumGC - baselineStats.NumGC

			// Calculate performance metrics
			throughput := float64(result.TotalRecords) / duration.Seconds()
			peakMemoryMB := float64(peakMemory) / (1024 * 1024)

			// Create test result
			testResult := StreamingTestResult{
				Name:               test.name,
				RecordCount:        result.TotalRecords,
				Duration:           duration,
				ThroughputRPS:      throughput,
				PeakMemoryBytes:    peakMemory,
				AverageMemoryBytes: avgMemory,
				GCRuns:             gcRuns,
			}
			results = append(results, testResult)

			// Validate performance requirements
			if duration > test.maxDuration {
				t.Errorf("Test %s took too long: %v (max: %v)", 
					test.name, duration, test.maxDuration)
			}

			if throughput < test.minThroughput {
				t.Errorf("Test %s throughput too low: %.2f RPS (min: %.2f)", 
					test.name, throughput, test.minThroughput)
			}

			if peakMemoryMB > test.maxMemoryMB {
				t.Errorf("Test %s memory usage too high: %.2f MB (max: %.2f MB)", 
					test.name, peakMemoryMB, test.maxMemoryMB)
			}

			// Validate record counts
			if result.TotalRecords != test.numRecords {
				t.Errorf("Expected %d records, got %d", test.numRecords, result.TotalRecords)
			}

			t.Logf("Streaming Test %s Results:", test.name)
			t.Logf("  Records: %d", result.TotalRecords)
			t.Logf("  Valid: %d", result.ValidRecords)
			t.Logf("  Invalid: %d", result.InvalidRecords)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Throughput: %.2f RPS", throughput)
			t.Logf("  Peak Memory: %s (%.2f MB)", formatBytes(peakMemory), peakMemoryMB)
			t.Logf("  Avg Memory: %s", formatBytes(avgMemory))
			t.Logf("  GC Runs: %d", gcRuns)
			t.Logf("  Memory/Record: %.2f bytes", float64(peakMemory)/float64(result.TotalRecords))
		})
	}

	// Print summary comparison
	t.Logf("\nStreaming Performance Summary:")
	t.Logf("%-15s %10s %12s %15s %12s %8s", 
		"Test", "Records", "Duration", "Throughput", "Peak Mem", "GC Runs")
	t.Logf("%-15s %10s %12s %15s %12s %8s", 
		"----", "-------", "--------", "----------", "--------", "-------")
	
	for _, result := range results {
		t.Logf("%-15s %10d %12v %12.2f RPS %9s %8d", 
			result.Name, 
			result.RecordCount, 
			result.Duration, 
			result.ThroughputRPS,
			formatBytes(result.PeakMemoryBytes),
			result.GCRuns)
	}
}

// TestStreamingMemoryConstancy tests that memory usage remains constant regardless of file size
func TestStreamingMemoryConstancy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming memory constancy tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-memory-constancy-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with different file sizes to ensure memory usage doesn't scale linearly
	fileSizes := []int{1000, 5000, 10000, 25000, 50000}
	memoryUsages := make([]uint64, len(fileSizes))

	for i, size := range fileSizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			// Create test file
			inputFile := filepath.Join(tempDir, fmt.Sprintf("constancy_%d.csv", size))
			createStreamingTestFile(t, inputFile, size, false)

			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, fmt.Sprintf("constancy_output_%d.csv", size)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    false,
			}

			// Monitor memory during processing
			memoryReadings := make(chan uint64, 1000)
			monitorDone := make(chan bool)
			go monitorMemoryUsage(monitorDone, memoryReadings)

			// Force GC before test
			runtime.GC()
			runtime.GC()

			// Process the file
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()

			// Stop monitoring
			close(monitorDone)
			close(memoryReadings)

			if err != nil {
				t.Fatalf("Memory constancy test failed for size %d: %v", size, err)
			}

			// Find peak memory usage
			var peakMemory uint64
			for reading := range memoryReadings {
				if reading > peakMemory {
					peakMemory = reading
				}
			}

			memoryUsages[i] = peakMemory

			if result.TotalRecords != size {
				t.Errorf("Expected %d records, got %d", size, result.TotalRecords)
			}

			t.Logf("Size %d: Peak memory %s", size, formatBytes(peakMemory))
		})
	}

	// Analyze memory scaling
	t.Logf("\nMemory Constancy Analysis:")
	baseSize := fileSizes[0]
	baseMemory := memoryUsages[0]

	for i, size := range fileSizes {
		if i == 0 {
			continue
		}

		memory := memoryUsages[i]
		sizeRatio := float64(size) / float64(baseSize)
		memoryRatio := float64(memory) / float64(baseMemory)

		t.Logf("Size ratio: %.2fx, Memory ratio: %.2fx", sizeRatio, memoryRatio)

		// Memory usage should not scale linearly with file size for streaming
		// Allow some growth but flag if memory scales more than 2x for 50x size increase
		if sizeRatio > 10 && memoryRatio > 3.0 {
			t.Errorf("Memory usage scaling too high: %.2fx memory for %.2fx size increase", 
				memoryRatio, sizeRatio)
		}
	}
}

// BenchmarkStreamingThroughput benchmarks streaming throughput
func BenchmarkStreamingThroughput(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "csv-h3-throughput-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	numRecords := 10000
	inputFile := filepath.Join(tempDir, "throughput_test.csv")
	createStreamingTestFile(&testing.T{}, inputFile, numRecords, false)

	cfg := &config.Config{
		InputFile:  inputFile,
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
		cfg.OutputFile = filepath.Join(tempDir, fmt.Sprintf("throughput_output_%d.csv", i))
		
		orchestrator := service.NewOrchestrator(cfg)
		result, err := orchestrator.ProcessFile()
		if err != nil {
			b.Fatalf("Throughput benchmark failed: %v", err)
		}

		if result.TotalRecords != numRecords {
			b.Errorf("Expected %d records, got %d", numRecords, result.TotalRecords)
		}
	}
}

// TestStreamingErrorHandlingPerformance tests performance when handling errors
func TestStreamingErrorHandlingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping streaming error handling performance tests in short mode")
	}

	tempDir, err := os.MkdirTemp("", "csv-h3-error-perf-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	errorRates := []struct {
		name        string
		numRecords  int
		errorRate   float64 // percentage of invalid records
		maxDuration time.Duration
	}{
		{"Low_Error_1pct", 10000, 0.01, 30 * time.Second},
		{"Medium_Error_10pct", 10000, 0.10, 35 * time.Second},
		{"High_Error_25pct", 10000, 0.25, 40 * time.Second},
		{"Very_High_Error_50pct", 10000, 0.50, 50 * time.Second},
	}

	for _, test := range errorRates {
		t.Run(test.name, func(t *testing.T) {
			// Create test file with specified error rate
			inputFile := filepath.Join(tempDir, fmt.Sprintf("error_test_%s.csv", test.name))
			createErrorTestFile(t, inputFile, test.numRecords, test.errorRate)

			cfg := &config.Config{
				InputFile:  inputFile,
				OutputFile: filepath.Join(tempDir, fmt.Sprintf("error_output_%s.csv", test.name)),
				LatColumn:  "latitude",
				LngColumn:  "longitude",
				Resolution: 8,
				HasHeaders: true,
				Overwrite:  true,
				Verbose:    true, // Enable verbose to test error logging performance
			}

			start := time.Now()
			orchestrator := service.NewOrchestrator(cfg)
			result, err := orchestrator.ProcessFile()
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Error handling test %s failed: %v", test.name, err)
			}

			// Validate performance
			if duration > test.maxDuration {
				t.Errorf("Error handling test %s took too long: %v (max: %v)", 
					test.name, duration, test.maxDuration)
			}

			// Validate results
			expectedInvalid := int(float64(test.numRecords) * test.errorRate)
			expectedValid := test.numRecords - expectedInvalid
			_ = expectedValid // Used for validation logic
			
			// Allow some tolerance for rounding
			tolerance := int(float64(test.numRecords) * 0.02) // 2% tolerance
			
			if abs(result.InvalidRecords-expectedInvalid) > tolerance {
				t.Errorf("Expected ~%d invalid records, got %d", expectedInvalid, result.InvalidRecords)
			}

			throughput := float64(result.TotalRecords) / duration.Seconds()

			t.Logf("Error Handling Test %s Results:", test.name)
			t.Logf("  Total Records: %d", result.TotalRecords)
			t.Logf("  Valid Records: %d", result.ValidRecords)
			t.Logf("  Invalid Records: %d", result.InvalidRecords)
			t.Logf("  Error Rate: %.2f%%", float64(result.InvalidRecords)/float64(result.TotalRecords)*100)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Throughput: %.2f RPS", throughput)
		})
	}
}

// createErrorTestFile creates a test file with a specified error rate
func createErrorTestFile(t *testing.T, filePath string, numRecords int, errorRate float64) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create error test file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "latitude", "longitude", "name"}); err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}

	errorCount := int(float64(numRecords) * errorRate)
	errorInterval := numRecords / errorCount
	if errorInterval == 0 {
		errorInterval = 1
	}

	for i := 0; i < numRecords; i++ {
		var lat, lng string
		
		if i%errorInterval == 0 && errorCount > 0 {
			// Generate invalid coordinate
			switch i % 4 {
			case 0:
				lat, lng = "91.0", "0.0"    // Invalid lat
			case 1:
				lat, lng = "0.0", "181.0"   // Invalid lng
			case 2:
				lat, lng = "", ""           // Empty
			case 3:
				lat, lng = "abc", "xyz"     // Malformed
			}
			errorCount--
		} else {
			// Generate valid coordinate
			latVal := float64((i%180)-90) + (float64(i%1000)/1000.0)
			lngVal := float64((i%360)-180) + (float64(i%1000)/1000.0)
			lat = fmt.Sprintf("%.6f", latVal)
			lng = fmt.Sprintf("%.6f", lngVal)
		}

		record := []string{
			fmt.Sprintf("ID_%06d", i),
			lat,
			lng,
			fmt.Sprintf("Location_%d", i),
		}

		if err := writer.Write(record); err != nil {
			t.Fatalf("Failed to write record %d: %v", i, err)
		}
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}