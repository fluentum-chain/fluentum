#!/bin/bash

# Install required protobuf plugins for Go
echo "Installing protobuf plugins..."

# Install standard Go protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Also install gogofaster as backup (used in some parts of the codebase)
go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest

echo "Protobuf plugins installed successfully!"
echo "You can now run: make proto-gen" 