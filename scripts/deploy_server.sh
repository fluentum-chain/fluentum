#!/bin/bash

# Fluentum Server Deployment Script
# This script helps deploy Fluentum on a Linux server

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
DEFAULT_USER="fluentum"
BINARY_PATH="/usr/local/bin/fluentumd"
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

# Function to check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.18+ first."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.18"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION+"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to build the binary
build_binary() {
    print_status "Building Fluentum binary..."
    
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Please run this script from the Fluentum project root."
        exit 1
    fi
    
    go build -o fluentumd ./cmd/fluentum
    
    if [ ! -f "fluentumd" ]; then
        print_error "Failed to build fluentumd binary"
        exit 1
    fi
    
    print_success "Binary built successfully"
}

# Function to install binary
install_binary() {
    print_status "Installing binary to $BINARY_PATH..."
    
    sudo cp fluentumd "$BINARY_PATH"
    sudo chmod +x "$BINARY_PATH"
    
    print_success "Binary installed successfully"
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
ExecStart=$BINARY_PATH start
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

# Function to show status
show_status() {
    print_status "Checking service status..."
    
    sudo systemctl status fluentumd --no-pager
    
    print_status "Recent logs:"
    sudo journalctl -u fluentumd -n 20 --no-pager
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

# Main deployment function
main() {
    echo "=========================================="
    echo "    Fluentum Server Deployment Script"
    echo "=========================================="
    echo ""
    
    check_root
    check_prerequisites
    
    # Ask user what they want to do
    echo "What would you like to do?"
    echo "1. Full deployment (build, install, configure, start)"
    echo "2. Build and install binary only"
    echo "3. Initialize and configure node only"
    echo "4. Create and start service only"
    echo "5. Show service status"
    echo "6. Exit"
    echo ""
    
    read -p "Enter your choice (1-6): " choice
    
    case $choice in
        1)
            build_binary
            install_binary
            fix_config_line_endings
            initialize_node
            configure_node
            create_service
            start_service
            show_next_steps
            ;;
        2)
            build_binary
            install_binary
            print_success "Binary built and installed"
            ;;
        3)
            initialize_node
            configure_node
            print_success "Node initialized and configured"
            ;;
        4)
            create_service
            start_service
            show_next_steps
            ;;
        5)
            show_status
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