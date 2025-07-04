#!/bin/bash

# Deploy Fluentum Testnet Nodes to All Servers
# This script sets up and runs testnet nodes on all 4 servers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Node configurations with specific IP addresses
declare -A NODE_CONFIGS=(
  ["fluentum-node1"]="34.44.129.207"
  ["fluentum-node3"]="34.44.82.114"
  ["fluentum-node4"]="34.68.180.153"
  ["fluentum-node5"]="34.72.252.153"
)

CHAIN_ID="fluentum-testnet-1"
CONFIG_TEMPLATE="config/testnet-config.toml"

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

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if fluentumd binary exists
    if ! command -v fluentumd &> /dev/null; then
        print_error "fluentumd binary not found. Please install it first."
        exit 1
    fi
    
    # Check if config template exists
    if [ ! -f "$CONFIG_TEMPLATE" ]; then
        print_error "testnet-config.toml not found. Please run this script from the project root."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to setup a single node
setup_node() {
    local node_name="$1"
    local ip_address="$2"
    local fluentum_home="/tmp/$node_name"
    
    print_status "Setting up $node_name at $fluentum_home (IP: $ip_address)"
    
    # Create node directory
    mkdir -p "$fluentum_home/config"
    
    # Copy config template
    cp "$CONFIG_TEMPLATE" "$fluentum_home/config/config.toml"
    
    # Update backend to pebble
    sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$fluentum_home/config/config.toml"
    
    # Update external_address with the specific IP
    sed -i "s/external_address = \"\"/external_address = \"$ip_address:26656\"/" "$fluentum_home/config/config.toml"
    
    # Update moniker
    sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$node_name\"/" "$fluentum_home/config/config.toml"
    
    # Initialize the node
    fluentumd init "$node_name" --chain-id $CHAIN_ID --home "$fluentum_home"
    
    print_success "Node $node_name setup complete"
}

# Function to create systemd service for a node
create_service() {
    local node_name="$1"
    local fluentum_home="/tmp/$node_name"
    local service_name="fluentum-$node_name"
    local service_file="/etc/systemd/system/$service_name.service"
    
    print_status "Creating systemd service for $node_name..."
    
    # Create service file
    sudo tee "$service_file" > /dev/null << EOF
[Unit]
Description=Fluentum Testnet Node - $node_name
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=/usr/local/bin/fluentumd start --home $fluentum_home
Restart=on-failure
RestartSec=3
LimitNOFILE=4096
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    print_success "Systemd service created for $node_name"
}

# Function to start service for a node
start_service() {
    local node_name="$1"
    local service_name="fluentum-$node_name"
    
    print_status "Starting service for $node_name..."
    
    sudo systemctl enable "$service_name"
    sudo systemctl start "$service_name"
    
    # Wait a moment and check status
    sleep 3
    
    if sudo systemctl is-active --quiet "$service_name"; then
        print_success "Service for $node_name started successfully"
    else
        print_error "Failed to start service for $node_name"
        sudo systemctl status "$service_name"
        return 1
    fi
}

# Function to show configuration for a node
show_node_config() {
    local node_name="$1"
    local ip_address="$2"
    local fluentum_home="/tmp/$node_name"
    
    echo ""
    print_status "Configuration for $node_name:"
    echo "  Node Name: $node_name"
    echo "  IP Address: $ip_address"
    echo "  Chain ID: $CHAIN_ID"
    echo "  Home Directory: $fluentum_home"
    echo "  Service Name: fluentum-$node_name"
    
    if [ -d "$fluentum_home" ]; then
        echo "  Config Files:"
        ls -la "$fluentum_home/config/"
        
        echo ""
        print_status "External address configuration:"
        grep "external_address" "$fluentum_home/config/config.toml"
    fi
}

