package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the configuration for CSV processing
type Config struct {
	InputFile     string
	OutputFile    string
	LatColumn     string
	LngColumn     string
	Resolution    int  // H3 resolution level (0-15)
	HasHeaders    bool
	Overwrite     bool
	Verbose       bool
}

// Record represents a single CSV record with coordinate data
type Record struct {
	OriginalData []string // All original CSV columns
	Latitude     float64  // Parsed latitude value
	Longitude    float64  // Parsed longitude value
	H3Index      string   // Generated H3 index
	LineNumber   int      // Original line number for error reporting
	IsValid      bool     // Whether record has valid coordinates
}

// Processor defines the interface for CSV file processing
type Processor interface {
	ProcessFile(config Config) error
	ValidateColumns(headers []string, config Config) error
}

// Reader handles CSV file reading with column detection
type Reader struct {
	file      *os.File
	csvReader *csv.Reader
	headers   []string
	latIndex  int
	lngIndex  int
	hasHeaders bool
}

// NewReader creates a new CSV reader
func NewReader(filename string, config Config) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	reader := &Reader{
		file:       file,
		csvReader:  csvReader,
		hasHeaders: config.HasHeaders,
		latIndex:   -1,
		lngIndex:   -1,
	}

	// Read headers if present
	if config.HasHeaders {
		headers, err := csvReader.Read()
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read headers: %w", err)
		}
		reader.headers = headers
	}

	// Detect column indices
	if err := reader.detectColumns(config); err != nil {
		file.Close()
		return nil, err
	}

	return reader, nil
}

// detectColumns identifies latitude and longitude column indices
func (r *Reader) detectColumns(config Config) error {
	// If we have headers, try to find columns by name
	if r.hasHeaders && len(r.headers) > 0 {
		r.latIndex = r.findColumnByName(config.LatColumn, []string{"lat", "latitude", "y"})
		r.lngIndex = r.findColumnByName(config.LngColumn, []string{"lng", "lon", "longitude", "x"})
	} else {
		// Try to parse column specifications as indices
		if latIdx, err := strconv.Atoi(config.LatColumn); err == nil && latIdx >= 0 {
			r.latIndex = latIdx
		}
		if lngIdx, err := strconv.Atoi(config.LngColumn); err == nil && lngIdx >= 0 {
			r.lngIndex = lngIdx
		}
	}

	// Validate that we found both columns
	if r.latIndex == -1 {
		return fmt.Errorf("latitude column not found: %s", config.LatColumn)
	}
	if r.lngIndex == -1 {
		return fmt.Errorf("longitude column not found: %s", config.LngColumn)
	}

	return nil
}

// findColumnByName searches for a column by name with fallback options
func (r *Reader) findColumnByName(specified string, fallbacks []string) int {
	// First try the specified column name
	if specified != "" {
		for i, header := range r.headers {
			if strings.EqualFold(strings.TrimSpace(header), strings.TrimSpace(specified)) {
				return i
			}
		}
	}

	// If not found, try fallback names
	for _, fallback := range fallbacks {
		for i, header := range r.headers {
			if strings.EqualFold(strings.TrimSpace(header), fallback) {
				return i
			}
		}
	}

	return -1
}

// ReadRecord reads the next record from the CSV file
func (r *Reader) ReadRecord() (*Record, error) {
	row, err := r.csvReader.Read()
	if err != nil {
		return nil, err
	}

	// Validate that we have enough columns
	if len(row) <= r.latIndex || len(row) <= r.lngIndex {
		return nil, fmt.Errorf("row has insufficient columns: expected at least %d, got %d", 
			max(r.latIndex, r.lngIndex)+1, len(row))
	}

	record := &Record{
		OriginalData: make([]string, len(row)),
		LineNumber:   int(r.csvReader.InputOffset()),
		IsValid:      false,
	}

	// Copy original data
	copy(record.OriginalData, row)

	// Parse coordinates - we'll validate them later in the processing pipeline
	latStr := strings.TrimSpace(row[r.latIndex])
	lngStr := strings.TrimSpace(row[r.lngIndex])

	if latStr == "" || lngStr == "" {
		return record, nil // Return invalid record for empty coordinates
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return record, nil // Return invalid record for unparseable coordinates
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return record, nil // Return invalid record for unparseable coordinates
	}

	record.Latitude = lat
	record.Longitude = lng
	record.IsValid = true

	return record, nil
}

// GetHeaders returns the CSV headers if available
func (r *Reader) GetHeaders() []string {
	return r.headers
}

// GetLatIndex returns the latitude column index
func (r *Reader) GetLatIndex() int {
	return r.latIndex
}

// GetLngIndex returns the longitude column index
func (r *Reader) GetLngIndex() int {
	return r.lngIndex
}

// Close closes the CSV reader and underlying file
func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ValidateColumns validates that the required columns exist in the headers
func ValidateColumns(headers []string, config Config) error {
	if !config.HasHeaders {
		// For files without headers, validate that column indices are reasonable
		if config.LatColumn == "" || config.LngColumn == "" {
			return fmt.Errorf("column specifications required when HasHeaders is false")
		}
		return nil
	}

	// Create a temporary reader to test column detection
	tempReader := &Reader{
		headers:    headers,
		hasHeaders: true,
	}

	if err := tempReader.detectColumns(config); err != nil {
		return fmt.Errorf("column validation failed: %w", err)
	}

	return nil
}

// StreamingProcessor implements streaming CSV processing
type StreamingProcessor struct {
	validator interface {
		ValidateCoordinates(lat, lng float64) error
	}
	h3Generator interface {
		Generate(lat, lng float64, resolution int) (string, error)
	}
}

