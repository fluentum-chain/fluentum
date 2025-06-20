#!/bin/bash

# Fluentum Ubuntu Installation Script
# This script installs Fluentum Core on Ubuntu systems

set -e

echo "🚀 Installing Fluentum Core on Ubuntu..."

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

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root"
   exit 1
fi

# Update system packages
print_status "Updating system packages..."
sudo apt update

# Install required dependencies
print_status "Installing required dependencies..."
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    pkg-config \
    libssl-dev \
    libgmp-dev \
    libtool \
    autoconf \
    automake \
    cmake \
    clang \
    clang-format

# Install Go if not already installed
if ! command -v go &> /dev/null; then
    print_status "Installing Go..."
    GO_VERSION="1.24.4"
    GO_ARCH="linux-amd64"
    
    # Download and install Go
    wget -q "https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz" -O /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    
    print_status "Go installed successfully"
else
    print_status "Go is already installed"
fi

# Verify Go installation
if ! command -v go &> /dev/null; then
    print_error "Go installation failed. Please install Go manually."
    exit 1
fi

print_status "Go version: $(go version)"

# Create build directory
print_status "Setting up build environment..."
mkdir -p build

# Build Fluentum
print_status "Building Fluentum Core..."
make build

# Install Fluentum
print_status "Installing Fluentum Core..."
make install

# Verify installation
if command -v fluentum &> /dev/null; then
    print_status "Fluentum Core installed successfully!"
    echo ""
    echo "🎉 Installation completed successfully!"
    echo ""
    echo "Available commands:"
    echo "  fluentum version    - Check version"
    echo "  fluentum init       - Initialize a new node"
    echo "  fluentum node       - Start the node"
    echo "  fluentum --help     - Show all commands"
    echo ""
    echo "Version: $(fluentum version)"
else
    print_error "Installation failed. Please check the build output above."
    exit 1
fi

print_status "Installation completed successfully!" 