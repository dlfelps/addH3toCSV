package main

import (
	"csv-h3-tool/internal/validator"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Open the test CSV file
	file, err := os.Open("test_coordinates.csv")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		return
	}

	// Create validator
	v := validator.NewCoordinateValidator()

	fmt.Println("Testing Coordinate Validator with test_coordinates.csv")
	fmt.Println(strings.Repeat("=", 60))

	validCount := 0
	invalidCount := 0

	// Process each record (skip header)
	for i, record := range records {
		if i == 0 {
			fmt.Printf("%-20s %-12s %-12s %-30s %s\n", "Name", "Latitude", "Longitude", "Description", "Validation Result")
			fmt.Println(strings.Repeat("-", 100))
			continue
		}

		if len(record) < 4 {
			fmt.Printf("Row %d: Insufficient columns\n", i+1)
			invalidCount++
			continue
		}

		name := record[0]
		latStr := record[1]
		lngStr := record[2]
		description := record[3]

		// Parse and validate latitude
		lat, latErr := validator.ParseAndValidateCoordinate(latStr, "latitude")
		
		// Parse and validate longitude
		lng, lngErr := validator.ParseAndValidateCoordinate(lngStr, "longitude")

		// Check overall validation
		var result string
		if latErr != nil {
			result = fmt.Sprintf("❌ Lat Error: %s", latErr.Error())
			invalidCount++
		} else if lngErr != nil {
			result = fmt.Sprintf("❌ Lng Error: %s", lngErr.Error())
			invalidCount++
		} else {
			// Final coordinate validation
			if coordErr := v.ValidateCoordinates(lat, lng); coordErr != nil {
				result = fmt.Sprintf("❌ Coord Error: %s", coordErr.Error())
				invalidCount++
			} else {
				result = fmt.Sprintf("✅ Valid (%.6f, %.6f)", lat, lng)
				validCount++
			}
		}

		fmt.Printf("%-20s %-12s %-12s %-30s %s\n", 
			truncate(name, 20), 
			truncate(latStr, 12), 
			truncate(lngStr, 12), 
			truncate(description, 30), 
			result)
	}

	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("Summary: %d valid coordinates, %d invalid coordinates\n", validCount, invalidCount)
	fmt.Printf("Success rate: %.1f%%\n", float64(validCount)/float64(validCount+invalidCount)*100)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}