#!/bin/bash

echo "=== Force Rebuild Script ==="
echo "This script will force a complete rebuild of the fluentumd binary"

# Stop any running processes
echo "Stopping any running fluentumd processes..."
pkill -f fluentumd || true
sleep 2

# Clean everything
echo "Cleaning build artifacts..."
make clean || true
rm -rf build/
rm -rf vendor/
go clean -cache -modcache -testcache

# Remove any existing binary
echo "Removing existing binary..."
sudo rm -f /usr/local/bin/fluentumd || true
rm -f build/fluentumd || true

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy

# Build with verbose output
echo "Building with verbose output..."
make build V=1

# Check if build was successful
if [ -f "build/fluentumd" ]; then
    echo "Build successful!"
    echo "Binary size: $(ls -lh build/fluentumd)"
    echo "Binary timestamp: $(stat -c %y build/fluentumd 2>/dev/null || stat -f %Sm build/fluentumd 2>/dev/null || echo 'Cannot get timestamp')"
    
    # Install to system path
    echo "Installing binary to system path..."
    sudo cp build/fluentumd /usr/local/bin/ || cp build/fluentumd ~/.local/bin/ || echo "Could not install to system path"
    
    # Verify installation
    if command -v fluentumd >/dev/null 2>&1; then
        echo "Binary installed successfully at: $(which fluentumd)"
        echo "Version: $(fluentumd version 2>/dev/null || echo 'Version command failed')"
    else
        echo "Binary not found in PATH, using local build: ./build/fluentumd"
    fi
else
    echo "Build failed!"
    exit 1
fi

echo "=== Rebuild completed ===" 