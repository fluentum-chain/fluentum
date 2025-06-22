#!/bin/bash

# State Sync Feature Build Script
set -e

echo "Building State Sync Feature..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the feature directory
cd "$SCRIPT_DIR"

# Initialize go module if not already done
if [ ! -f "go.mod" ]; then
    echo "Initializing go module..."
    go mod init github.com/fluentum-chain/fluentum/features/state_sync
fi

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Run tests
echo "Running tests..."
go test -v ./...

# Build the feature (if there's a main package)
if [ -f "main.go" ]; then
    echo "Building feature binary..."
    go build -o state_sync .
fi

echo "State Sync Feature build completed successfully!" 