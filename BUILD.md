# CSV H3 Tool - Build and Distribution Guide

This document provides comprehensive instructions for building and distributing the CSV H3 Tool across different platforms.

## Quick Start

### Windows
```cmd
# Local build (current platform only)
scripts\build-local.bat

# Cross-platform build (requires Linux/WSL)
scripts\build.bat
```

### Linux/macOS
```bash
# Make build script executable
chmod +x scripts/build.sh

# Local build
go build -o csv-h3-tool ./cmd

# Cross-platform build
./scripts/build.sh

# Using Makefile
make build
make release
```

## Build Requirements

### System Requirements
- Go 1.19 or later
- Git (for version information)
- Platform-specific tools for cross-compilation

### Dependencies
The project uses Go modules. All dependencies will be automatically downloaded:
- `github.com/uber/h3-go/v4` - H3 geospatial indexing
- `github.com/spf13/cobra` - CLI framework
- Standard Go libraries

### Platform Support

| Platform | Architecture | Status | Notes |
|----------|-------------|--------|-------|
| Linux | AMD64 | ✅ Supported | Primary development platform |
| Linux | ARM64 | ✅ Supported | Tested on ARM64 systems |
| Windows | AMD64 | ✅ Supported | Fully tested |
| Windows | ARM64 | ❌ Not Supported | H3 library limitation |
| macOS | AMD64 | ✅ Supported | Intel Macs |
| macOS | ARM64 | ✅ Supported | Apple Silicon Macs |
| FreeBSD | AMD64 | ✅ Supported | Basic support |

## Build Scripts

### 1. Local Build Scripts

#### Windows: `scripts/build-local.bat`
Builds for the current Windows platform only.

```cmd
scripts\build-local.bat [OPTIONS]

Options:
  -h, --help          Show help message
  -c, --clean         Clean build artifacts before building
  -t, --test          Run tests before building
```

#### Linux/macOS: Local Build
```bash
# Simple build
go build -o csv-h3-tool ./cmd

# With version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

go build -ldflags "${LDFLAGS}" -o csv-h3-tool ./cmd
```

### 2. Cross-Platform Build Scripts

#### Linux/macOS: `scripts/build.sh`
Comprehensive cross-platform build script with advanced features.

```bash
./scripts/build.sh [OPTIONS]

Options:
  -h, --help          Show help message
  -v, --version       Show version information
  -p, --platform      Build for specific platform (e.g., linux/amd64)
  -a, --all           Build for all platforms (default)
  -c, --clean         Clean build artifacts before building
  -t, --test          Run tests before building
  -r, --race          Enable race detection (development builds only)
  --no-package        Skip creating packages
  --no-checksums      Skip generating checksums

Examples:
  ./scripts/build.sh                    # Build for all platforms
  ./scripts/build.sh -p linux/amd64     # Build for Linux AMD64 only
  ./scripts/build.sh -c -t              # Clean, test, then build
  ./scripts/build.sh --no-package       # Build binaries only
```

#### Windows: `scripts/build.bat`
Cross-platform build for Windows (limited by H3 library constraints).

```cmd
scripts\build.bat [OPTIONS]

Options:
  -h, --help          Show help message
  -c, --clean         Clean build artifacts before building
  -t, --test          Run tests before building
  --no-package        Skip creating packages
```

### 3. Makefile Targets

```bash
# Build commands
make build              # Build for current platform
make build-dev          # Build with race detection
make install            # Install to GOPATH/bin
make clean              # Clean build artifacts

# Release commands
make release            # Complete release build
make release-build      # Build binaries for all platforms
make release-package    # Create packages (tar.gz, zip)
make release-checksums  # Generate SHA256 checksums
make release-clean      # Clean release artifacts

# Test commands
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make test-performance   # Run performance tests

# Development commands
make fmt                # Format code
make lint               # Run linter
make vet                # Run go vet
make check              # Run fmt, vet, lint, and short tests
```

## Build Configuration

### Version Information
The build system automatically embeds version information:

- **Version**: Git tag or commit hash
- **Build Time**: UTC timestamp
- **Git Commit**: Short commit hash

This information is displayed with `csv-h3-tool --version`.

### Build Flags
The following linker flags are used to embed version information:

```bash
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"
```

### Build Configuration File
`scripts/build-config.json` contains platform-specific build configuration:

```json
{
  "platforms": {
    "linux/amd64": {
      "output": "linux-amd64",
      "supported": true,
      "package_format": "tar.gz"
    },
    "windows/arm64": {
      "output": "windows-arm64.exe",
      "supported": false,
      "reason": "H3 library does not support Windows ARM64"
    }
  }
}
```

