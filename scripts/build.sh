#!/bin/bash

# Fluentum Core Build Script
# This script handles the build process with automatic dependency management

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.24.4 or later."
        exit 1
    fi
    
    if ! command_exists make; then
        print_error "Make is not installed. Please install make."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to handle dependencies
handle_dependencies() {
    print_status "Managing dependencies..."
    
    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        print_error "go.mod file not found. Are you in the correct directory?"
        exit 1
    fi
    
    # Run go mod tidy
    print_status "Running go mod tidy..."
    if go mod tidy; then
        print_success "Dependencies updated successfully"
    else
        print_error "Failed to update dependencies"
        exit 1
    fi
    
    # Verify dependencies
    print_status "Verifying dependencies..."
    if go mod verify; then
        print_success "Dependencies verified successfully"
    else
        print_error "Dependency verification failed"
        exit 1
    fi
}

# Function to build the project
build_project() {
    print_status "Building Fluentum Core..."
    
    # Clean previous builds
    print_status "Cleaning previous builds..."
    make clean
    
    # Build the project
    print_status "Building with automatic dependency management..."
    if make build; then
        print_success "Build completed successfully"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Function to run tests
run_tests() {
    print_status "Running tests..."
    
    if make test; then
        print_success "Tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

# Function to show build info
show_build_info() {
    print_status "Build Information:"
    echo "  Go version: $(go version)"
    echo "  Build target: $(pwd)/build/fluentumd"
    echo "  Build time: $(date)"
    
    if [ -f "build/fluentumd" ]; then
        echo "  Binary size: $(du -h build/fluentumd | cut -f1)"
        echo "  Binary created: $(ls -la build/fluentumd)"
    fi
}

# Main function
main() {
    echo "=========================================="
    echo "    Fluentum Core Build Script"
    echo "=========================================="
    echo ""
    
    # Parse command line arguments
    SKIP_DEPS=false
    SKIP_TESTS=false
    VERBOSE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --skip-deps     Skip dependency management"
                echo "  --skip-tests    Skip running tests"
                echo "  --verbose       Enable verbose output"
                echo "  --help          Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0                    # Full build with deps and tests"
                echo "  $0 --skip-deps        # Build without dependency management"
                echo "  $0 --skip-tests       # Build without running tests"
                echo "  $0 --skip-deps --skip-tests  # Quick build only"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_prerequisites
    
    # Handle dependencies (unless skipped)
    if [ "$SKIP_DEPS" = false ]; then
        handle_dependencies
    else
        print_warning "Skipping dependency management"
    fi
    
    # Build the project
    build_project
    
    # Run tests (unless skipped)
    if [ "$SKIP_TESTS" = false ]; then
        run_tests
    else
        print_warning "Skipping tests"
    fi
    
    # Show build information
    show_build_info
    
    print_success "Build process completed successfully!"
    echo ""
    echo "Next steps:"
    echo "  - Run 'make install' to install the binary"
    echo "  - Run 'make init-node' to initialize a new node"
    echo "  - Run 'make start' to start the node"
    echo "  - Run 'make help' for more commands"
}

# Run main function with all arguments
main "$@" 