# Function to show all nodes summary
show_summary() {
    echo ""
    print_success "All nodes deployment completed!"
    echo ""
    echo "Node Summary:"
    for node_name in "${!NODE_CONFIGS[@]}"; do
        local ip_address="${NODE_CONFIGS[$node_name]}"
        local service_name="fluentum-$node_name"
        echo "  $node_name (${ip_address}):"
        echo "    - Service: $service_name"
        echo "    - Home: /tmp/$node_name"
        echo "    - Status: sudo systemctl status $service_name"
        echo "    - Logs: sudo journalctl -u $service_name -f"
        echo ""
    done
    
    echo "Useful commands:"
    echo "- Check all services: sudo systemctl status fluentum-*"
    echo "- View all logs: sudo journalctl -u fluentum-* -f"
    echo "- Restart all: for node in fluentum-node1 fluentum-node3 fluentum-node4 fluentum-node5; do sudo systemctl restart fluentum-\$node; done"
    echo ""
    echo "RPC Endpoints:"
    for node_name in "${!NODE_CONFIGS[@]}"; do
        local ip_address="${NODE_CONFIGS[$node_name]}"
        echo "  $node_name: http://$ip_address:26657"
    done
    
    echo ""
    echo "P2P Endpoints:"
    for node_name in "${!NODE_CONFIGS[@]}"; do
        local ip_address="${NODE_CONFIGS[$node_name]}"
        echo "  $node_name: $ip_address:26656"
    done
}

# Function to cleanup existing nodes
cleanup_nodes() {
    print_status "Cleaning up existing nodes..."
    
    for node_name in "${!NODE_CONFIGS[@]}"; do
        local fluentum_home="/tmp/$node_name"
        local service_name="fluentum-$node_name"
        
        # Stop and disable service if exists
        if sudo systemctl list-unit-files | grep -q "$service_name"; then
            sudo systemctl stop "$service_name" 2>/dev/null || true
            sudo systemctl disable "$service_name" 2>/dev/null || true
        fi
        
        # Remove service file
        sudo rm -f "/etc/systemd/system/$service_name.service"
        
        # Remove node directory
        rm -rf "$fluentum_home"
    done
    
    # Reload systemd
    sudo systemctl daemon-reload
    
    print_success "Cleanup completed"
}

# Main function
main() {
    echo "=========================================="
    echo "    Deploy Fluentum Testnet Nodes"
    echo "=========================================="
    echo ""
    
    check_prerequisites
    
    echo "Available nodes:"
    for node_name in "${!NODE_CONFIGS[@]}"; do
        local ip_address="${NODE_CONFIGS[$node_name]}"
        echo "  $node_name: $ip_address"
    done
    echo ""
    
    echo "What would you like to do?"
    echo "1. Setup all nodes only"
    echo "2. Setup all nodes and create systemd services"
    echo "3. Setup all nodes, create services, and start"
    echo "4. Show all node configurations"
    echo "5. Cleanup all nodes"
    echo "6. Exit"
    echo ""
    
    read -p "Enter your choice (1-6): " choice
    
    case $choice in
        1)
            print_status "Setting up all nodes..."
            for node_name in "${!NODE_CONFIGS[@]}"; do
                local ip_address="${NODE_CONFIGS[$node_name]}"
                setup_node "$node_name" "$ip_address"
            done
            print_success "All nodes setup complete"
            ;;
        2)
            print_status "Setting up all nodes and creating services..."
            for node_name in "${!NODE_CONFIGS[@]}"; do
                local ip_address="${NODE_CONFIGS[$node_name]}"
                setup_node "$node_name" "$ip_address"
                create_service "$node_name"
            done
            sudo systemctl daemon-reload
            print_success "All nodes and services created"
            ;;
        3)
            print_status "Setting up all nodes, creating services, and starting..."
            for node_name in "${!NODE_CONFIGS[@]}"; do
                local ip_address="${NODE_CONFIGS[$node_name]}"
                setup_node "$node_name" "$ip_address"
                create_service "$node_name"
            done
            sudo systemctl daemon-reload
            
            print_status "Starting all services..."
            for node_name in "${!NODE_CONFIGS[@]}"; do
                start_service "$node_name"
            done
            
            show_summary
            ;;
        4)
            print_status "Showing all node configurations..."
            for node_name in "${!NODE_CONFIGS[@]}"; do
                local ip_address="${NODE_CONFIGS[$node_name]}"
                show_node_config "$node_name" "$ip_address"
            done
            ;;
        5)
            read -p "Are you sure you want to cleanup all nodes? (y/N): " confirm
            if [[ $confirm =~ ^[Yy]$ ]]; then
                cleanup_nodes
            else
                print_status "Cleanup cancelled"
            fi
            ;;
        6)
            print_status "Exiting..."
            exit 0
            ;;
        *)
            print_error "Invalid choice"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
