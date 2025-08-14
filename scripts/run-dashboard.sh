#!/bin/bash

# CSV H3 Tool - Test Dashboard Runner
# ==================================

set -e

echo "🚀 Starting Test Dashboard..."
echo

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

# Navigate to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Run the dashboard
go run scripts/test-dashboard.go

echo
echo "✨ Dashboard complete!"
echo
echo "💡 Quick commands:"
echo "   make test           # Run all tests"
echo "   make test-unit      # Run unit tests only"
echo "   make test-integration # Run integration tests"
echo "   make coverage       # Generate coverage report"