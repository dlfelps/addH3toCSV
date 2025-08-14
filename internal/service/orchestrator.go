package service

import (
	"fmt"
	"os"
	"time"

	"csv-h3-tool/internal/config"
	"csv-h3-tool/internal/csv"
	"csv-h3-tool/internal/errors"
	"csv-h3-tool/internal/h3"
	"csv-h3-tool/internal/logging"
	"csv-h3-tool/internal/validator"
)

// Orchestrator coordinates all components to process CSV files
type Orchestrator struct {
	validator   validator.Validator
	h3Generator h3.Generator
	processor   csv.Processor
	config      *config.Config
	logger      *logging.Logger
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
	logger := logging.NewDefaultLogger(cfg.Verbose)
	
	processor := csv.NewStreamingProcessor(validator, &h3GeneratorAdapter{
		generator: h3Generator,
	})

	return &Orchestrator{
		validator:   validator,
		h3Generator: h3Generator,
		processor:   processor,
		config:      cfg,
		logger:      logger,
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

	o.logger.Info("Starting CSV processing")
	o.logger.Info("Input file: %s", o.config.InputFile)
	o.logger.Info("Output file: %s", o.config.OutputFile)
	o.logger.Info("H3 Resolution: %d (%s)", o.config.Resolution, o.config.GetResolutionDescription())

	// Validate configuration
	if err := o.config.Validate(); err != nil {
		configErr := errors.NewConfigError("", "", "configuration validation failed", err)
		o.logger.LogError(configErr)
		return nil, configErr
	}

	// Pre-validate CSV structure
	if err := o.validateCSVStructure(); err != nil {
		csvErr := errors.NewCSVError(o.config.InputFile, 0, 0, "", "", "CSV structure validation failed", err)
		o.logger.LogError(csvErr)
		return nil, csvErr
	}

	// Process the file with progress reporting
	result, err := o.processWithProgress()
	if err != nil {
		processErr := errors.NewProcessingError("file_processing", 0, "file processing failed", err)
		o.logger.LogError(processErr)
		return nil, processErr
	}

	result.ProcessingTime = time.Since(startTime)
	result.OutputFile = o.config.OutputFile

	// Log processing summary
	o.logger.LogProcessingSummary(result.TotalRecords, result.ValidRecords, result.InvalidRecords, result.ProcessingTime)

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
		return errors.NewFileError(o.config.InputFile, "open", err)
	}
	defer reader.Close()

	// Validate column configuration
	headers := reader.GetHeaders()
	if err := o.processor.ValidateColumns(headers, csv.Config{
		LatColumn:  o.config.LatColumn,
		LngColumn:  o.config.LngColumn,
		HasHeaders: o.config.HasHeaders,
	}); err != nil {
		return errors.NewValidationError("columns", "", 0, "column validation failed", err)
	}

	o.logger.Info("CSV structure validated successfully")
	if o.config.HasHeaders {
		o.logger.Debug("Headers: %v", headers)
		o.logger.Debug("Latitude column: %s (index %d)", o.config.LatColumn, reader.GetLatIndex())
		o.logger.Debug("Longitude column: %s (index %d)", o.config.LngColumn, reader.GetLngIndex())
	}

	return nil
}

// processWithProgress processes the CSV file with progress reporting
func (o *Orchestrator) processWithProgress() (*ProcessResult, error) {
	// Get file info for validation
	_, err := os.Stat(o.config.InputFile)
	if err != nil {
		return nil, errors.NewFileError(o.config.InputFile, "stat", err)
	}

	// Open input file
	reader, err := csv.NewReader(o.config.InputFile, csv.Config{
		InputFile:  o.config.InputFile,
		LatColumn:  o.config.LatColumn,
		LngColumn:  o.config.LngColumn,
		HasHeaders: o.config.HasHeaders,
	})
	if err != nil {
		return nil, errors.NewFileError(o.config.InputFile, "open", err)
	}
	defer reader.Close()

	// Create output writer
	writer, err := csv.NewWriter(o.config.OutputFile, reader.GetHeaders(), csv.Config{
		OutputFile: o.config.OutputFile,
		HasHeaders: o.config.HasHeaders,
		Overwrite:  o.config.Overwrite,
	})
	if err != nil {
		return nil, errors.NewFileError(o.config.OutputFile, "create", err)
	}
	defer writer.Close()

	// Create processing logger
	processLogger := logging.NewProcessingLogger(o.logger, o.config.InputFile, 0)

	// Process records with progress tracking
	result := &ProcessResult{}
	errorCollector := errors.NewErrorCollector(100) // Collect up to 100 errors
	
	// Create streaming processor with our components
	streamProcessor := csv.NewStreamingProcessor(o.validator, &h3GeneratorAdapter{
		generator: o.h3Generator,
	})

	// Process the stream with enhanced error handling
	err = streamProcessor.ProcessStream(reader, csv.Config{
		InputFile:  o.config.InputFile,
		OutputFile: o.config.OutputFile,
		Resolution: o.config.Resolution,
		Verbose:    o.config.Verbose,
	}, func(record *csv.Record) error {
		// Update counters
		result.TotalRecords++
		
		if record.IsValid {
			result.ValidRecords++
			processLogger.LogRecordProcessed(record.LineNumber, true, record.H3Index)
		} else {
			result.InvalidRecords++
			processLogger.LogRecordProcessed(record.LineNumber, false, "")
			
			// Log specific error details if available
			if record.Latitude != 0 || record.Longitude != 0 {
				processLogger.LogCoordinateError(record.LineNumber, record.Latitude, record.Longitude, 
					"coordinates", "invalid coordinate values")
			} else {
				processLogger.LogSkippedRecord(record.LineNumber, "empty or malformed coordinates")
			}
		}

		// Write record to output
		if err := writer.WriteRecord(record); err != nil {
			writeErr := errors.NewFileError(o.config.OutputFile, "write", err)
			errorCollector.Add(writeErr)
			o.logger.LogError(writeErr)
			return writeErr
		}
		
		return nil
	})

	if err != nil {
		return nil, errors.NewProcessingError("stream_processing", 0, "stream processing failed", err)
	}

	// Ensure all data is written
	if err := writer.Flush(); err != nil {
		return nil, errors.NewFileError(o.config.OutputFile, "flush", err)
	}

	// Log completion
	processLogger.Complete(time.Since(time.Now()), result.ValidRecords, result.InvalidRecords)

	// Report collected errors if any
	if errorCollector.HasErrors() {
		o.logger.Warn("Processing completed with %d errors", errorCollector.Count())
		if o.config.Verbose {
			for _, err := range errorCollector.Errors() {
				o.logger.LogError(err)
			}
		}
	}

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
		return errors.NewValidationError("validator", "", 0, "validator component is not initialized", nil)
	}
	
	if o.h3Generator == nil {
		return errors.NewValidationError("h3Generator", "", 0, "H3 generator component is not initialized", nil)
	}
	
	if o.processor == nil {
		return errors.NewValidationError("processor", "", 0, "CSV processor component is not initialized", nil)
	}
	
	if o.config == nil {
		return errors.NewValidationError("config", "", 0, "configuration is not initialized", nil)
	}
	
	if o.logger == nil {
		return errors.NewValidationError("logger", "", 0, "logger component is not initialized", nil)
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