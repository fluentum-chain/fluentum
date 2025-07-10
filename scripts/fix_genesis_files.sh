#!/bin/bash

# Fix Genesis Files Script
# Copies the correct genesis.json from codebase to all testnet nodes

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

# Source node configuration
source "$(dirname "$0")/nodes.conf"

# Check if genesis file exists in codebase
GENESIS_SOURCE="$(dirname "$0")/../config/genesis.json"
if [ ! -f "$GENESIS_SOURCE" ]; then
    print_error "Genesis file not found at $GENESIS_SOURCE"
    exit 1
fi

print_status "Found genesis file at: $GENESIS_SOURCE"

# Function to fix genesis file on a node
fix_genesis_on_node() {
    local node_name=$1
    local node_ip=${NODE_IPS[$node_name]}
    
    print_status "Fixing genesis file on $node_name ($node_ip)..."
    
    # Copy genesis file to node
    if scp "$GENESIS_SOURCE" "ktang@$node_ip:/opt/fluentum/$node_name/config/genesis.json"; then
        print_success "Genesis file copied to $node_name"
        
        # Set proper permissions
        ssh "ktang@$node_ip" "chown ktang:ktang /opt/fluentum/$node_name/config/genesis.json && chmod 644 /opt/fluentum/$node_name/config/genesis.json"
        
        # Validate JSON on remote node
        if ssh "ktang@$node_ip" "python3 -m json.tool /opt/fluentum/$node_name/config/genesis.json > /dev/null 2>&1"; then
            print_success "Genesis file validation passed on $node_name"
        else
            print_error "Genesis file validation failed on $node_name"
        fi
    else
        print_error "Failed to copy genesis file to $node_name"
    fi
}

# Fix genesis files on all nodes
for node_name in "${VALID_NODES[@]}"; do
    fix_genesis_on_node "$node_name"
done

print_success "Genesis file fix completed for all nodes"
echo ""
echo "Next steps:"
echo "1. Restart the fluentum-testnet.service on each node"
echo "2. Run the health check script to verify all nodes are working"
echo ""
echo "Commands to restart services:"
for node_name in "${VALID_NODES[@]}"; do
    local node_ip=${NODE_IPS[$node_name]}
    echo "  ssh ktang@$node_ip 'sudo systemctl restart fluentum-testnet.service'"
done 