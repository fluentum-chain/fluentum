#!/bin/bash

# Fluentum Testnet Health Check Script
# Monitors all testnet nodes defined in nodes.conf
# 
# Configuration:
# - Node list and IPs: sourced from scripts/nodes.conf
# - Local node: automatically detected from hostname or running services
#   (can override with LOCAL_NODE_NAME env var)
# - Chain ID: sourced from nodes.conf
# - Ports: sourced from nodes.conf (RPC_PORT, P2P_PORT)
#
# Usage: ./testnet_health_check.sh
#        LOCAL_NODE_NAME=fluentum-node2 ./testnet_health_check.sh
#
# Auto-detection priority:
# 1. LOCAL_NODE_NAME environment variable (if set)
# 2. Hostname containing "fluentum-node" pattern (extracts node number)
# 3. Running fluentum-testnet.service detection (extracts from service description)
# 4. Fallback to fluentum-node1
#
# Service naming: All nodes use "fluentum-testnet.service" regardless of node number

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

# Source centralized node configuration
NODES_CONF="$(dirname "$0")/nodes.conf"
if [ ! -f "$NODES_CONF" ]; then
    echo "Error: nodes.conf file not found at $NODES_CONF"
    exit 1
fi
source "$NODES_CONF"

# Testnet configuration
TESTNET_CHAIN_ID="$CHAIN_ID"

# Server configurations
# Use standard ports for all nodes
# Format: ["node_name"]="<ip>:26657"
declare -A SERVERS=()
for node_name in "${VALID_NODES[@]}"; do
    SERVERS["$node_name"]="${NODE_IPS[$node_name]}:$RPC_PORT"
done

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

# Function to detect local node name
detect_local_node() {
    # If LOCAL_NODE_NAME is set, use it
    if [ -n "$LOCAL_NODE_NAME" ]; then
        # Validate the provided node name
        if [[ " ${VALID_NODES[@]} " =~ " $LOCAL_NODE_NAME " ]]; then
            echo "$LOCAL_NODE_NAME"
            return 0
        else
            print_warning "LOCAL_NODE_NAME ($LOCAL_NODE_NAME) is not in VALID_NODES"
        fi
    fi
    
    # Try to detect from hostname first
    local hostname=$(hostname -s)  # Using -s to get just the hostname without domain
    if [[ "$hostname" == *"fluentum-node"* ]]; then
        # Try exact match first
        if [[ " ${VALID_NODES[@]} " =~ " $hostname " ]]; then
            echo "$hostname"
            return 0
        fi
        
        # Try to extract node name from hostname (e.g., fluentum-node1)
        local node_name=$(echo "$hostname" | grep -o "fluentum-node[0-9]*" | head -1)
        if [ -n "$node_name" ] && [[ " ${VALID_NODES[@]} " =~ " $node_name " ]]; then
            echo "$node_name"
            return 0
        fi
    fi
    
    # Try to detect by checking running services
    # First check for node-specific services (fluentum-node1, fluentum-node2, etc.)
    for node_name in "${VALID_NODES[@]}"; do
        local service_name="fluentum-$node_name"
        if systemctl is-active --quiet "$service_name.service" 2>/dev/null; then
            echo "$node_name"
            return 0
        fi
    done
    
    # Then check for the generic service name
    local generic_services=("fluentum-testnet.service" "fluentumd.service")
    for service in "${generic_services[@]}"; do
        if systemctl is-active --quiet "$service" 2>/dev/null; then
            # Try to determine node name from service properties
            local working_dir=$(systemctl show "$service" --property=WorkingDirectory --value 2>/dev/null || echo "")
            local exec_start=$(systemctl show "$service" --property=ExecStart --value 2>/dev/null || echo "")
            
            # Check working directory for node name
            if [[ "$working_dir" == *"fluentum-node"* ]]; then
                local node_name=$(echo "$working_dir" | grep -o "fluentum-node[0-9]*" | head -1)
                if [ -n "$node_name" ] && [[ " ${VALID_NODES[@]} " =~ " $node_name " ]]; then
                    echo "$node_name"
                    return 0
                fi
            fi
            
            # Check ExecStart for --home parameter
            if [[ "$exec_start" == *"--home"* ]]; then
                local home_path=$(echo "$exec_start" | grep -o -- '--home[ =][^ ]*' | cut -d' ' -f2 | cut -d'=' -f2)
                if [[ "$home_path" == *"fluentum-node"* ]]; then
                    local node_name=$(basename "$home_path")
                    if [ -n "$node_name" ] && [[ " ${VALID_NODES[@]} " =~ " $node_name " ]]; then
                        echo "$node_name"
                        return 0
                    fi
                fi
            fi
        fi
    done
    
    # Try to determine from config directory
    local config_dirs=("/opt/fluentum" "/root/.fluentum" "/home/*/.fluentum")
    for dir in "${config_dirs[@]}"; do
        if [ -d "$dir" ]; then
            for node_name in "${VALID_NODES[@]}"; do
                if [ -d "$dir/$node_name" ] || [ -d "$dir/config" ]; then
                    # If we find a node-specific directory, use that
                    if [ -d "$dir/$node_name" ] && [[ " ${VALID_NODES[@]} " =~ " $node_name " ]]; then
                        echo "$node_name"
                        return 0
                    # If we find a config directory, try to get node name from config
                    elif [ -d "$dir/config" ]; then
                        local config_toml="$dir/config/config.toml"
                        if [ -f "$config_toml" ]; then
                            local moniker=$(grep -oP '^moniker\s*=\s*"\K[^"]+' "$config_toml" 2>/dev/null || echo "")
                            if [ -n "$moniker" ] && [[ " ${VALID_NODES[@]} " =~ " $moniker " ]]; then
                                echo "$moniker"
                                return 0
                            fi
                        fi
                    fi
                fi
            done
        fi
    done
    
    # If we still can't determine, prompt the user
    print_warning "Could not automatically determine node name. Please set LOCAL_NODE_NAME environment variable."
    echo "Available nodes: ${VALID_NODES[*]}"
    
    # Fallback to first node if we can't determine
    echo "${VALID_NODES[0]}"
    return 1
}

