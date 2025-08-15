package performance

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"csv-h3-tool/internal/h3"
)

// CoordinatePair represents a latitude/longitude pair
type CoordinatePair struct {
	Lat float64
	Lng float64
}

// generateTestCoordinates generates test coordinates for benchmarking
func generateTestCoordinates(count int) []CoordinatePair {
	// Use the new random source (Go 1.20+)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	coords := make([]CoordinatePair, count)
	
	for i := 0; i < count; i++ {
		// Generate realistic coordinates covering the globe
		lat := (r.Float64() * 180) - 90   // -90 to 90
		lng := (r.Float64() * 360) - 180  // -180 to 180
		coords[i] = CoordinatePair{Lat: lat, Lng: lng}
	}
	
	return coords
}

// generateSpecificCoordinates generates coordinates for specific test scenarios
func generateSpecificCoordinates() map[string][]CoordinatePair {
	return map[string][]CoordinatePair{
		"major_cities": {
			{40.7128, -74.0060},  // New York
			{51.5074, -0.1278},   // London
			{35.6762, 139.6503},  // Tokyo
			{48.8566, 2.3522},    // Paris
			{-33.8688, 151.2093}, // Sydney
			{55.7558, 37.6176},   // Moscow
			{19.0760, 72.8777},   // Mumbai
			{-23.5505, -46.6333}, // São Paulo
		},
		"boundary_cases": {
			{90.0, 0.0},     // North Pole
			{-90.0, 0.0},    // South Pole
			{0.0, 0.0},      // Equator/Prime Meridian
			{0.0, 180.0},    // Antimeridian East
			{0.0, -180.0},   // Antimeridian West
			{89.999, 179.999},   // Near max boundaries
			{-89.999, -179.999}, // Near min boundaries
		},
		"high_precision": {
			{40.123456789012345, -74.987654321098765},
			{51.507400000000000, -0.127800000000000},
			{35.676200000000001, 139.650300000000002},
		},
	}
}

// BenchmarkH3Generation benchmarks H3 index generation at different resolutions
func BenchmarkH3Generation(b *testing.B) {
	generator := h3.NewH3Generator()
	coords := generateTestCoordinates(1000)
	
	resolutions := []int{0, 3, 6, 8, 10, 12, 15}
	
	for _, resolution := range resolutions {
		b.Run(fmt.Sprintf("Resolution_%d", resolution), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				coord := coords[i%len(coords)]
				_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(resolution))
				if err != nil {
					b.Fatalf("H3 generation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkH3GenerationBatch benchmarks batch H3 generation
func BenchmarkH3GenerationBatch(b *testing.B) {
	generator := h3.NewH3Generator()
	
	batchSizes := []int{10, 100, 1000, 10000}
	
	for _, batchSize := range batchSizes {
		coords := generateTestCoordinates(batchSize)
		
		b.Run(fmt.Sprintf("Batch_%d", batchSize), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				for _, coord := range coords {
					_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(8))
					if err != nil {
						b.Fatalf("Batch H3 generation failed: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkH3GenerationScenarios benchmarks specific coordinate scenarios
func BenchmarkH3GenerationScenarios(b *testing.B) {
	generator := h3.NewH3Generator()
	scenarios := generateSpecificCoordinates()
	
	for scenarioName, coords := range scenarios {
		b.Run(scenarioName, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				coord := coords[i%len(coords)]
				_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(8))
				if err != nil {
					b.Fatalf("Scenario %s H3 generation failed: %v", scenarioName, err)
				}
			}
		})
	}
}

// TestH3GenerationPerformance tests H3 generation performance characteristics
func TestH3GenerationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping H3 performance tests in short mode")
	}

	generator := h3.NewH3Generator()
	
	performanceTests := []struct {
		name           string
		numCoords      int
		resolution     int
		maxDuration    time.Duration
		minThroughput  float64 // coordinates per second
	}{
		{"Small_1K_Res8", 1000, 8, 1 * time.Second, 500},
		{"Medium_10K_Res8", 10000, 8, 10 * time.Second, 500},
		{"Large_100K_Res8", 100000, 8, 60 * time.Second, 500},
		{"Small_1K_Res15", 1000, 15, 2 * time.Second, 250},
		{"Medium_10K_Res15", 10000, 15, 20 * time.Second, 250},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			coords := generateTestCoordinates(test.numCoords)
			
			start := time.Now()
			successCount := 0
			
			for _, coord := range coords {
				_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(test.resolution))
				if err != nil {
					t.Errorf("H3 generation failed for coord (%.6f, %.6f): %v", 
						coord.Lat, coord.Lng, err)
				} else {
					successCount++
				}
			}
			
			duration := time.Since(start)
			throughput := float64(successCount) / duration.Seconds()
			
			// Validate performance requirements
			if duration > test.maxDuration {
				t.Errorf("Performance test %s took too long: %v (max: %v)", 
					test.name, duration, test.maxDuration)
			}
			
			if throughput < test.minThroughput {
				t.Errorf("Performance test %s throughput too low: %.2f coords/sec (min: %.2f)", 
					test.name, throughput, test.minThroughput)
			}
			
			t.Logf("H3 Performance Test %s Results:", test.name)
			t.Logf("  Coordinates: %d", test.numCoords)
			t.Logf("  Resolution: %d", test.resolution)
			t.Logf("  Duration: %v", duration)
			t.Logf("  Throughput: %.2f coords/sec", throughput)
			t.Logf("  Success rate: %.2f%%", float64(successCount)/float64(test.numCoords)*100)
		})
	}
}

// TestH3ResolutionPerformanceComparison compares performance across resolutions
func TestH3ResolutionPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping H3 resolution comparison tests in short mode")
	}

	generator := h3.NewH3Generator()
	coords := generateTestCoordinates(10000)
	resolutions := []int{0, 3, 6, 8, 10, 12, 15}
	
	results := make(map[int]time.Duration)
	
	for _, resolution := range resolutions {
		start := time.Now()
		
		for _, coord := range coords {
			_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(resolution))
			if err != nil {
				t.Errorf("H3 generation failed at resolution %d: %v", resolution, err)
			}
		}
		
		duration := time.Since(start)
		results[resolution] = duration
		
		throughput := float64(len(coords)) / duration.Seconds()
		t.Logf("Resolution %d: %v (%.2f coords/sec)", resolution, duration, throughput)
	}
	
	// Analyze performance scaling
	baselineRes := 8
	baselineDuration := results[baselineRes]
	
	for _, resolution := range resolutions {
		if resolution == baselineRes {
			continue
		}
		
		duration := results[resolution]
		ratio := float64(duration) / float64(baselineDuration)
		
		t.Logf("Resolution %d vs %d performance ratio: %.2fx", 
			resolution, baselineRes, ratio)
		
		// Higher resolutions should not be dramatically slower
		if resolution > baselineRes && ratio > 3.0 {
			t.Errorf("Resolution %d is significantly slower than baseline (%.2fx)", 
				resolution, ratio)
		}
	}
}

