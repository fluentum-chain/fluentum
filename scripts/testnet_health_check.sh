#!/bin/bash

# Fluentum Testnet Health Check Script
# Monitors all 4 testnet nodes

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

# Testnet configuration
TESTNET_CHAIN_ID="fluentum-testnet-1"

# Server configurations
# Use standard ports for all nodes
# Format: ["node_name"]="<ip>:26657"
declare -A SERVERS=(
    ["fluentum-node1"]="34.44.129.207:26657"
    ["fluentum-node2"]="34.44.82.114:26657"
    ["fluentum-node3"]="34.68.180.153:26657"
    ["fluentum-node4"]="34.72.252.153:26657"
    ["fluentum-node5"]="35.225.118.226:26657"
)

# Function to check node health
check_node_health() {
    local node_name=$1
    local endpoint=$2
    
    echo "Checking $node_name ($endpoint)..."
    
    # Check if node is reachable
    if curl -s --max-time 10 "http://$endpoint/status" > /dev/null 2>&1; then
        print_success "$node_name is reachable"
        
        # Get node status
        local status=$(curl -s --max-time 10 "http://$endpoint/status")
        
        if [ $? -eq 0 ] && [ -n "$status" ]; then
            # Extract information using jq if available, otherwise use grep
            if command -v jq &> /dev/null; then
                local chain_id=$(echo "$status" | jq -r '.result.node_info.network' 2>/dev/null || echo "unknown")
                local latest_height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "unknown")
                local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up' 2>/dev/null || echo "unknown")
                
                echo "  Chain ID: $chain_id"
                echo "  Latest Block Height: $latest_height"
                echo "  Catching Up: $catching_up"
                
                # Check if chain ID matches
                if [ "$chain_id" = "$TESTNET_CHAIN_ID" ]; then
                    print_success "  Chain ID matches testnet"
                else
                    print_warning "  Chain ID mismatch: expected $TESTNET_CHAIN_ID, got $chain_id"
                fi
                
                # Check if node is caught up
                if [ "$catching_up" = "false" ]; then
                    print_success "  Node is caught up"
                else
                    print_warning "  Node is still catching up"
                fi
                
            else
                # Fallback without jq
                echo "  Status: Available (install jq for detailed info)"
            fi
        else
            print_error "  Failed to parse status response"
        fi
        
    else
        print_error "$node_name is not reachable"
    fi
    
    echo ""
}

# Function to check local node
check_local_node() {
    echo "Checking local node..."
    
    # Check if service is running
    if systemctl is-active --quiet fluentum-testnet.service; then
        print_success "Local fluentum-testnet service is running"
    else
        print_error "Local fluentum-testnet service is not running"
        return 1
    fi
    
    # Wait/retry for local RPC endpoint
    local rpc_ready=false
    for i in {1..5}; do
        if curl -s --max-time 2 "http://localhost:26657/status" > /dev/null 2>&1; then
            rpc_ready=true
            break
        fi
        sleep 1
    done

    if [ "$rpc_ready" = true ]; then
        print_success "Local RPC endpoint is responding"
    else
        print_error "Local RPC endpoint is not responding after 5 seconds"
    fi
    
    # Check logs for errors
    local recent_errors=$(journalctl -u fluentum-testnet.service --since "5 minutes ago" | grep -i error | wc -l)
    if [ "$recent_errors" -gt 0 ]; then
        print_warning "Found $recent_errors errors in recent logs"
    else
        print_success "No recent errors in logs"
    fi
    
    echo ""
}

# Function to check network connectivity
check_network_connectivity() {
    echo "Checking network connectivity..."
    
    for node_name in "${!SERVERS[@]}"; do
        local endpoint=${SERVERS[$node_name]}
        local ip=$(echo "$endpoint" | cut -d: -f1)
        local rpc_port=26657
        local p2p_port=26656
        
        # Test TCP connectivity for RPC
        if timeout 5 bash -c "</dev/tcp/$ip/$rpc_port" 2>/dev/null; then
            print_success "TCP connection to $node_name ($ip:$rpc_port) successful"
        else
            print_error "TCP connection to $node_name ($ip:$rpc_port) failed"
        fi
        # Test TCP connectivity for P2P
        if timeout 5 bash -c "</dev/tcp/$ip/$p2p_port" 2>/dev/null; then
            print_success "TCP connection to $node_name ($ip:$p2p_port) successful (P2P)"
        else
            print_error "TCP connection to $node_name ($ip:$p2p_port) failed (P2P)"
        fi
    done
    
    echo ""
}

# Function to check consensus
check_consensus() {
    echo "Checking consensus status..."
    
    local heights=()
    local chain_ids=()
    
    # Collect data from all nodes
    for node_name in "${!SERVERS[@]}"; do
        local endpoint=${SERVERS[$node_name]}
        local status=$(curl -s --max-time 10 "http://$endpoint/status" 2>/dev/null)
        
        if [ $? -eq 0 ] && [ -n "$status" ]; then
            if command -v jq &> /dev/null; then
                local height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")
                local chain_id=$(echo "$status" | jq -r '.result.node_info.network' 2>/dev/null || echo "unknown")
                
                heights+=("$height")
                chain_ids+=("$chain_id")
                
                echo "  $node_name: Height $height, Chain $chain_id"
            fi
        fi
    done
    
    # Check if all nodes have the same chain ID
    if [ ${#chain_ids[@]} -gt 0 ]; then
        local first_chain_id=${chain_ids[0]}
        local chain_id_match=true
        
        for chain_id in "${chain_ids[@]}"; do
            if [ "$chain_id" != "$first_chain_id" ]; then
                chain_id_match=false
                break
            fi
        done
        
        if [ "$chain_id_match" = true ]; then
            print_success "All nodes have the same chain ID: $first_chain_id"
        else
            print_error "Chain ID mismatch detected"
        fi
    fi
    
    # Check if block heights are close
    if [ ${#heights[@]} -gt 0 ]; then
        local max_height=0
        local min_height=999999
        
        for height in "${heights[@]}"; do
            if [ "$height" -gt "$max_height" ] 2>/dev/null; then
                max_height=$height
            fi
            if [ "$height" -lt "$min_height" ] 2>/dev/null; then
                min_height=$height
            fi
        done
        
        local height_diff=$((max_height - min_height))
        if [ "$height_diff" -le 5 ]; then
            print_success "Block heights are synchronized (max diff: $height_diff)"
        else
            print_warning "Block heights are not synchronized (max diff: $height_diff)"
        fi
    fi
    
    echo ""
}

# Main execution
main() {
    echo "=== Fluentum Testnet Health Check ==="
    echo "Chain ID: $TESTNET_CHAIN_ID"
    echo "Timestamp: $(date)"
    echo ""
    
    # Check local node first
    check_local_node
    
    # Check network connectivity
    check_network_connectivity
    
    # Check each remote node
    for node_name in "${!SERVERS[@]}"; do
        check_node_health "$node_name" "${SERVERS[$node_name]}"
    done
    
    # Check consensus
    check_consensus
    
    echo "=== Health Check Complete ==="
}

# Run main function
main "$@" 