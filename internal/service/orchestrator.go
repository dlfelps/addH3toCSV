package service

import (
	"fmt"
	"os"
	"time"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/csv"
	"csv-h3-tool/internal/h3"
	"csv-h3-tool/internal/validator"
)

// Orchestrator coordinates all components to process CSV files
type Orchestrator struct {
	validator   validator.Validator
	h3Generator h3.Generator
	processor   csv.Processor
	config      *config.Config
}

// h3GeneratorAdapter adapts the h3.Generator interface to work with csv.StreamingProcessor
type h3GeneratorAdapter struct {
	generator h3.Generator
}

func (a *h3GeneratorAdapter) Generate(lat, lng float64, resolution int) (string, error) {
	return a.generator.Generate(lat, lng, h3.H3Resolution(resolution))
}

// NewOrchestrator creates a new orchestrator with all required components
func NewOrchestrator(cfg *config.Config) *Orchestrator {
	validator := validator.NewCoordinateValidator()
	h3Generator := h3.NewH3Generator()
	processor := csv.NewStreamingProcessor(validator, &h3GeneratorAdapter{
		generator: h3Generator,
	})

	return &Orchestrator{
		validator:   validator,
		h3Generator: h3Generator,
		processor:   processor,
		config:      cfg,
	}
}

// ProcessResult contains the results of processing a CSV file
type ProcessResult struct {
	TotalRecords   int
	ValidRecords   int
	InvalidRecords int
	ProcessingTime time.Duration
	OutputFile     string
}

