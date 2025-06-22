#!/bin/bash

# ZK Rollup Feature Build Script
set -e

echo "Building ZK Rollup Feature..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the feature directory
cd "$SCRIPT_DIR"

# Initialize go module if not already done
if [ ! -f "go.mod" ]; then
    echo "Initializing go module..."
    go mod init github.com/fluentum-chain/fluentum/features/zk_rollup
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
    go build -o zk_rollup .
fi

echo "ZK Rollup Feature build completed successfully!" 