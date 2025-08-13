package h3

import (
	"fmt"
	"csv-h3-tool/internal/validator"
	"github.com/uber/h3-go/v4"
)

// H3Resolution represents the H3 resolution level (0-15)
type H3Resolution int

const (
	// H3 resolution levels with approximate edge lengths
	ResolutionCountry    H3Resolution = 0  // ~1107.71 km
	ResolutionState      H3Resolution = 1  // ~418.68 km
	ResolutionMetro      H3Resolution = 2  // ~158.24 km
	ResolutionCity       H3Resolution = 3  // ~59.81 km
	ResolutionDistrict   H3Resolution = 4  // ~22.61 km
	ResolutionNeighbor   H3Resolution = 5  // ~8.54 km
	ResolutionBlock      H3Resolution = 6  // ~3.23 km
	ResolutionBuilding   H3Resolution = 7  // ~1.22 km
	ResolutionStreet     H3Resolution = 8  // ~461.35 m (default)
	ResolutionIntersect  H3Resolution = 9  // ~174.38 m
	ResolutionProperty   H3Resolution = 10 // ~65.91 m
	ResolutionRoom       H3Resolution = 11 // ~24.91 m
	ResolutionDesk       H3Resolution = 12 // ~9.42 m
	ResolutionChair      H3Resolution = 13 // ~3.56 m
	ResolutionBook       H3Resolution = 14 // ~1.35 m
	ResolutionPage       H3Resolution = 15 // ~0.51 m
)

// Generator defines the interface for H3 index generation
type Generator interface {
	Generate(lat, lng float64, resolution H3Resolution) (string, error)
	ValidateCoordinates(lat, lng float64) error
	ValidateResolution(resolution H3Resolution) error
}

// BaseGenerator provides basic validation functionality for H3 generators
type BaseGenerator struct {
	validator validator.Validator
}

// ValidateCoordinates validates latitude and longitude using the validator module
func (g *BaseGenerator) ValidateCoordinates(lat, lng float64) error {
	return g.validator.ValidateCoordinates(lat, lng)
}

// ValidateResolution validates that the H3 resolution is within valid range (0-15)
func (g *BaseGenerator) ValidateResolution(resolution H3Resolution) error {
	if resolution < 0 || resolution > 15 {
		return fmt.Errorf("H3 resolution %d is out of valid range [0, 15]", resolution)
	}
	return nil
}

// H3Generator implements the Generator interface using Uber's H3 library
type H3Generator struct {
	BaseGenerator
}

// NewH3Generator creates a new H3 generator
func NewH3Generator() *H3Generator {
	return &H3Generator{
		BaseGenerator: BaseGenerator{
			validator: validator.NewCoordinateValidator(),
		},
	}
}

// Generate creates an H3 index for the given coordinates and resolution
func (g *H3Generator) Generate(lat, lng float64, resolution H3Resolution) (string, error) {
	// Validate coordinates first
	if err := g.ValidateCoordinates(lat, lng); err != nil {
		return "", fmt.Errorf("coordinate validation failed: %w", err)
	}

	// Validate resolution
	if err := g.ValidateResolution(resolution); err != nil {
		return "", fmt.Errorf("resolution validation failed: %w", err)
	}

	// Create H3 LatLng struct
	latLng := h3.NewLatLng(lat, lng)

	// Generate H3 index
	cell, err := h3.LatLngToCell(latLng, int(resolution))
	if err != nil {
		return "", fmt.Errorf("failed to generate H3 index: %w", err)
	}

	// Convert to string representation
	return cell.String(), nil
}