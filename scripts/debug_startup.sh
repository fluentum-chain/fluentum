#!/bin/bash

# Debug Startup Script
# This script helps debug the fluentumd startup issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "=========================================="
echo "    Debug Startup Script"
echo "=========================================="
echo ""

# Check if fluentumd binary exists
if [ ! -f "./build/fluentumd" ]; then
    print_error "fluentumd binary not found at ./build/fluentumd"
    exit 1
fi

print_success "Found fluentumd binary: ./build/fluentumd"

# Check home directory
HOME_DIR="/opt/fluentum"
if [ ! -d "$HOME_DIR" ]; then
    print_error "Home directory not found: $HOME_DIR"
    exit 1
fi

print_success "Found home directory: $HOME_DIR"

# Check config files
if [ ! -f "$HOME_DIR/config/config.toml" ]; then
    print_error "Config file not found: $HOME_DIR/config/config.toml"
    exit 1
fi

if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    print_error "Genesis file not found: $HOME_DIR/config/genesis.json"
    exit 1
fi

print_success "Found config files"

# Test fluentumd help
print_status "Testing fluentumd help..."
./build/fluentumd --help | head -20

# Test fluentumd start help
print_status "Testing fluentumd start help..."
./build/fluentumd start --help | head -20

# Test the exact command from systemd
print_status "Testing the exact systemd command..."
CMD="./build/fluentumd start --home $HOME_DIR --moniker fluentum-node1 --chain-id fluentum-testnet-1 --testnet --log_level info"

echo "Command: $CMD"
echo ""

# Test with verbose output
print_status "Running command with verbose output..."
$CMD 2>&1 || {
    print_error "Command failed with exit code $?"
    echo ""
    print_status "Let's try without --testnet flag..."
    CMD2="./build/fluentumd start --home $HOME_DIR --moniker fluentum-node1 --chain-id fluentum-testnet-1 --log_level info"
    echo "Command: $CMD2"
    $CMD2 2>&1 || {
        print_error "Command without --testnet also failed with exit code $?"
        echo ""
        print_status "Let's try with minimal arguments..."
        CMD3="./build/fluentumd start --home $HOME_DIR"
        echo "Command: $CMD3"
        $CMD3 2>&1 || {
            print_error "Minimal command also failed with exit code $?"
        }
    }
}

echo ""
print_status "Debug completed!" 