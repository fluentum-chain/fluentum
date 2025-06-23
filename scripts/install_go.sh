#!/bin/bash

# Fluentum Go 1.20.14 Installation Script
# Required for Cosmos SDK v0.47.x compatibility

set -e

echo "🚀 Installing Go 1.20.14 for Fluentum Core..."
echo "📋 This version is required for optimal Cosmos SDK v0.47.x compatibility"

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if command -v apt-get &> /dev/null; then
        OS="ubuntu"
    elif command -v yum &> /dev/null; then
        OS="centos"
    else
        echo "❌ Unsupported Linux distribution"
        exit 1
    fi
else
    echo "❌ This script is for Linux systems only"
    exit 1
fi

# Check if Go is already installed
if command -v go &> /dev/null; then
    CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "📦 Current Go version: $CURRENT_VERSION"
    
    if [[ "$CURRENT_VERSION" == "go1.20.14" ]]; then
        echo "✅ Go 1.20.14 is already installed!"
        exit 0
    else
        echo "⚠️  Updating from Go $CURRENT_VERSION to 1.20.14..."
    fi
fi

# Download and install Go 1.20.14
echo "📥 Downloading Go 1.20.14..."
cd /tmp
wget -q https://go.dev/dl/go1.20.14.linux-amd64.tar.gz

if [[ ! -f "go1.20.14.linux-amd64.tar.gz" ]]; then
    echo "❌ Failed to download Go 1.20.14"
    exit 1
fi

echo "📦 Installing Go 1.20.14..."
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz

# Set up environment
echo "🔧 Setting up environment..."
if [[ "$OS" == "ubuntu" ]]; then
    # Ubuntu/Debian
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    if ! grep -q "/usr/local/go/bin" ~/.profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    fi
else
    # CentOS/RHEL
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
fi

# Clean up
rm -f go1.20.14.linux-amd64.tar.gz

# Verify installation
echo "✅ Verifying installation..."
export PATH=$PATH:/usr/local/go/bin
GO_VERSION=$(go version)

if [[ "$GO_VERSION" == *"go1.20.14"* ]]; then
    echo "🎉 Go 1.20.14 installed successfully!"
    echo "📋 Version: $GO_VERSION"
    echo ""
    echo "💡 Next steps:"
    echo "   1. Restart your terminal or run: source ~/.bashrc"
    echo "   2. Navigate to your Fluentum project"
    echo "   3. Run: go mod tidy"
    echo "   4. Run: make build"
else
    echo "❌ Installation verification failed"
    echo "Expected: go version go1.20.14 linux/amd64"
    echo "Got: $GO_VERSION"
    exit 1
fi

echo ""
echo "🔗 For more information, see:"
echo "   - VERSION_COMPATIBILITY.md"
echo "   - FINAL_CHECKLIST.md" 