#!/bin/bash

# Deploy Fluentum Testnet Nodes to All Servers
# This script sets up and runs testnet nodes on all servers

[ "$DEBUG" = "1" ] && set -x
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

now() { date '+%Y-%m-%d %H:%M:%S'; }

print_status() {
    echo -e "$(now) ${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "$(now) ${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "$(now) ${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "$(now) ${RED}[ERROR]${NC} $1" >&2
}

CONFIG_TEMPLATE="config/testnet-config.toml"
source "$(dirname "$0")/nodes.conf"

FAILED_NODES=()

check_prerequisites() {
    print_status "Checking prerequisites..."
    if ! command -v fluentumd &> /dev/null; then
        print_error "fluentumd binary not found. Please install it first."
        exit 1
    fi
    if [ ! -f "$CONFIG_TEMPLATE" ]; then
        print_error "testnet-config.toml not found. Please run this script from the project root."
        exit 1
    fi
    print_success "Prerequisites check passed"
}

setup_node() {
    local node_name="$1"
    local ip_address="${NODE_IPS[$node_name]}"
    local fluentum_home="/tmp/$node_name"
    print_status "Setting up $node_name at $fluentum_home (IP: $ip_address)"
    mkdir -p "$fluentum_home/config"
    cp "$CONFIG_TEMPLATE" "$fluentum_home/config/config.toml"
    sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$fluentum_home/config/config.toml"
    sed -i "s/external_address = \"\"/external_address = \"$ip_address:$P2P_PORT\"/" "$fluentum_home/config/config.toml"
    sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$node_name\"/" "$fluentum_home/config/config.toml"
    if ! INIT_OUT=$(fluentumd init "$node_name" 2>&1); then
        print_error "Node $node_name initialization failed. Output:\n$INIT_OUT"
        FAILED_NODES+=("$node_name (init)")
        return 1
    fi
    print_success "Node $node_name setup complete"
}

create_service() {
    local node_name="$1"
    local fluentum_home="/tmp/$node_name"
    local service_name="fluentum-$node_name"
    local service_file="/etc/systemd/system/$service_name.service"
    print_status "Creating systemd service for $node_name..."
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

start_service() {
    local node_name="$1"
    local service_name="fluentum-$node_name"
    print_status "Starting service for $node_name..."
    if ! START_OUT=$(sudo systemctl restart "$service_name" 2>&1); then
        print_error "Failed to start service for $node_name. Output:\n$START_OUT"
        sudo systemctl status "$service_name"
        FAILED_NODES+=("$node_name (service)")
        return 1
    fi
    sleep 3
    if sudo systemctl is-active --quiet "$service_name"; then
        print_success "Service for $node_name started successfully"
    else
        print_error "Failed to start service for $node_name"
        sudo systemctl status "$service_name"
        FAILED_NODES+=("$node_name (service-active)")
        return 1
    fi
}

show_node_config() {
    local node_name="$1"
    local ip_address="${NODE_IPS[$node_name]}"
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

show_summary() {
    echo ""
    print_success "All nodes deployment completed!"
    if [ ${#FAILED_NODES[@]} -ne 0 ]; then
        print_error "Some nodes failed to deploy or start: ${FAILED_NODES[*]}"
    fi
    echo ""
    echo "Node Summary:"
    for node_name in "${VALID_NODES[@]}"; do
        local ip_address="${NODE_IPS[$node_name]}"
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
    echo "- Restart all: for node in ${VALID_NODES[@]}; do sudo systemctl restart fluentum-$node; done"
    echo ""
    echo "RPC Endpoints:"
    for node_name in "${VALID_NODES[@]}"; do
        local ip_address="${NODE_IPS[$node_name]}"
        echo "  $node_name: http://$ip_address:$RPC_PORT"
    done
    echo ""
    echo "P2P Endpoints:"
    for node_name in "${VALID_NODES[@]}"; do
        local ip_address="${NODE_IPS[$node_name]}"
        echo "  $node_name: $ip_address:$P2P_PORT"
    done
}

check_prerequisites
for node_name in "${VALID_NODES[@]}"; do
    setup_node "$node_name"
    create_service "$node_name"
    start_service "$node_name"
    show_node_config "$node_name"
done
show_summary
