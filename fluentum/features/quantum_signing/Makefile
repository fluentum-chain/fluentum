# Quantum Signing Plugin Makefile

.PHONY: build clean test install uninstall benchmark

# Build configuration
PLUGIN_NAME = quantum_signer
PLUGIN_EXT = .so
BUILD_DIR = build
PLUGIN_PATH = $(BUILD_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)

# Go build flags
GO_FLAGS = -buildmode=plugin -ldflags="-s -w"
GO_TAGS = 

# Platform-specific settings
ifeq ($(OS),Windows_NT)
    PLUGIN_EXT = .dll
    GO_FLAGS += -H windowsgui
else ifeq ($(shell uname -s),Darwin)
    PLUGIN_EXT = .dylib
    GO_FLAGS += -H darwin
else
    PLUGIN_EXT = .so
    GO_FLAGS += -H linux
endif

# Update plugin path for platform
PLUGIN_PATH = $(BUILD_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)

# Default target
all: build

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build the plugin
build: $(BUILD_DIR)
	@echo "Building quantum signing plugin..."
	@echo "Platform: $(shell uname -s)"
	@echo "Plugin will be: $(PLUGIN_PATH)"
	go build $(GO_FLAGS) -o $(PLUGIN_PATH) ./quantum_signer.go
	@echo "Build complete: $(PLUGIN_PATH)"

# Build with debug information
build-debug: $(BUILD_DIR)
	@echo "Building quantum signing plugin with debug info..."
	go build -buildmode=plugin -gcflags="all=-N -l" -o $(PLUGIN_PATH) ./quantum_signer.go
	@echo "Debug build complete: $(PLUGIN_PATH)"

# Build optimized version
build-optimized: $(BUILD_DIR)
	@echo "Building optimized quantum signing plugin..."
	go build $(GO_FLAGS) -ldflags="-s -w -extldflags=-static" -o $(PLUGIN_PATH) ./quantum_signer.go
	@echo "Optimized build complete: $(PLUGIN_PATH)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install plugin to system directory
install: build
	@echo "Installing quantum signing plugin..."
	@if [ -f $(PLUGIN_PATH) ]; then \
		sudo cp $(PLUGIN_PATH) /usr/local/lib/fluentum/plugins/; \
		sudo chmod 755 /usr/local/lib/fluentum/plugins/$(PLUGIN_NAME)$(PLUGIN_EXT); \
		echo "Plugin installed to /usr/local/lib/fluentum/plugins/"; \
	else \
		echo "Error: Plugin not found at $(PLUGIN_PATH)"; \
		exit 1; \
	fi

# Uninstall plugin
uninstall:
	@echo "Uninstalling quantum signing plugin..."
	sudo rm -f /usr/local/lib/fluentum/plugins/$(PLUGIN_NAME)$(PLUGIN_EXT)
	@echo "Plugin uninstalled"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...
	@echo "Benchmarks complete"

# Validate plugin
validate: build
	@echo "Validating plugin..."
	@if [ -f $(PLUGIN_PATH) ]; then \
		echo "Plugin file exists: $(PLUGIN_PATH)"; \
		ls -la $(PLUGIN_PATH); \
		file $(PLUGIN_PATH); \
	else \
		echo "Error: Plugin not found"; \
		exit 1; \
	fi

# Show plugin information
info: build
	@echo "Plugin Information:"
	@echo "==================="
	@echo "Name: $(PLUGIN_NAME)"
	@echo "Path: $(PLUGIN_PATH)"
	@echo "Size: $(shell ls -lh $(PLUGIN_PATH) | awk '{print $$5}')"
	@echo "Type: $(shell file $(PLUGIN_PATH))"
	@echo "==================="

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go mod download
	@echo "Development setup complete"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting complete"

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...
	@echo "Linting complete"

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...
	@echo "Security scan complete"

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the quantum signing plugin"
	@echo "  build-debug    - Build with debug information"
	@echo "  build-optimized - Build optimized version"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  install        - Install plugin to system"
	@echo "  uninstall      - Uninstall plugin"
	@echo "  benchmark      - Run benchmarks"
	@echo "  validate       - Validate plugin"
	@echo "  info           - Show plugin information"
	@echo "  dev-setup      - Setup development environment"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  security       - Security scan"
	@echo "  help           - Show this help" 