## Distribution

### Release Artifacts

A complete release build creates the following artifacts in `dist/`:

```
dist/
├── csv-h3-tool-linux-amd64
├── csv-h3-tool-linux-arm64
├── csv-h3-tool-windows-amd64.exe
├── csv-h3-tool-darwin-amd64
├── csv-h3-tool-darwin-arm64
├── csv-h3-tool-freebsd-amd64
└── packages/
    ├── csv-h3-tool-v1.0.0-linux-amd64.tar.gz
    ├── csv-h3-tool-v1.0.0-linux-arm64.tar.gz
    ├── csv-h3-tool-v1.0.0-windows-amd64.zip
    ├── csv-h3-tool-v1.0.0-darwin-amd64.tar.gz
    ├── csv-h3-tool-v1.0.0-darwin-arm64.tar.gz
    ├── csv-h3-tool-v1.0.0-freebsd-amd64.tar.gz
    └── checksums.txt
```

### Package Formats
- **Linux/macOS/FreeBSD**: `.tar.gz` archives
- **Windows**: `.zip` archives
- **Checksums**: SHA256 hashes in `checksums.txt`

### Installation

#### From Release Package
1. Download the appropriate package for your platform
2. Extract the archive
3. Move the binary to a directory in your PATH

```bash
# Linux/macOS example
tar -xzf csv-h3-tool-v1.0.0-linux-amd64.tar.gz
sudo mv csv-h3-tool-linux-amd64 /usr/local/bin/csv-h3-tool
```

#### From Source
```bash
# Install to GOPATH/bin
go install ./cmd

# Or use make
make install
```

## Development Builds

### Race Detection
Enable race detection for development builds:

```bash
# Linux/macOS
./scripts/build.sh -r

# Manual build
go build -race -o csv-h3-tool-dev ./cmd
```

### Debug Builds
For debugging, build without optimizations:

```bash
go build -gcflags="all=-N -l" -o csv-h3-tool-debug ./cmd
```

## Continuous Integration

### GitHub Actions Example
```yaml
name: Build and Release

on:
  push:
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.19'
    
    - name: Run tests
      run: make test
    
    - name: Build release
      run: make release
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: release-packages
        path: dist/packages/
```

## Troubleshooting

### Common Issues

#### H3 Library Build Constraints
**Problem**: Build fails with "build constraints exclude all Go files"
**Solution**: The H3 library doesn't support all platform combinations. Use the build scripts which handle unsupported platforms gracefully.

#### Cross-Compilation from Windows
**Problem**: Cross-compilation fails when building from Windows
**Solution**: Use WSL or a Linux environment for cross-platform builds. The Windows build scripts focus on Windows-only builds.

#### Missing Git Information
**Problem**: Version shows as "dev" or "unknown"
**Solution**: Ensure you're building from a Git repository with at least one commit.

#### Permission Errors
**Problem**: Build script permission denied
**Solution**: 
```bash
chmod +x scripts/build.sh
```

### Build Environment Setup

#### Linux/macOS
```bash
# Install Go
curl -L https://go.dev/dl/go1.19.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone <repository>
cd csv-h3-tool
make build
```

#### Windows
```cmd
# Install Go from https://golang.org/dl/
# Clone and build
git clone <repository>
cd csv-h3-tool
scripts\build-local.bat
```

## Performance Considerations

### Build Performance
- Use `go build -a` to force rebuild of all packages
- Use `GOCACHE=off` to disable build cache for clean builds
- Parallel builds are handled automatically by Go

### Binary Size Optimization
```bash
# Reduce binary size
go build -ldflags="-s -w" -o csv-h3-tool ./cmd

# With UPX compression (if available)
upx --best csv-h3-tool
```

## Security Considerations

### Checksums
Always verify checksums when distributing binaries:

```bash
# Generate checksums
sha256sum dist/packages/* > checksums.txt

# Verify checksums
sha256sum -c checksums.txt
```

### Code Signing
For production releases, consider code signing:

```bash
# Example with GPG
gpg --detach-sign --armor csv-h3-tool-linux-amd64
```

## Support

For build-related issues:
1. Check this documentation
2. Review the build scripts
3. Check the GitHub Issues
4. Ensure all dependencies are installed
5. Verify Go version compatibility

## Contributing

When modifying build scripts:
1. Test on multiple platforms
2. Update this documentation
3. Ensure backward compatibility
4. Add appropriate error handling
5. Update the build configuration file if needed