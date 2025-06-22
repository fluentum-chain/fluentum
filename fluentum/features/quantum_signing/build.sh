#!/bin/sh

# Quantum Signing Feature Build Script
set -e

echo "Building Quantum Signing Feature..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the feature directory
cd "$SCRIPT_DIR"

# Initialize go module if not already done
if [ ! -f "go.mod" ]; then
    echo "Initializing go module..."
    go mod init github.com/fluentum-chain/fluentum/features/quantum_signing
fi

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build the feature (normal build)
echo "Building feature (normal build)..."
go build -tags "!plugin" -o quantum_signing feature.go

# Build the plugin (shared library)
echo "Building plugin (shared library)..."
go build -buildmode=plugin -tags "plugin" -o quantum_signing.so plugin.go

# Run tests
echo "Running tests..."
go test -v ./...

echo "Quantum Signing Feature build completed successfully!"
echo ""
echo "Generated files:"
echo "  - quantum_signing (feature binary)"
echo "  - quantum_signing.so (plugin shared library)"
echo ""
echo "Usage:"
echo "  Feature: Import and use in your application"
echo "  Plugin:  Load dynamically with plugin.LoadSignerPlugin('./quantum_signing.so')"

cd "$(dirname "$0")/lib"
make
cp quantum.so .. 