#!/bin/bash

# Quantum Signer Plugin Build Script
set -e

echo "Building Quantum Signer Plugin..."

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the plugin directory
cd "$SCRIPT_DIR"

# Initialize go module if not already done
if [ ! -f "go.mod" ]; then
    echo "Initializing go module..."
    go mod init github.com/fluentum-chain/fluentum/plugins/quantum_signer
fi

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build the plugin as a shared library
echo "Building plugin as shared library..."
go build -buildmode=plugin -o quantum_signer.so quantum_signer.go

echo "Quantum Signer Plugin built successfully!"
echo "Plugin file: quantum_signer.so"
echo ""
echo "Usage:"
echo "  import \"github.com/fluentum-chain/fluentum/fluentum/core/plugin\""
echo "  err := plugin.LoadSignerPlugin(\"./quantum_signer.so\")" 