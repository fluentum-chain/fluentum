#!/bin/bash

# Unified Fluentum Testnet Node Deployment Script
# Usage: ./deploy_node.sh <node-name> <node-index>

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

print_error() {
    echo -e "$(now) ${RED}[ERROR]${NC} $1" >&2
}

# Source centralized node configuration
source "$(dirname "$0")/nodes.conf"

# Check arguments
if [ $# -ne 2 ]; then
    print_error "Usage: $0 <node-name> <node-index>"
    echo "Example: $0 fluentum-node1 1"
    exit 1
fi

NODE_NAME=$1
NODE_INDEX=$2

# Validate node name
if [[ ! " ${VALID_NODES[@]} " =~ " ${NODE_NAME} " ]]; then
    print_error "Invalid node name: $NODE_NAME"
    echo "Valid options: ${VALID_NODES[*]}"
    exit 1
fi

NODE_IP="${NODE_IPS[$NODE_NAME]}"

print_status "Deploying Fluentum testnet node: $NODE_NAME (index: $NODE_INDEX, IP: $NODE_IP)"

# Check for fluentumd binary in multiple possible locations
FLUENTUMD=""
# Dynamically detect Go binary output directory
GO_BIN_PATH=""
if command -v go &> /dev/null; then
    GO_BIN_PATH=$(go env GOBIN)
    if [ -z "$GO_BIN_PATH" ]; then
        GO_BIN_PATH="$(go env GOPATH)/bin"
    fi
    echo "[DEBUG] Detected Go binary path: $GO_BIN_PATH"
fi

POSSIBLE_PATHS=(
    "/usr/local/bin/fluentumd"
    "$HOME/go/bin/fluentumd"
    "$GOPATH/bin/fluentumd"
    "$GO_BIN_PATH/fluentumd"
    "$(pwd)/build/fluentumd"
    "$(pwd)/bin/fluentumd"
)

for path in "${POSSIBLE_PATHS[@]}"; do
    if [ -f "$path" ]; then
        FLUENTUMD="$path"
        print_status "Found fluentumd at $FLUENTUMD"
        break
    fi
done

if [ -z "$FLUENTUMD" ]; then
    print_status "fluentumd binary not found in any standard locations. Building from source..."
    
    # Ensure Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH. Please install Go 1.18+ and try again."
        exit 1
    fi
    
    # Build fluentumd
    print_status "Building fluentumd..."
    if ! BUILD_OUT=$(make install 2>&1); then
        print_error "Failed to build fluentumd. Output:\n$BUILD_OUT"
        echo "Make sure you have all build dependencies installed."
        echo "You may need to install: build-essential git"
        exit 1
    fi
    
    # Check build output locations
    for path in "${POSSIBLE_PATHS[@]}"; do
        if [ -f "$path" ]; then
            FLUENTUMD="$path"
            print_success "Successfully built fluentumd at $FLUENTUMD"
            break
        fi
    done
    
    if [ -z "$FLUENTUMD" ]; then
        print_error "fluentumd binary not found after build. Checked paths: ${POSSIBLE_PATHS[*]}"
        echo "Build output was:"
        echo "$BUILD_OUT"
        exit 1
    fi
else
    print_status "Found existing fluentumd at $FLUENTUMD"
fi

# Ensure fluentumd is in /usr/local/bin for systemd service
if [ "$FLUENTUMD" != "/usr/local/bin/fluentumd" ]; then
    print_status "Installing fluentumd to /usr/local/bin/..."
    sudo cp "$FLUENTUMD" /usr/local/bin/
    sudo chmod +x /usr/local/bin/fluentumd
    FLUENTUMD="/usr/local/bin/fluentumd"
    print_success "fluentumd installed to $FLUENTUMD"
fi

# Set home directory
FLUENTUM_HOME="/opt/fluentum/$NODE_NAME"

# Create home directory
sudo mkdir -p "$FLUENTUM_HOME"
sudo chown $USER:$USER "$FLUENTUM_HOME"
mkdir -p "$FLUENTUM_HOME/config" "$FLUENTUM_HOME/data" "$FLUENTUM_HOME/logs"

# Copy and customize config.toml
if [ ! -f "config/testnet-config.toml" ]; then
    print_error "config/testnet-config.toml not found."
    exit 1
fi
cp config/testnet-config.toml "$FLUENTUM_HOME/config/config.toml"

# Set node configuration
sed -i "s/^moniker *=.*/moniker = \"$NODE_NAME\"/" "$FLUENTUM_HOME/config/config.toml"
sed -i "s|^external_address *=.*|external_address = \"$NODE_IP:$P2P_PORT\"|" "$FLUENTUM_HOME/config/config.toml"
sed -i "s|^laddr *=.*|laddr = \"tcp://0.0.0.0:$P2P_PORT\"|" "$FLUENTUM_HOME/config/config.toml"
sed -i "s|^laddr *=.*|laddr = \"tcp://0.0.0.0:$RPC_PORT\"|" "$FLUENTUM_HOME/config/config.toml"

# Set persistent peers if configured for this node
if [ -n "${PERSISTENT_PEERS[$NODE_NAME]}" ]; then
    print_status "Setting persistent peers for $NODE_NAME"
    sed -i "s|^persistent_peers *=.*|persistent_peers = \"${PERSISTENT_PEERS[$NODE_NAME]}\"|" "$FLUENTUM_HOME/config/config.toml"
else
    print_status "No persistent peers configured for $NODE_NAME"
fi

# Initialize the node
print_status "Initializing node..."
# Set HOME environment variable for the init command
if ! INIT_OUT=$(HOME="$FLUENTUM_HOME" $FLUENTUMD init "$NODE_NAME" 2>&1); then
    print_error "Node initialization failed. Output:\n$INIT_OUT"
    exit 1
else
    print_success "Node initialized successfully"
fi

# Check for genesis file and copy it if it exists
if [ -f "config/genesis.json" ]; then
    print_status "Copying genesis file to node config..."
    cp config/genesis.json "$FLUENTUM_HOME/config/"
    print_success "Genesis file copied successfully"
else
    print_error "genesis.json not found in config/ directory. Please ensure you have a valid genesis file."
    print_status "You can obtain the genesis file from an existing node or generate one using 'fluentumd init-genesis'"
    print_status "Place it in the config/ directory and run this script again."
    exit 1
fi

# Create systemd service
SERVICE_NAME="fluentum-$NODE_NAME"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

print_status "Creating systemd service..."
sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Fluentum Testnet Node - $NODE_NAME
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$FLUENTUM_HOME
ExecStart=$FLUENTUMD start --home $FLUENTUM_HOME --moniker $NODE_NAME --chain-id fluentum-testnet-1 --testnet --log_level info
Restart=on-failure
RestartSec=3
LimitNOFILE=4096
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
print_status "Starting systemd service..."
if ! START_OUT=$(sudo systemctl restart "$SERVICE_NAME" 2>&1); then
    print_error "Failed to start service. Output:\n$START_OUT"
    sudo systemctl status "$SERVICE_NAME"
    exit 1
fi

sleep 3
if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    print_success "Service started successfully"
else
    print_error "Failed to start service"
    sudo systemctl status "$SERVICE_NAME"
    exit 1
fi

print_success "Node deployment completed!"
echo ""
echo "Node Information:"
echo "  Node Name: $NODE_NAME"
echo "  IP Address: $NODE_IP"
echo "  Home Directory: $FLUENTUM_HOME"
echo "  Service Name: $SERVICE_NAME"
echo ""
echo "Useful Commands:"
echo "  Check status: sudo systemctl status $SERVICE_NAME"
echo "  View logs: sudo journalctl -u $SERVICE_NAME -f"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Stop: sudo systemctl stop $SERVICE_NAME"
echo ""
echo "RPC Endpoint: http://$NODE_IP:$RPC_PORT"
echo "P2P Endpoint: $NODE_IP:$P2P_PORT" 