#!/bin/bash

# Generate protobuf files using buf (recommended approach)
# This script handles the protobuf generation for the Fluentum project using buf

set -e

# Change to the project root directory
cd "$(dirname "$0")/.."

echo "Generating protobuf files using buf from $(pwd)..."

# Install Go protobuf plugins if not already installed
echo "Installing Go protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files using buf
echo "Running buf generate..."
cd proto
buf generate

echo "Protobuf generation completed successfully!" 