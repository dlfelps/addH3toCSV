# CSV H3 Tool - Test Suite

This directory contains comprehensive tests for the CSV H3 Tool, including integration tests, performance tests, and benchmarks.

## Test Structure

```
test/
├── integration/           # Integration tests
│   ├── integration_test.go       # End-to-end workflow tests
│   ├── comprehensive_test.go     # Comprehensive scenario tests
│   ├── benchmark_test.go         # Integration benchmarks
│   └── test_runner.go           # Test utilities
├── performance/          # Performance and memory tests
│   ├── memory_test.go           # Memory usage tests
│   ├── h3_performance_test.go   # H3 generation performance
│   └── streaming_test.go        # Streaming performance tests
├── testdata/            # Test data files
│   ├── edge_cases.csv
│   ├── malformed.csv
│   ├── different_delimiters.csv
│   └── tab_delimited.csv
└── run_tests.go         # Test runner utility
```

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests Only
```bash
make test-integration
```

### Performance Tests Only
```bash
make test-performance
```

### Short Mode (Skip Long-Running Tests)
```bash
make test-short
```

### Benchmarks
```bash
make bench
```

## Test Categories

### Integration Tests (`test/integration/`)

**End-to-End Workflow Tests** (`integration_test.go`)
- Tests complete workflow from CLI to output
- Validates different configuration scenarios
- Tests error handling and edge cases
- Verifies output format preservation

**Comprehensive Scenario Tests** (`comprehensive_test.go`)
- Real-world city coordinates
- Mixed valid/invalid data
- Boundary coordinates (poles, meridians)
- High precision coordinates
- Custom column configurations
- Different H3 resolutions
- Whitespace handling

**CLI Integration Tests**
- Command-line argument parsing
- Flag validation
- Help output
- Error reporting

### Performance Tests (`test/performance/`)

**Memory Usage Tests** (`memory_test.go`)
- Memory scaling with file size
- Streaming memory efficiency
- Memory leak detection
- Garbage collection behavior

**H3 Performance Tests** (`h3_performance_test.go`)
- H3 generation performance across resolutions
- Batch processing performance
- Coordinate validation performance
- Concurrent generation testing

**Streaming Performance Tests** (`streaming_test.go`)
- Streaming throughput testing
- Memory constancy validation
- Error handling performance
- Large file processing

## Test Data

### Sample Files
- `edge_cases.csv` - High precision, scientific notation, boundary values
- `malformed.csv` - Various malformed CSV scenarios
- `different_delimiters.csv` - Semicolon-delimited data
- `tab_delimited.csv` - Tab-delimited data

### Generated Test Data
Tests automatically generate various test scenarios:
- Large files (1K-100K records)
- Mixed valid/invalid coordinates
- Different error rates
- Boundary conditions

## Benchmarks

### H3 Generation Benchmarks
- Performance across different resolutions (0-15)
- Batch processing efficiency
- Memory allocation patterns

### Integration Benchmarks
- End-to-end processing performance
- Memory usage during processing
- Streaming vs batch processing

### Performance Benchmarks
- Throughput measurements
- Memory scaling analysis
- Concurrent processing performance

## Test Utilities

### Test Runner (`test_runner.go`)
Provides utilities for:
- Creating test files
- Validating output
- Comparing files
- Performance measurement

### Makefile Targets
- `test` - Run all tests
- `test-unit` - Unit tests only
- `test-integration` - Integration tests only
- `test-performance` - Performance tests only
- `bench` - Run benchmarks
- `coverage` - Generate coverage report

## Performance Expectations

### Throughput
- **Small files (1K records)**: >200 records/sec
- **Medium files (10K records)**: >300 records/sec
- **Large files (50K+ records)**: >300 records/sec

### Memory Usage
- **Streaming processing**: <200MB regardless of file size
- **Memory scaling**: Should not scale linearly with file size
- **Memory leaks**: No significant growth over multiple iterations

### H3 Generation
- **Resolution 8**: >10,000 coordinates/sec
- **Resolution 15**: >5,000 coordinates/sec
- **Validation**: >50,000 coordinates/sec

## Test Coverage

The test suite covers:
- ✅ All CLI options and flags
- ✅ All supported CSV formats
- ✅ All H3 resolution levels (0-15)
- ✅ Error handling and recovery
- ✅ Memory efficiency
- ✅ Performance characteristics
- ✅ Edge cases and boundary conditions
- ✅ Concurrent processing
- ✅ Large file handling

## Running Specific Tests

### Run a specific test
```bash
go test -v -run TestEndToEndWorkflow ./test/integration/
```

### Run benchmarks for specific component
```bash
go test -bench=BenchmarkH3Generation ./test/performance/
```

### Run with coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run performance tests (not in short mode)
```bash
go test -v ./test/performance/
```

## Continuous Integration

The test suite is designed to run in CI environments:
- Uses temporary directories for isolation
- Cleans up resources automatically
- Provides detailed logging and reporting
- Supports timeout configuration
- Validates performance requirements

## Contributing

When adding new features:
1. Add unit tests in the appropriate `internal/` package
2. Add integration tests for end-to-end scenarios
3. Add performance tests for performance-critical features
4. Update this README if adding new test categories
5. Ensure all tests pass before submitting PR

## Troubleshooting

### Tests Taking Too Long
- Use `-short` flag to skip long-running tests
- Run specific test categories instead of all tests
- Check system resources (memory, CPU)

### Memory Test Failures
- Ensure sufficient system memory
- Close other applications during testing
- Check for memory leaks in recent changes

### Performance Test Failures
- Run on a dedicated test machine
- Ensure consistent system load
- Check performance regression against baseline