// BenchmarkH3ValidationPerformance benchmarks coordinate validation performance
func BenchmarkH3ValidationPerformance(b *testing.B) {
	generator := h3.NewH3Generator()
	
	// Mix of valid and invalid coordinates
	testCases := []struct {
		name   string
		coords []CoordinatePair
	}{
		{
			"valid_coords",
			[]CoordinatePair{
				{40.7128, -74.0060},
				{51.5074, -0.1278},
				{35.6762, 139.6503},
			},
		},
		{
			"invalid_coords",
			[]CoordinatePair{
				{91.0, 0.0},    // Invalid lat
				{0.0, 181.0},   // Invalid lng
				{-91.0, -181.0}, // Both invalid
			},
		},
		{
			"boundary_coords",
			[]CoordinatePair{
				{90.0, 180.0},   // Max valid
				{-90.0, -180.0}, // Min valid
				{0.0, 0.0},      // Origin
			},
		},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				coord := tc.coords[i%len(tc.coords)]
				// Just validate, don't generate H3 index
				err := generator.ValidateCoordinates(coord.Lat, coord.Lng)
				_ = err // Ignore error for benchmark
			}
		})
	}
}

// TestH3ConcurrentGeneration tests concurrent H3 generation performance
func TestH3ConcurrentGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent H3 tests in short mode")
	}

	generator := h3.NewH3Generator()
	coords := generateTestCoordinates(10000)
	
	concurrencyLevels := []int{1, 2, 4, 8, 16}
	
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			start := time.Now()
			
			// Divide work among goroutines
			coordsPerWorker := len(coords) / concurrency
			done := make(chan bool, concurrency)
			errors := make(chan error, concurrency)
			
			for i := 0; i < concurrency; i++ {
				startIdx := i * coordsPerWorker
				endIdx := startIdx + coordsPerWorker
				if i == concurrency-1 {
					endIdx = len(coords) // Last worker takes remaining coords
				}
				
				go func(workerCoords []CoordinatePair) {
					defer func() { done <- true }()
					
					for _, coord := range workerCoords {
						_, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(8))
						if err != nil {
							errors <- err
							return
						}
					}
				}(coords[startIdx:endIdx])
			}
			
			// Wait for all workers to complete
			for i := 0; i < concurrency; i++ {
				select {
				case <-done:
					// Worker completed successfully
				case err := <-errors:
					t.Errorf("Concurrent H3 generation failed: %v", err)
				case <-time.After(30 * time.Second):
					t.Fatalf("Concurrent test timed out")
				}
			}
			
			duration := time.Since(start)
			throughput := float64(len(coords)) / duration.Seconds()
			
			t.Logf("Concurrent H3 Test (concurrency=%d):", concurrency)
			t.Logf("  Coordinates: %d", len(coords))
			t.Logf("  Duration: %v", duration)
			t.Logf("  Throughput: %.2f coords/sec", throughput)
		})
	}
}

// BenchmarkH3MemoryUsage benchmarks memory usage during H3 generation
func BenchmarkH3MemoryUsage(b *testing.B) {
	generator := h3.NewH3Generator()
	coords := generateTestCoordinates(1000)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, coord := range coords {
			h3Index, err := generator.Generate(coord.Lat, coord.Lng, h3.H3Resolution(8))
			if err != nil {
				b.Fatalf("H3 generation failed: %v", err)
			}
			// Use the result to prevent optimization
			_ = h3Index
		}
	}
}