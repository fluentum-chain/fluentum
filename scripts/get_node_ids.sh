#!/bin/bash

# Script to get node IDs from all nodes and update persistent peers configuration
# Usage: ./get_node_ids.sh

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Node information from your provided data
declare -A NODE_IPS=(
  ["fluentum-node1"]="34.44.129.207"
  ["fluentum-node2"]="34.44.82.114"
  ["fluentum-node3"]="34.68.180.153"
  ["fluentum-node4"]="34.72.252.153"
  ["fluentum-node5"]="35.225.118.226"
)

VALID_NODES=("fluentum-node1" "fluentum-node2" "fluentum-node3" "fluentum-node4" "fluentum-node5")
RPC_PORT=26657

print_status "Getting node IDs from all nodes..."

# Array to store node ID and IP pairs
declare -a NODE_PAIRS=()

for node_name in "${VALID_NODES[@]}"; do
    local_ip="${NODE_IPS[$node_name]}"
    
    print_status "Checking $node_name ($local_ip)..."
    
    # Try to get node ID from RPC
    if curl -s --max-time 10 "http://$local_ip:$RPC_PORT/status" > /dev/null 2>&1; then
        node_id=$(curl -s --max-time 10 "http://$local_ip:$RPC_PORT/status" | jq -r '.result.node_info.id' 2>/dev/null || echo "")
        
        if [ -n "$node_id" ] && [ "$node_id" != "null" ]; then
            print_success "$node_name: $node_id"
            NODE_PAIRS+=("$node_id@$local_ip:26656")
        else
            print_warning "$node_name: Could not get node ID"
        fi
    else
        print_error "$node_name: Not reachable"
    fi
done

echo ""
print_status "Current node IDs and IPs:"
for pair in "${NODE_PAIRS[@]}"; do
    echo "  $pair"
done

echo ""
print_status "Persistent peers configuration:"
if [ ${#NODE_PAIRS[@]} -gt 0 ]; then
    persistent_peers=$(IFS=','; echo "${NODE_PAIRS[*]}")
    echo "persistent_peers = \"$persistent_peers\""
else
    print_error "No valid node IDs found"
fi

echo ""
print_status "To update the configuration files, run:"
echo "  # Update config/config.toml"
echo "  sed -i 's|persistent_peers = \".*\"|persistent_peers = \"$persistent_peers\"|' config/config.toml"
echo ""
echo "  # Update config/testnet-config.toml"
echo "  sed -i 's|persistent_peers = \".*\"|persistent_peers = \"$persistent_peers\"|' config/testnet-config.toml" 