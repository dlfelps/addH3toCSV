@echo off
REM CSV H3 Tool - Local Build Script for Windows
REM ============================================

setlocal enabledelayedexpansion

REM Configuration
set BINARY_NAME=csv-h3-tool
set CMD_PATH=./cmd

REM Get version information
for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev

for /f "delims=" %%i in ('powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'"') do set BUILD_TIME=%%i

for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

REM Build flags
set LDFLAGS=-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%

echo ================================
echo CSV H3 Tool - Local Build
echo ================================
echo.
echo Version: %VERSION%
echo Build Time: %BUILD_TIME%
echo Git Commit: %GIT_COMMIT%
echo.

REM Parse command line arguments
set CLEAN_FIRST=false
set RUN_TESTS=false

:parse_args
if "%1"=="" goto start_build
if "%1"=="-h" goto show_help
if "%1"=="--help" goto show_help
if "%1"=="-c" set CLEAN_FIRST=true
if "%1"=="--clean" set CLEAN_FIRST=true
if "%1"=="-t" set RUN_TESTS=true
if "%1"=="--test" set RUN_TESTS=true
shift
goto parse_args

:show_help
echo Usage: %0 [OPTIONS]
echo.
echo Options:
echo   -h, --help          Show this help message
echo   -c, --clean         Clean build artifacts before building
echo   -t, --test          Run tests before building
echo.
echo This script builds for the current platform only.
echo For cross-platform builds, use Linux/macOS with the build.sh script.
goto :eof

:start_build

REM Clean if requested
if "%CLEAN_FIRST%"=="true" (
    echo Cleaning build artifacts...
    if exist "%BINARY_NAME%.exe" del "%BINARY_NAME%.exe"
    if exist "%BINARY_NAME%-dev.exe" del "%BINARY_NAME%-dev.exe"
    echo ✓ Cleaned build artifacts
    echo.
)

REM Run tests if requested
if "%RUN_TESTS%"=="true" (
    echo Running tests...
    go test -v ./...
    if errorlevel 1 (
        echo ✗ Tests failed
        exit /b 1
    )
    echo ✓ All tests passed
    echo.
)

REM Build for current platform
echo Building for current platform (Windows)...
go build -ldflags "%LDFLAGS%" -o "%BINARY_NAME%.exe" "%CMD_PATH%"
if errorlevel 1 (
    echo ✗ Build failed
    exit /b 1
)
echo ✓ Built %BINARY_NAME%.exe

echo.
echo Build completed successfully!
echo Binary: %BINARY_NAME%.exe

REM Test the binary
echo.
echo Testing binary...
%BINARY_NAME%.exe --version
if errorlevel 1 (
    echo ✗ Binary test failed
    exit /b 1
)
echo ✓ Binary test passed

endlocal