#!/bin/bash

# Continue Fluentum deployment from binary installation
# This script continues where deploy_server.sh left off

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_CHAIN_ID="fluentum-1"
DEFAULT_MONIKER="fluentum-node"
CONFIG_DIR="$HOME/.fluentumd"
SERVICE_FILE="/etc/systemd/system/fluentumd.service"

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

# Function to fix config file line endings
fix_config_line_endings() {
    print_status "Fixing config file line endings..."
    
    if [ -f "config/config.toml" ]; then
        # Convert Windows line endings to Unix line endings
        dos2unix config/config.toml 2>/dev/null || sed -i 's/\r$//' config/config.toml
        print_success "Config file line endings fixed"
    else
        print_warning "Config file not found, will be created during initialization"
    fi
}

# Function to initialize node
initialize_node() {
    print_status "Initializing Fluentum node..."
    
    # Get moniker from user or use default
    read -p "Enter node moniker (default: $DEFAULT_MONIKER): " MONIKER
    MONIKER=${MONIKER:-$DEFAULT_MONIKER}
    
    # Get chain ID from user or use default
    read -p "Enter chain ID (default: $DEFAULT_CHAIN_ID): " CHAIN_ID
    CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
    
    # Initialize the node
    fluentumd init "$MONIKER" --chain-id "$CHAIN_ID"
    
    print_success "Node initialized successfully"
}

# Function to configure node for server
configure_node() {
    print_status "Configuring node for server deployment..."
    
    CONFIG_FILE="$CONFIG_DIR/config/config.toml"
    
    # Backup original config
    cp "$CONFIG_FILE" "$CONFIG_FILE.backup"
    
    # Configure RPC to allow external access
    sed -i 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' "$CONFIG_FILE"
    
    # Configure P2P to allow external connections
    sed -i 's|laddr = "tcp://0.0.0.0:26656"|laddr = "tcp://0.0.0.0:26656"|' "$CONFIG_FILE"
    
    # Enable Prometheus metrics
    sed -i 's|prometheus = false|prometheus = true|' "$CONFIG_FILE"
    
    # Configure CORS for web access
    sed -i 's|cors_allowed_origins = \[\]|cors_allowed_origins = \["*"\]|' "$CONFIG_FILE"
    
    print_success "Node configured for server deployment"
}

# Function to create systemd service
create_service() {
    print_status "Creating systemd service..."
    
    # Get user from user input or use default
    read -p "Enter service user (default: $USER): " SERVICE_USER
    SERVICE_USER=${SERVICE_USER:-$USER}
    
    # Create service file
    sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Fluentum Validator Node
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$HOME
ExecStart=/usr/local/bin/fluentumd start
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
    print_status "Starting Fluentum service..."
    
    sudo systemctl enable fluentumd
    sudo systemctl start fluentumd
    
    # Wait a moment and check status
    sleep 3
    
    if sudo systemctl is-active --quiet fluentumd; then
        print_success "Fluentum service started successfully"
    else
        print_error "Failed to start Fluentum service"
        sudo systemctl status fluentumd
        exit 1
    fi
}

# Function to show next steps
show_next_steps() {
    echo ""
    print_success "Deployment completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Check service status: sudo systemctl status fluentumd"
    echo "2. View logs: sudo journalctl -u fluentumd -f"
    echo "3. Check RPC endpoint: curl http://localhost:26657/status"
    echo "4. Configure firewall: sudo ufw allow 26656/tcp 26657/tcp 26660/tcp"
    echo "5. Set up monitoring: ./scripts/health_check.sh"
    echo ""
    echo "Configuration files:"
    echo "- Config: $CONFIG_DIR/config/config.toml"
    echo "- Genesis: $CONFIG_DIR/config/genesis.json"
    echo "- Keys: $CONFIG_DIR/config/priv_validator_key.json"
    echo ""
    echo "Useful commands:"
    echo "- Stop service: sudo systemctl stop fluentumd"
    echo "- Restart service: sudo systemctl restart fluentumd"
    echo "- View logs: sudo journalctl -u fluentumd -f"
    echo "- Check sync: curl http://localhost:26657/status | jq '.result.sync_info.catching_up'"
}

# Main function
main() {
    echo "=========================================="
    echo "    Continue Fluentum Deployment"
    echo "=========================================="
    echo ""
    
    echo "Binary is already installed. What would you like to do next?"
    echo "1. Initialize and configure node"
    echo "2. Create and start service"
    echo "3. Full remaining deployment (initialize, configure, start)"
    echo "4. Exit"
    echo ""
    
    read -p "Enter your choice (1-4): " choice
    
    case $choice in
        1)
            fix_config_line_endings
            initialize_node
            configure_node
            print_success "Node initialized and configured"
            ;;
        2)
            create_service
            start_service
            show_next_steps
            ;;
        3)
            fix_config_line_endings
            initialize_node
            configure_node
            create_service
            start_service
            show_next_steps
            ;;
        4)
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