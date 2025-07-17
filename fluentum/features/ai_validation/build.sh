# AI Validator build script temporarily disabled.

#!/bin/bash

# Build script for Fluentum AI-Validation Core with QMoE Consensus
# This script builds the QMoE validator as a shared library plugin

set -e

# Configuration
PLUGIN_NAME="qmoe_validator"
PLUGIN_DIR="fluentum/features/ai_validation"
BUILD_DIR="build"
OUTPUT_DIR="plugins"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.19 or later."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_info "Found Go version: $GO_VERSION"
}

# Check if required tools are installed
check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check for required Go modules
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Please run 'go mod init' first."
        exit 1
    fi
    
    # Check for CGO support
    if [ "$CGO_ENABLED" != "1" ]; then
        print_warning "CGO is not enabled. Enabling CGO for plugin build..."
        export CGO_ENABLED=1
    fi
    
    print_success "Dependencies check completed"
}

# Create build directories
create_directories() {
    print_info "Creating build directories..."
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$BUILD_DIR/$PLUGIN_DIR"
    
    print_success "Build directories created"
}

# Build the QMoE validator plugin
build_plugin() {
    print_info "Building QMoE validator plugin..."
    
    cd "$PLUGIN_DIR"
    
    # Build as shared library plugin
    print_info "Compiling QMoE validator with -buildmode=plugin..."
    
    go build -buildmode=plugin \
        -o "../../$OUTPUT_DIR/${PLUGIN_NAME}.so" \
        -ldflags="-s -w" \
        .
    
    if [ $? -eq 0 ]; then
        print_success "QMoE validator plugin built successfully"
    else
        print_error "Failed to build QMoE validator plugin"
        exit 1
    fi
    
    cd - > /dev/null
}

# Build for different platforms
build_multi_platform() {
    print_info "Building for multiple platforms..."
    
    PLATFORMS=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        
        print_info "Building for $GOOS/$GOARCH..."
        
        # Set environment variables for cross-compilation
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        export CGO_ENABLED=1
        
        # Determine file extension
        if [ "$GOOS" = "windows" ]; then
            EXT=".dll"
        else
            EXT=".so"
        fi
        
        # Build plugin
        cd "$PLUGIN_DIR"
        
        go build -buildmode=plugin \
            -o "../../$OUTPUT_DIR/${PLUGIN_NAME}_${GOOS}_${GOARCH}${EXT}" \
            -ldflags="-s -w" \
            .
        
        if [ $? -eq 0 ]; then
            print_success "Built for $GOOS/$GOARCH"
        else
            print_warning "Failed to build for $GOOS/$GOARCH"
        fi
        
        cd - > /dev/null
    done
    
    # Reset to host platform
    unset GOOS GOARCH
}

# Run tests
run_tests() {
    print_info "Running tests..."
    
    cd "$PLUGIN_DIR"
    
    go test -v ./...
    
    if [ $? -eq 0 ]; then
        print_success "Tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
    
    cd - > /dev/null
}

# Run benchmarks
run_benchmarks() {
    print_info "Running benchmarks..."
    
    cd "$PLUGIN_DIR"
    
    go test -bench=. -benchmem ./...
    
    cd - > /dev/null
}

# Generate documentation
generate_docs() {
    print_info "Generating documentation..."
    
    # Create documentation directory
    mkdir -p "docs/ai_validation"
    
    # Generate Go documentation
    cd "$PLUGIN_DIR"
    
    godoc -http=:6060 &
    DOC_PID=$!
    
    # Wait for godoc to start
    sleep 3
    
    # Generate HTML documentation
    curl -s "http://localhost:6060/pkg/github.com/fluentum-chain/fluentum/features/ai_validation/" > "../../docs/ai_validation/index.html"
    
    # Kill godoc server
    kill $DOC_PID
    
    cd - > /dev/null
    
    print_success "Documentation generated"
}

# Create plugin manifest
create_manifest() {
    print_info "Creating plugin manifest..."
    
    cat > "$OUTPUT_DIR/plugin_manifest.json" << EOF
{
    "name": "QMoE Validator",
    "version": "1.0.0",
    "description": "Quantized Mixture-of-Experts consensus validator for Fluentum",
    "type": "ai_validator",
    "entry_point": "AIValidatorPlugin",
    "platforms": {
        "linux/amd64": "${PLUGIN_NAME}_linux_amd64.so",
        "linux/arm64": "${PLUGIN_NAME}_linux_arm64.so",
        "darwin/amd64": "${PLUGIN_NAME}_darwin_amd64.so",
        "darwin/arm64": "${PLUGIN_NAME}_darwin_arm64.so",
        "windows/amd64": "${PLUGIN_NAME}_windows_amd64.dll"
    },
    "dependencies": {
        "go_version": ">=1.19",
        "cgo": true
    },
    "features": [
        "quantized_moe_consensus",
        "predictive_batching",
        "dynamic_quantization",
        "gas_optimization",
        "adaptive_thresholds"
    ],
    "performance": {
        "target_gas_savings": "40%",
        "inference_time": "<10ms",
        "memory_usage": "<100MB"
    }
}
EOF
    
    print_success "Plugin manifest created"
}

# Install plugin
install_plugin() {
    print_info "Installing plugin..."
    
    # Create installation directory
    INSTALL_DIR="/usr/local/lib/fluentum/plugins"
    sudo mkdir -p "$INSTALL_DIR"
    
    # Copy plugin files
    sudo cp "$OUTPUT_DIR"/* "$INSTALL_DIR/"
    
    # Set permissions
    sudo chmod 755 "$INSTALL_DIR"/*
    
    print_success "Plugin installed to $INSTALL_DIR"
}

# Clean build artifacts
clean() {
    print_info "Cleaning build artifacts..."
    
    rm -rf "$BUILD_DIR"
    rm -rf "$OUTPUT_DIR"
    
    print_success "Build artifacts cleaned"
}

# Show help
show_help() {
    echo "Fluentum AI-Validation Core Build Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  build           Build the QMoE validator plugin (default)"
    echo "  multi           Build for multiple platforms"
    echo "  test            Run tests"
    echo "  benchmark       Run benchmarks"
    echo "  docs            Generate documentation"
    echo "  install         Install plugin to system"
    echo "  clean           Clean build artifacts"
    echo "  help            Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  CGO_ENABLED     Enable CGO (default: 1)"
    echo "  GOOS            Target operating system"
    echo "  GOARCH          Target architecture"
    echo ""
}

# Main function
main() {
    local action=${1:-build}
    
    print_info "Starting Fluentum AI-Validation Core build..."
    print_info "Action: $action"
    
    case $action in
        build)
            check_go
            check_dependencies
            create_directories
            build_plugin
            create_manifest
            ;;
        multi)
            check_go
            check_dependencies
            create_directories
            build_multi_platform
            create_manifest
            ;;
        test)
            run_tests
            ;;
        benchmark)
            run_benchmarks
            ;;
        docs)
            generate_docs
            ;;
        install)
            install_plugin
            ;;
        clean)
            clean
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown action: $action"
            show_help
            exit 1
            ;;
    esac
    
    print_success "Build process completed successfully!"
}

# Run main function with all arguments
main "$@" 