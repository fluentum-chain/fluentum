#!/bin/bash

# Quick Fluentum Testnet Node Deployment Script
# Usage: ./deploy_node.sh <node-name> <node-index>

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

# Check arguments
if [ $# -ne 2 ]; then
    print_error "Usage: $0 <node-name> <node-index>"
    echo "Example: $0 fluentum-node1 1"
    exit 1
fi

NODE_NAME=$1
NODE_INDEX=$2

# Validate node name
VALID_NODES=("fluentum-node1" "fluentum-node2" "fluentum-node3" "fluentum-node4" "fluentum-node5")
if [[ ! " ${VALID_NODES[@]} " =~ " ${NODE_NAME} " ]]; then
    print_error "Invalid node name: $NODE_NAME"
    echo "Valid options: ${VALID_NODES[*]}"
    exit 1
fi

# Validate node index
if ! [[ "$NODE_INDEX" =~ ^[1-5]$ ]]; then
    print_error "Invalid node index: $NODE_INDEX (must be 1-5)"
    exit 1
fi

print_status "Deploying Fluentum testnet node: $NODE_NAME (index: $NODE_INDEX)"

# Check if fluentumd exists
if [ ! -f "./build/fluentumd" ]; then
    print_error "fluentumd binary not found. Please build the project first:"
    echo "  make build"
    exit 1
fi

# Run the setup script
print_status "Running setup script..."
if [ -f "./scripts/setup_testnet.sh" ]; then
    chmod +x ./scripts/setup_testnet.sh
    ./scripts/setup_testnet.sh "$NODE_NAME" "$NODE_INDEX"
else
    print_error "Setup script not found: ./scripts/setup_testnet.sh"
    exit 1
fi

print_success "Node deployment completed!"
echo ""
echo "Next steps:"
echo "1. Start the node: sudo systemctl start fluentum-testnet.service"
echo "2. Check status: sudo systemctl status fluentum-testnet.service"
echo "3. View logs: sudo journalctl -u fluentum-testnet.service -f"
echo ""
echo "Remember to start nodes in order: node1 -> node2 -> node3 -> node4" 