// ProcessFile orchestrates the complete CSV processing workflow
func (o *Orchestrator) ProcessFile() (*ProcessResult, error) {
	startTime := time.Now()

	if o.config.Verbose {
		fmt.Printf("Starting CSV processing...\n")
		fmt.Printf("Input file: %s\n", o.config.InputFile)
		fmt.Printf("Output file: %s\n", o.config.OutputFile)
		fmt.Printf("H3 Resolution: %d (%s)\n", o.config.Resolution, o.config.GetResolutionDescription())
	}

	// Validate configuration
	if err := o.config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Pre-validate CSV structure
	if err := o.validateCSVStructure(); err != nil {
		return nil, fmt.Errorf("CSV structure validation failed: %w", err)
	}

	// Process the file with progress reporting
	result, err := o.processWithProgress()
	if err != nil {
		return nil, fmt.Errorf("file processing failed: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	result.OutputFile = o.config.OutputFile

	if o.config.Verbose {
		fmt.Printf("Processing completed in %v\n", result.ProcessingTime)
		fmt.Printf("Results: %d total, %d valid, %d invalid records\n", 
			result.TotalRecords, result.ValidRecords, result.InvalidRecords)
	}

	return result, nil
}

// validateCSVStructure performs pre-processing validation of the CSV file
func (o *Orchestrator) validateCSVStructure() error {
	// Open the file to read headers
	reader, err := csv.NewReader(o.config.InputFile, csv.Config{
		InputFile:  o.config.InputFile,
		LatColumn:  o.config.LatColumn,
		LngColumn:  o.config.LngColumn,
		HasHeaders: o.config.HasHeaders,
	})
	if err != nil {
		return fmt.Errorf("failed to open CSV file for validation: %w", err)
	}
	defer reader.Close()

	// Validate column configuration
	headers := reader.GetHeaders()
	if err := o.processor.ValidateColumns(headers, csv.Config{
		LatColumn:  o.config.LatColumn,
		LngColumn:  o.config.LngColumn,
		HasHeaders: o.config.HasHeaders,
	}); err != nil {
		return fmt.Errorf("column validation failed: %w", err)
	}

	if o.config.Verbose {
		fmt.Printf("CSV structure validated successfully\n")
		if o.config.HasHeaders {
			fmt.Printf("Headers: %v\n", headers)
			fmt.Printf("Latitude column: %s (index %d)\n", o.config.LatColumn, reader.GetLatIndex())
			fmt.Printf("Longitude column: %s (index %d)\n", o.config.LngColumn, reader.GetLngIndex())
		}
	}

	return nil
}

// processWithProgress processes the CSV file with progress reporting
func (o *Orchestrator) processWithProgress() (*ProcessResult, error) {
	// Get file size for progress calculation
	fileInfo, err := os.Stat(o.config.InputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Create progress reporter
	progressReporter := NewProgressReporter(fileSize, o.config.Verbose)

	// Open input file
	reader, err := csv.NewReader(o.config.InputFile, csv.Config{
		InputFile:  o.config.InputFile,
		LatColumn:  o.config.LatColumn,
		LngColumn:  o.config.LngColumn,
		HasHeaders: o.config.HasHeaders,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer reader.Close()

	// Create output writer
	writer, err := csv.NewWriter(o.config.OutputFile, reader.GetHeaders(), csv.Config{
		OutputFile: o.config.OutputFile,
		HasHeaders: o.config.HasHeaders,
		Overwrite:  o.config.Overwrite,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer writer.Close()

	// Process records with progress tracking
	result := &ProcessResult{}
	
	// Create streaming processor with our components
	streamProcessor := csv.NewStreamingProcessor(o.validator, &h3GeneratorAdapter{
		generator: o.h3Generator,
	})

	// Process the stream with progress reporting
	err = streamProcessor.ProcessStream(reader, csv.Config{
		InputFile:  o.config.InputFile,
		OutputFile: o.config.OutputFile,
		Resolution: o.config.Resolution,
		Verbose:    o.config.Verbose,
	}, func(record *csv.Record) error {
		// Update progress
		progressReporter.UpdateProgress(record.LineNumber)
		
		// Update counters
		result.TotalRecords++
		if record.IsValid {
			result.ValidRecords++
		} else {
			result.InvalidRecords++
		}

		// Write record to output
		return writer.WriteRecord(record)
	})

	if err != nil {
		return nil, fmt.Errorf("stream processing failed: %w", err)
	}

	// Ensure all data is written
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush output: %w", err)
	}

	progressReporter.Complete()
	return result, nil
}

// ProgressReporter handles progress reporting for large file processing
type ProgressReporter struct {
	fileSize      int64
	verbose       bool
	lastReported  time.Time
	reportInterval time.Duration
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(fileSize int64, verbose bool) *ProgressReporter {
	return &ProgressReporter{
		fileSize:       fileSize,
		verbose:        verbose,
		lastReported:   time.Now(),
		reportInterval: 2 * time.Second, // Report progress every 2 seconds
	}
}

// UpdateProgress updates the progress based on current line number
func (p *ProgressReporter) UpdateProgress(lineNumber int) {
	if !p.verbose {
		return
	}

	now := time.Now()
	if now.Sub(p.lastReported) < p.reportInterval {
		return
	}

	// Estimate progress based on line number (rough approximation)
	if lineNumber > 0 && lineNumber%1000 == 0 {
		fmt.Printf("Processed %d records...\n", lineNumber)
		p.lastReported = now
	}
}

// Complete marks progress reporting as complete
func (p *ProgressReporter) Complete() {
	if p.verbose {
		fmt.Printf("Processing complete.\n")
	}
}

// ValidateComponents ensures all required components are properly initialized
func (o *Orchestrator) ValidateComponents() error {
	if o.validator == nil {
		return fmt.Errorf("validator component is not initialized")
	}
	
	if o.h3Generator == nil {
		return fmt.Errorf("H3 generator component is not initialized")
	}
	
	if o.processor == nil {
		return fmt.Errorf("CSV processor component is not initialized")
	}
	
	if o.config == nil {
		return fmt.Errorf("configuration is not initialized")
	}
	
	return nil
}

// GetConfig returns the current configuration
func (o *Orchestrator) GetConfig() *config.Config {
	return o.config
}

// SetConfig updates the configuration
func (o *Orchestrator) SetConfig(cfg *config.Config) {
	o.config = cfg
}