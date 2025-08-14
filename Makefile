# CSV H3 Tool - Makefile
# =======================

.PHONY: help build test test-unit test-integration test-performance test-all bench coverage clean install

# Default target
help:
	@echo "CSV H3 Tool - Available Commands"
	@echo "================================"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  install        Install the application"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Test Commands:"
	@echo "  test           Run all tests"
	@echo "  test-unit      Run unit tests only"
	@echo "  test-integration Run integration tests only"
	@echo "  test-performance Run performance tests only"
	@echo "  test-short     Run tests in short mode"
	@echo "  bench          Run benchmarks"
	@echo "  coverage       Generate coverage report"
	@echo ""
	@echo "Development Commands:"
	@echo "  fmt            Format code"
	@echo "  lint           Run linter"
	@echo "  vet            Run go vet"
	@echo "  mod-tidy       Tidy go modules"
	@echo ""
	@echo "Examples:"
	@echo "  make test-integration  # Run integration tests"
	@echo "  make bench            # Run performance benchmarks"
	@echo "  make coverage         # Generate coverage report"

# Build commands
build:
	@echo "Building CSV H3 Tool..."
	go build -o csv-h3-tool ./cmd

install:
	@echo "Installing CSV H3 Tool..."
	go install ./cmd

clean:
	@echo "Cleaning build artifacts..."
	rm -f csv-h3-tool
	rm -f csv-h3-tool.exe
	rm -f coverage.out
	rm -f coverage.html
	rm -rf test/tmp/

# Test commands
test: test-unit test-integration test-performance

test-unit:
	@echo "Running unit tests..."
	go test -v -race ./internal/...

test-integration:
	@echo "Running integration tests..."
	go test -v -timeout=10m ./test/integration/...

test-performance:
	@echo "Running performance tests..."
	go test -v -timeout=15m ./test/performance/...

test-short:
	@echo "Running tests in short mode..."
	go test -short -v ./...

test-all:
	@echo "Running all tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...

# Test visualization
test-dashboard:
	@echo "Generating test dashboard..."
	@chmod +x scripts/run-dashboard.sh
	@./scripts/run-dashboard.sh

test-report:
	@echo "Generating HTML test report..."
	go run scripts/generate-test-report.go

test-watch:
	@echo "Watching tests (requires entr)..."
	find . -name "*.go" | entr -c make test-short

# Benchmark commands
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -v ./test/integration/...
	go test -bench=. -benchmem -v ./test/performance/...
	go test -bench=. -benchmem -v ./internal/h3/...
	go test -bench=. -benchmem -v ./internal/csv/...

bench-h3:
	@echo "Running H3 benchmarks..."
	go test -bench=BenchmarkH3 -benchmem -v ./test/performance/...

bench-memory:
	@echo "Running memory benchmarks..."
	go test -bench=BenchmarkMemory -benchmem -v ./test/performance/...

bench-streaming:
	@echo "Running streaming benchmarks..."
	go test -bench=BenchmarkStreaming -benchmem -v ./test/performance/...

# Coverage commands
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out

coverage-summary:
	@echo "Coverage summary:"
	go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	go tool cover -func=coverage.out | tail -1

# Development commands
fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run

vet:
	@echo "Running go vet..."
	go vet ./...

mod-tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Quality checks
check: fmt vet lint test-short

# Performance profiling
profile-cpu:
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./test/performance/...
	go tool pprof cpu.prof

profile-memory:
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./test/performance/...
	go tool pprof mem.prof

# Test data management
create-test-data:
	@echo "Creating test data files..."
	@mkdir -p test/testdata
	@echo "name,latitude,longitude,description" > test/testdata/sample.csv
	@echo "New York,40.7128,-74.0060,Valid coordinates" >> test/testdata/sample.csv
	@echo "London,51.5074,-0.1278,Valid coordinates" >> test/testdata/sample.csv
	@echo "Invalid,91.0,0.0,Invalid latitude" >> test/testdata/sample.csv

clean-test-data:
	@echo "Cleaning test data..."
	rm -rf test/tmp/
	rm -f test/testdata/output_*.csv

# Integration test scenarios
test-scenarios:
	@echo "Running specific test scenarios..."
	go test -v -run TestComprehensiveScenarios ./test/integration/...

test-error-handling:
	@echo "Testing error handling..."
	go test -v -run TestErrorHandling ./test/integration/...

test-memory-usage:
	@echo "Testing memory usage..."
	go test -v -run TestMemoryUsage ./test/performance/...

# Continuous Integration targets
ci-test:
	@echo "Running CI tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

ci-bench:
	@echo "Running CI benchmarks..."
	go test -bench=. -benchmem -short ./...

# Docker targets (if needed)
docker-build:
	@echo "Building Docker image..."
	docker build -t csv-h3-tool .

docker-test:
	@echo "Running tests in Docker..."
	docker run --rm csv-h3-tool make test

# Release targets
release-build:
	@echo "Building release binaries..."
	GOOS=linux GOARCH=amd64 go build -o csv-h3-tool-linux-amd64 ./cmd
	GOOS=windows GOARCH=amd64 go build -o csv-h3-tool-windows-amd64.exe ./cmd
	GOOS=darwin GOARCH=amd64 go build -o csv-h3-tool-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build -o csv-h3-tool-darwin-arm64 ./cmd

# Documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Development workflow
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev-test:
	@echo "Running development tests..."
	make fmt
	make vet
	make test-short

# Performance monitoring
perf-monitor:
	@echo "Running performance monitoring..."
	go test -bench=. -benchmem -count=5 ./test/performance/... | tee perf-results.txt

# Memory leak detection
test-leaks:
	@echo "Testing for memory leaks..."
	go test -v -run TestMemoryLeaks ./test/performance/...

# Stress testing
stress-test:
	@echo "Running stress tests..."
	go test -v -run TestLargeFileHandling ./test/integration/...
	go test -v -run TestStreamingMemoryEfficiency ./test/performance/...

# Validation
validate:
	@echo "Running validation tests..."
	go test -v -run TestBenchmarkValidation ./test/integration/...
	go test -v -run TestOutputFormatPreservation ./test/integration/...

# Help for specific test categories
help-tests:
	@echo "Test Categories:"
	@echo "==============="
	@echo ""
	@echo "Unit Tests (./internal/...):"
	@echo "  - Validator tests"
	@echo "  - H3 generator tests"
	@echo "  - CSV processor tests"
	@echo "  - CLI tests"
	@echo "  - Configuration tests"
	@echo ""
	@echo "Integration Tests (./test/integration/...):"
	@echo "  - End-to-end workflow tests"
	@echo "  - CLI integration tests"
	@echo "  - Error handling tests"
	@echo "  - File format preservation tests"
	@echo ""
	@echo "Performance Tests (./test/performance/...):"
	@echo "  - Memory usage tests"
	@echo "  - H3 generation performance"
	@echo "  - Streaming efficiency tests"
	@echo "  - Benchmark tests"