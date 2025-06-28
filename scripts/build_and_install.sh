#!/bin/bash

# Fluentum Build and Install Script
# This script builds and installs the Fluentum binary with the correct name

set -e

echo "🚀 Building and installing Fluentum Core..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

print_status "Go version: $(go version)"

# Set GOPATH if not set
if [ -z "$GOPATH" ]; then
    export GOPATH="$HOME/go"
    print_status "Setting GOPATH to $GOPATH"
fi

# Create build directory
print_status "Creating build directory..."
mkdir -p build

# Clean previous builds
print_status "Cleaning previous builds..."
rm -f build/fluentumd
rm -f build/fluentum

# Build the binary with the correct name
print_status "Building Fluentum Core as fluentumd..."
CGO_ENABLED=0 go build -mod=readonly -ldflags "-X github.com/fluentum-chain/fluentum/version.TMCoreSemVer=$(git describe --tags --always --dirty) -s -w" -tags 'tendermint,badgerdb' -o build/fluentumd ./cmd/fluentum/

if [ $? -eq 0 ]; then
    print_status "Build successful!"
else
    print_error "Build failed!"
    exit 1
fi

# Create GOPATH/bin directory if it doesn't exist
print_status "Creating GOPATH/bin directory..."
mkdir -p "$GOPATH/bin"

# Copy binary to GOPATH/bin
print_status "Installing fluentumd to $GOPATH/bin..."
cp build/fluentumd "$GOPATH/bin/"

# Make it executable
chmod +x "$GOPATH/bin/fluentumd"

# Verify installation
if [ -f "$GOPATH/bin/fluentumd" ]; then
    print_status "Installation successful!"
    print_status "Binary location: $GOPATH/bin/fluentumd"
    print_status "Binary size: $(du -h "$GOPATH/bin/fluentumd" | cut -f1)"
    
    # Test the binary
    print_status "Testing binary..."
    if "$GOPATH/bin/fluentumd" version > /dev/null 2>&1; then
        print_status "Binary test successful!"
        echo ""
        echo "🎉 Fluentum Core has been successfully installed as 'fluentumd'"
        echo ""
        echo "You can now use the following commands:"
        echo "  fluentumd version     - Check version"
        echo "  fluentumd init        - Initialize a new node"
        echo "  fluentumd start       - Start the node"
        echo ""
    else
        print_warning "Binary test failed, but installation completed"
        print_warning "You may need to check the binary manually"
    fi
else
    print_error "Installation failed! Binary not found at $GOPATH/bin/fluentumd"
    exit 1
fi 