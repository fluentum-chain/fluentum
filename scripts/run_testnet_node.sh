#!/bin/bash

# Run Fluentum Testnet Node Setup
# This script sets up and runs a testnet node on the server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NODE_NAME="fluentum-node1"
NODE_IP="34.44.129.207"
CHAIN_ID="fluentum-testnet-1"
FLUENTUM_HOME="/tmp/$NODE_NAME"

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
    if [ ! -f "config/testnet-config.toml" ]; then
        print_error "testnet-config.toml not found. Please run this script from the project root."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to setup node
setup_node() {
    print_status "Setting up $NODE_NAME at $FLUENTUM_HOME (IP: $NODE_IP)"
    
    # Create node directory
    mkdir -p "$FLUENTUM_HOME/config"
    
    # Copy config template
    cp "config/testnet-config.toml" "$FLUENTUM_HOME/config/config.toml"
    
    # Update backend to pebble
    sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"
    
    # Update external_address with the specific IP
    sed -i "s/external_address = \"\"/external_address = \"$NODE_IP:26656\"/" "$FLUENTUM_HOME/config/config.toml"
    
    # Initialize the node (without --testnet flag)
    fluentumd init "$NODE_NAME" --chain-id $CHAIN_ID --home "$FLUENTUM_HOME"
    
    print_success "Node setup complete"
}

# Function to show configuration
show_configuration() {
    print_status "Node configuration:"
    echo "  Node Name: $NODE_NAME"
    echo "  IP Address: $NODE_IP"
    echo "  Chain ID: $CHAIN_ID"
    echo "  Home Directory: $FLUENTUM_HOME"
    echo ""
    
    print_status "Configuration files:"
    ls -la "$FLUENTUM_HOME"
    ls -la "$FLUENTUM_HOME/config/"
    
    echo ""
    print_status "Backend configuration:"
    grep -A 2 -B 2 "backend" "$FLUENTUM_HOME/config/config.toml"
    
    echo ""
    print_status "External address configuration:"
    grep "external_address" "$FLUENTUM_HOME/config/config.toml"
}

# Function to create systemd service
create_service() {
    print_status "Creating systemd service..."
    
    SERVICE_FILE="/etc/systemd/system/fluentum-testnet.service"
    
    # Create service file
    sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Fluentum Testnet Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=/usr/local/bin/fluentumd start --home $FLUENTUM_HOME
Restart=on-failure
RestartSec=3
LimitNOFILE=4096
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd
    sudo systemctl daemon-reload
    
    print_success "Systemd service created"
}

# Function to start service
start_service() {
    print_status "Starting Fluentum testnet service..."
    
    sudo systemctl enable fluentum-testnet
    sudo systemctl start fluentum-testnet
    
    # Wait a moment and check status
    sleep 3
    
    if sudo systemctl is-active --quiet fluentum-testnet; then
        print_success "Fluentum testnet service started successfully"
    else
        print_error "Failed to start Fluentum testnet service"
        sudo systemctl status fluentum-testnet
        exit 1
    fi
}

# Function to show next steps
show_next_steps() {
    echo ""
    print_success "Testnet node deployment completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Check service status: sudo systemctl status fluentum-testnet"
    echo "2. View logs: sudo journalctl -u fluentum-testnet -f"
    echo "3. Check RPC endpoint: curl http://localhost:26657/status"
    echo "4. Configure firewall: sudo ufw allow 26656/tcp 26657/tcp 26660/tcp"
    echo ""
    echo "Configuration files:"
    echo "- Config: $FLUENTUM_HOME/config/config.toml"
    echo "- Genesis: $FLUENTUM_HOME/config/genesis.json"
    echo "- Keys: $FLUENTUM_HOME/config/priv_validator_key.json"
    echo ""
    echo "Useful commands:"
    echo "- Stop service: sudo systemctl stop fluentum-testnet"
    echo "- Restart service: sudo systemctl restart fluentum-testnet"
    echo "- View logs: sudo journalctl -u fluentum-testnet -f"
    echo "- Check sync: curl http://localhost:26657/status | jq '.result.sync_info.catching_up'"
    echo ""
    echo "Manual start command:"
    echo "fluentumd start --home $FLUENTUM_HOME"
}

# Main function
main() {
    echo "=========================================="
    echo "    Fluentum Testnet Node Setup"
    echo "=========================================="
    echo ""
    
    check_prerequisites
    
    echo "What would you like to do?"
    echo "1. Setup node only"
    echo "2. Setup node and create systemd service"
    echo "3. Setup node, create service, and start"
    echo "4. Show current configuration"
    echo "5. Exit"
    echo ""
    
    read -p "Enter your choice (1-5): " choice
    
    case $choice in
        1)
            setup_node
            show_configuration
            ;;
        2)
            setup_node
            create_service
            show_configuration
            print_success "Service created. Start with: sudo systemctl start fluentum-testnet"
            ;;
        3)
            setup_node
            create_service
            start_service
            show_next_steps
            ;;
        4)
            if [ -d "$FLUENTUM_HOME" ]; then
                show_configuration
            else
                print_warning "Node not set up yet. Run option 1 first."
            fi
            ;;
        5)
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