# Function to check local node
check_local_node() {
    echo "Checking local node..."
    
    # Detect local node name
    local local_node_name
    local_node_name=$(detect_local_node)
    local detection_status=$?
    
    if [ $detection_status -ne 0 ]; then
        print_warning "Node detection had issues, but continuing with $local_node_name"
    fi
    
    print_status "Detected local node: $local_node_name"
    
    # Try both service name patterns
    local service_names=("fluentum-$local_node_name" "fluentum-testnet" "fluentumd")
    local service_found=false
    local active_service=""
    
    for service in "${service_names[@]}"; do
        if systemctl is-active --quiet "$service" 2>/dev/null || systemctl is-active --quiet "$service.service" 2>/dev/null; then
            service_found=true
            active_service="$service"
            # Remove .service suffix if present for consistent display
            active_service="${active_service%.service}"
            print_success "Local $active_service service is running"
            break
        fi
    done
    
    if [ "$service_found" = false ]; then
        print_error "No Fluentum service is running. Tried: ${service_names[*]}"
        return 1
    fi
    
    # Get the actual service name with .service suffix for journalctl
    local full_service_name="$active_service"
    if ! [[ "$full_service_name" == *".service" ]]; then
        full_service_name="${full_service_name}.service"
    fi
    
    # Check if RPC port is open
    local rpc_ready=false
    local rpc_url="http://localhost:$RPC_PORT/status"
    
    print_status "Checking RPC endpoint at $rpc_url"
    
    # Try with retries
    for i in {1..5}; do
        if curl -s --max-time 2 "$rpc_url" > /dev/null 2>&1; then
            rpc_ready=true
            break
        fi
        print_status "Waiting for RPC endpoint... (attempt $i/5)"
        sleep 2
    done
    
    if [ "$rpc_ready" = true ]; then
        print_success "RPC endpoint is responding"
        
        # Get node status
        local status=$(curl -s --max-time 5 "$rpc_url" 2>/dev/null)
        if [ -n "$status" ]; then
            if command -v jq &> /dev/null; then
                local latest_block=$(echo "$status" | jq -r '.result.sync_info.latest_block_height // empty' 2>/dev/null)
                local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up // empty' 2>/dev/null)
                local node_id=$(echo "$status" | jq -r '.result.node_info.id // empty' 2>/dev/null)
                
                echo "  Latest Block: ${latest_block:-unknown}"
                echo "  Catching Up: ${catching_up:-unknown}"
                echo "  Node ID: ${node_id:0:10}...${node_id: -10}"
                
                if [ "$catching_up" = "false" ]; then
                    print_success "Node is caught up"
                elif [ "$catching_up" = "true" ]; then
                    print_warning "Node is still catching up"
                fi
            else
                print_warning "jq not found. Install jq for detailed status information."
            fi
        fi
    else
        print_error "RPC endpoint is not responding after 10 seconds"
    fi
    
    # Check logs for errors (last 10 minutes)
    print_status "Checking recent logs for errors..."
    local log_check_cmd="journalctl -u \"$full_service_name\" --since \"10 minutes ago\" | grep -i -E 'error|failed|exception|panic' | tail -n 10"
    local recent_errors=$(eval "$log_check_cmd" | wc -l)
    
    if [ "$recent_errors" -gt 0 ]; then
        print_warning "Found $recent_errors errors/warnings in recent logs"
        echo "Last few errors from logs (if any):"
        eval "$log_check_cmd" 2>/dev/null || echo "  (unable to retrieve logs)"
    else
        print_success "No recent errors found in logs"
    fi
    
    # Check disk space
    local disk_check=$(df -h / | awk 'NR==2 {print $5 " used (" $4 " free)"}')
    print_status "Disk Usage: $disk_check"
    
    # Check memory usage
    local mem_check=$(free -h | awk '/^Mem:/ {print $3 "/" $2 " used (" $4 " free)"}')
    print_status "Memory Usage: $mem_check"
    
    echo ""
}

# Function to check network connectivity
check_network_connectivity() {
    echo "Checking network connectivity..."
    
    for node_name in "${!SERVERS[@]}"; do
        local endpoint=${SERVERS[$node_name]}
        local ip=$(echo "$endpoint" | cut -d: -f1)
        local rpc_port=$RPC_PORT
        local p2p_port=$P2P_PORT
        
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

# Function to display configuration
display_config() {
    echo "=== Fluentum Testnet Health Check ==="
    echo "Configuration:"
    echo "  Chain ID: $TESTNET_CHAIN_ID"
    echo "  RPC Port: $RPC_PORT"
    echo "  P2P Port: $P2P_PORT"
    echo "  Local Node: $(detect_local_node)"
    echo "  Total Nodes: ${#SERVERS[@]}"
    echo "  Timestamp: $(date)"
    echo ""
}

# Main execution
main() {
    display_config
    
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