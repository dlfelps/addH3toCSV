package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestSuite represents a test suite configuration
type TestSuite struct {
	Name        string
	Path        string
	Description string
	Flags       []string
}

// TestResult represents the result of running a test suite
type TestResult struct {
	Suite    TestSuite
	Success  bool
	Duration time.Duration
	Output   string
	Error    error
}

func main() {
	var (
		verbose     = flag.Bool("v", false, "Verbose output")
		short       = flag.Bool("short", false, "Run tests in short mode")
		integration = flag.Bool("integration", false, "Run integration tests only")
		performance = flag.Bool("performance", false, "Run performance tests only")
		unit        = flag.Bool("unit", false, "Run unit tests only")
		benchmark   = flag.Bool("bench", false, "Run benchmarks")
		coverage    = flag.Bool("coverage", false, "Generate coverage report")
		timeout     = flag.String("timeout", "10m", "Test timeout")
		parallel    = flag.Int("parallel", 4, "Number of parallel test processes")
	)
	flag.Parse()

	fmt.Println("CSV H3 Tool - Test Runner")
	fmt.Println("========================")
	fmt.Println()

	// Define test suites
	testSuites := []TestSuite{
		{
			Name:        "Unit Tests",
			Path:        "./internal/...",
			Description: "Run all unit tests for internal packages",
			Flags:       []string{"-v"},
		},
		{
			Name:        "Integration Tests",
			Path:        "./test/integration",
			Description: "Run integration tests with sample data",
			Flags:       []string{"-v"},
		},
		{
			Name:        "Performance Tests",
			Path:        "./test/performance",
			Description: "Run performance and memory tests",
			Flags:       []string{"-v"},
		},
	}

	// Filter test suites based on flags
	var suitesToRun []TestSuite
	if *unit {
		suitesToRun = append(suitesToRun, testSuites[0])
	} else if *integration {
		suitesToRun = append(suitesToRun, testSuites[1])
	} else if *performance {
		suitesToRun = append(suitesToRun, testSuites[2])
	} else {
		suitesToRun = testSuites
	}

	// Build common flags
	commonFlags := []string{"test"}
	if *verbose {
		commonFlags = append(commonFlags, "-v")
	}
	if *short {
		commonFlags = append(commonFlags, "-short")
	}
	if *benchmark {
		commonFlags = append(commonFlags, "-bench=.")
	}
	if *coverage {
		commonFlags = append(commonFlags, "-coverprofile=coverage.out")
	}
	commonFlags = append(commonFlags, "-timeout", *timeout)
	commonFlags = append(commonFlags, fmt.Sprintf("-parallel=%d", *parallel))

	// Run test suites
	results := make([]TestResult, 0, len(suitesToRun))
	totalStart := time.Now()

	for _, suite := range suitesToRun {
		fmt.Printf("Running %s...\n", suite.Name)
		fmt.Printf("Description: %s\n", suite.Description)
		fmt.Printf("Path: %s\n", suite.Path)
		fmt.Println()

		result := runTestSuite(suite, commonFlags)
		results = append(results, result)

		if result.Success {
			fmt.Printf("✅ %s completed successfully in %v\n", suite.Name, result.Duration)
		} else {
			fmt.Printf("❌ %s failed after %v\n", suite.Name, result.Duration)
			if result.Error != nil {
				fmt.Printf("Error: %v\n", result.Error)
			}
		}

		if *verbose && result.Output != "" {
			fmt.Println("Output:")
			fmt.Println(result.Output)
		}
		fmt.Println()
	}

	totalDuration := time.Since(totalStart)

	// Print summary
	printSummary(results, totalDuration)

	// Exit with error code if any tests failed
	for _, result := range results {
		if !result.Success {
			os.Exit(1)
		}
	}
}

// runTestSuite runs a single test suite
func runTestSuite(suite TestSuite, commonFlags []string) TestResult {
	start := time.Now()

	// Build command
	args := append(commonFlags, suite.Flags...)
	args = append(args, suite.Path)

	cmd := exec.Command("go", args...)
	cmd.Dir = getProjectRoot()

	// Capture output
	output, err := cmd.CombinedOutput()

	return TestResult{
		Suite:    suite,
		Success:  err == nil,
		Duration: time.Since(start),
		Output:   string(output),
		Error:    err,
	}
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Try to find go.mod file
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}

// printSummary prints a summary of all test results
func printSummary(results []TestResult, totalDuration time.Duration) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	successCount := 0
	for _, result := range results {
		status := "FAIL"
		if result.Success {
			status = "PASS"
			successCount++
		}

		fmt.Printf("%-30s %s (%v)\n", result.Suite.Name, status, result.Duration)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Suites: %d\n", len(results))
	fmt.Printf("Passed: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(results)-successCount)
	fmt.Printf("Success Rate: %.1f%%\n", float64(successCount)/float64(len(results))*100)
	fmt.Printf("Total Duration: %v\n", totalDuration)
	fmt.Println(strings.Repeat("=", 80))

	// Additional information
	if successCount == len(results) {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Println("⚠️  Some tests failed. Check the output above for details.")
	}
}

// Additional utility functions for specific test scenarios

// runBenchmarks runs benchmark tests specifically
func runBenchmarks() {
	fmt.Println("Running Benchmarks...")
	fmt.Println("====================")

	benchmarkSuites := []string{
		"./test/integration",
		"./test/performance",
		"./internal/h3",
		"./internal/csv",
	}

	for _, suite := range benchmarkSuites {
		fmt.Printf("Benchmarking %s...\n", suite)
		
		cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-v", suite)
		cmd.Dir = getProjectRoot()
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ Benchmark failed for %s: %v\n", suite, err)
		} else {
			fmt.Printf("✅ Benchmark completed for %s\n", suite)
		}
		
		fmt.Println(string(output))
		fmt.Println()
	}
}

// runCoverageReport generates and displays coverage report
func runCoverageReport() {
	fmt.Println("Generating Coverage Report...")
	fmt.Println("============================")

	// Run tests with coverage
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = getProjectRoot()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Coverage test failed: %v\n", err)
		fmt.Println(string(output))
		return
	}

	// Generate HTML coverage report
	cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	cmd.Dir = getProjectRoot()
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to generate HTML coverage report: %v\n", err)
	} else {
		fmt.Println("✅ HTML coverage report generated: coverage.html")
	}

	// Display coverage summary
	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	cmd.Dir = getProjectRoot()
	
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to generate coverage summary: %v\n", err)
	} else {
		fmt.Println("Coverage Summary:")
		fmt.Println(string(output))
	}
}

// validateTestEnvironment checks if the test environment is properly set up
func validateTestEnvironment() error {
	// Check if Go is installed
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go is not installed or not in PATH")
	}

	// Check if we're in a Go module
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("not in a Go module directory (go.mod not found)")
	}

	// Check if required dependencies are available
	cmd := exec.Command("go", "mod", "verify")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Go module verification failed: %v", err)
	}

	return nil
}