// NewStreamingProcessor creates a new streaming processor
func NewStreamingProcessor(validator interface{ ValidateCoordinates(lat, lng float64) error }, 
	h3Generator interface{ Generate(lat, lng float64, resolution int) (string, error) }) *StreamingProcessor {
	return &StreamingProcessor{
		validator:   validator,
		h3Generator: h3Generator,
	}
}

// ProcessStream processes CSV records one by one using streaming
func (p *StreamingProcessor) ProcessStream(reader *Reader, config Config, recordHandler func(*Record) error) error {
	recordCount := 0
	validCount := 0
	errorCount := 0

	for {
		record, err := reader.ReadRecord()
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file reached
			}
			// Handle malformed rows gracefully - log and continue
			errorCount++
			if config.Verbose {
				fmt.Printf("Warning: Skipping malformed row at line %d: %v\n", recordCount+1, err)
			}
			continue
		}

		recordCount++

		// Process valid records
		if record.IsValid {
			// Validate coordinates using the validator
			if p.validator != nil {
				if err := p.validator.ValidateCoordinates(record.Latitude, record.Longitude); err != nil {
					record.IsValid = false
					errorCount++
					if config.Verbose {
						fmt.Printf("Warning: Invalid coordinates at line %d: %v\n", record.LineNumber, err)
					}
				}
			}

			// Generate H3 index for valid coordinates
			if record.IsValid && p.h3Generator != nil {
				h3Index, err := p.h3Generator.Generate(record.Latitude, record.Longitude, config.Resolution)
				if err != nil {
					record.IsValid = false
					errorCount++
					if config.Verbose {
						fmt.Printf("Warning: H3 generation failed at line %d: %v\n", record.LineNumber, err)
					}
				} else {
					record.H3Index = h3Index
					validCount++
				}
			}
		} else {
			errorCount++
			if config.Verbose {
				fmt.Printf("Warning: Skipping invalid record at line %d\n", record.LineNumber)
			}
		}

		// Call the record handler
		if err := recordHandler(record); err != nil {
			return fmt.Errorf("record handler failed at line %d: %w", record.LineNumber, err)
		}
	}

	if config.Verbose {
		fmt.Printf("Processing complete: %d total records, %d valid, %d errors\n", 
			recordCount, validCount, errorCount)
	}

	return nil
}

// ProcessFile implements the Processor interface for streaming processing
func (p *StreamingProcessor) ProcessFile(config Config) error {
	// Open input file
	reader, err := NewReader(config.InputFile, config)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer reader.Close()

	// Create output writer
	writer, err := NewWriter(config.OutputFile, reader.GetHeaders(), config)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer writer.Close()

	// Process records using streaming
	return p.ProcessStream(reader, config, func(record *Record) error {
		return writer.WriteRecord(record)
	})
}

// ValidateColumns implements the Processor interface
func (p *StreamingProcessor) ValidateColumns(headers []string, config Config) error {
	return ValidateColumns(headers, config)
}

// Writer handles CSV file writing with H3 index column
type Writer struct {
	file      *os.File
	csvWriter *csv.Writer
	headers   []string
	config    Config
}

// NewWriter creates a new CSV writer
func NewWriter(filename string, inputHeaders []string, config Config) (*Writer, error) {
	// Check if output file exists and handle overwrite
	if _, err := os.Stat(filename); err == nil && !config.Overwrite {
		return nil, fmt.Errorf("output file %s already exists (use overwrite option to replace)", filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file %s: %w", filename, err)
	}

	csvWriter := csv.NewWriter(file)

	// Prepare headers - add H3 index column as the last column
	var headers []string
	if inputHeaders != nil {
		headers = make([]string, len(inputHeaders)+1)
		copy(headers, inputHeaders)
		headers[len(inputHeaders)] = "h3_index"
	}

	writer := &Writer{
		file:      file,
		csvWriter: csvWriter,
		headers:   headers,
		config:    config,
	}

	// Write headers if present
	if config.HasHeaders && headers != nil {
		if err := csvWriter.Write(headers); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	return writer, nil
}

// WriteRecord writes a record to the CSV file
func (w *Writer) WriteRecord(record *Record) error {
	if record == nil {
		return fmt.Errorf("record is nil")
	}

	// Prepare output row - original data plus H3 index
	outputRow := make([]string, len(record.OriginalData)+1)
	copy(outputRow, record.OriginalData)
	
	// Add H3 index as the last column
	if record.IsValid && record.H3Index != "" {
		outputRow[len(record.OriginalData)] = record.H3Index
	} else {
		outputRow[len(record.OriginalData)] = "" // Empty H3 index for invalid records
	}

	if err := w.csvWriter.Write(outputRow); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	return nil
}

// WriteRecords writes multiple records to the CSV file
func (w *Writer) WriteRecords(records []*Record) error {
	for _, record := range records {
		if err := w.WriteRecord(record); err != nil {
			return err
		}
	}
	return nil
}

// Flush flushes any buffered data to the underlying file
func (w *Writer) Flush() error {
	w.csvWriter.Flush()
	return w.csvWriter.Error()
}

// Close closes the CSV writer and underlying file
func (w *Writer) Close() error {
	if w.csvWriter != nil {
		w.csvWriter.Flush()
		if err := w.csvWriter.Error(); err != nil {
			w.file.Close()
			return fmt.Errorf("error flushing CSV writer: %w", err)
		}
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// GetHeaders returns the output headers including H3 index column
func (w *Writer) GetHeaders() []string {
	return w.headers
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}