# CSV H3 Tool

A command-line tool written in Go that processes CSV files containing latitude and longitude coordinates, adding H3 index values as a new column using Uber's H3 geospatial indexing system.

## Features

- Process CSV files with latitude/longitude coordinates
- Generate H3 indexes using Uber's H3 library
- Configurable H3 resolution levels (0-15)
- Flexible column detection (by name or index)
- Streaming processing for large files
- Comprehensive error handling

## Installation

```bash
go build -o csv-h3-tool cmd/main.go
```

## Usage

```bash
./csv-h3-tool --input input.csv --output output.csv --lat-column latitude --lng-column longitude
```

## Options

- `--input, -i`: Input CSV file path (required)
- `--output, -o`: Output CSV file path (optional, defaults to input_h3.csv)
- `--lat-column`: Name or index of latitude column (default: "latitude")
- `--lng-column`: Name or index of longitude column (default: "longitude")
- `--resolution`: H3 resolution level 0-15 (default: 8)
- `--headers`: CSV has header row (default: true)
- `--overwrite`: Overwrite existing output file (default: false)
- `--verbose, -v`: Enable verbose logging

## H3 Resolution Levels

| Resolution | Avg Edge Length | Use Case |
|------------|----------------|----------|
| 0 | ~1107 km | Country level |
| 1 | ~418 km | State/Province level |
| 2 | ~158 km | Metropolitan area |
| 3 | ~59 km | City level |
| 4 | ~22 km | District level |
| 5 | ~8.5 km | Neighborhood |
| 6 | ~3.2 km | Large block |
| 7 | ~1.2 km | City block |
| 8 | ~461 m | Street level (default) |
| 9 | ~174 m | Intersection |
| 10 | ~65 m | Property level |
| 11 | ~24 m | Room level |
| 12 | ~9.4 m | Desk level |
| 13 | ~3.5 m | Chair level |
| 14 | ~1.3 m | Book level |
| 15 | ~0.5 m | Page level |

## Examples

```bash
# Basic usage
./csv-h3-tool -i locations.csv -o locations_with_h3.csv

# Custom column names and resolution
./csv-h3-tool -i data.csv --lat-column lat --lng-column lon --resolution 10

# High resolution for precise mapping
./csv-h3-tool -i data.csv --resolution 12
```

## Requirements

- Go 1.21 or higher

## License

MIT License