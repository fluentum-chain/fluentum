#!/bin/bash

# Simple Fluentum Testnet Deployment
# Uses the existing public-testnet-4node.sh approach

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're on the right server
CURRENT_IP=$(curl -s ifconfig.me 2>/dev/null || echo "unknown")
print_status "Current server IP: $CURRENT_IP"

# Node configurations
declare -A NODE_CONFIGS=(
  ["fluentum-node1"]="34.44.129.207"
  ["fluentum-node3"]="34.44.82.114"
  ["fluentum-node4"]="34.68.180.153"
  ["fluentum-node5"]="34.72.252.153"
)

# Find which node we are
CURRENT_NODE=""
for node_name in "${!NODE_CONFIGS[@]}"; do
    if [ "${NODE_CONFIGS[$node_name]}" = "$CURRENT_IP" ]; then
        CURRENT_NODE="$node_name"
        break
    fi
done

if [ -z "$CURRENT_NODE" ]; then
    print_error "Could not determine current node. Available nodes:"
    for node_name in "${!NODE_CONFIGS[@]}"; do
        echo "  $node_name: ${NODE_CONFIGS[$node_name]}"
    done
    exit 1
fi

print_success "Detected as $CURRENT_NODE"

# Setup the current node
FLUENTUM_HOME="/tmp/$CURRENT_NODE"
CHAIN_ID="fluentum-testnet-1"

print_status "Setting up $CURRENT_NODE at $FLUENTUM_HOME"

# Create node directory
mkdir -p "$FLUENTUM_HOME/config"

# Copy config template
cp "config/testnet-config.toml" "$FLUENTUM_HOME/config/config.toml"

# Update backend to pebble
sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"

# Update external_address with the specific IP
sed -i "s/external_address = \"\"/external_address = \"$CURRENT_IP:26656\"/" "$FLUENTUM_HOME/config/config.toml"

# Update moniker
sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$CURRENT_NODE\"/" "$FLUENTUM_HOME/config/config.toml"

# Try to initialize the node
print_status "Initializing node..."
if fluentumd init "$CURRENT_NODE" --chain-id $CHAIN_ID --home "$FLUENTUM_HOME"; then
    print_success "Node initialized successfully"
else
    print_error "Failed to initialize node. Trying alternative approach..."
    
    # Alternative: Use the existing public-testnet-4node.sh approach
    print_status "Using alternative initialization method..."
    
    # Create a minimal genesis file
    cat > "$FLUENTUM_HOME/config/genesis.json" << EOF
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "$CHAIN_ID",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1",
      "time_iota_ms": "1000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000",
      "max_bytes": "1048576"
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    },
    "version": {}
  },
  "validators": [],
  "app_hash": "",
  "app_state": {}
}
EOF

    # Create node key
    fluentumd gen-node-key --home "$FLUENTUM_HOME"
    
    # Create validator key
    fluentumd gen-validator-key --home "$FLUENTUM_HOME"
    
    print_success "Node setup completed with alternative method"
fi

# Create systemd service
SERVICE_NAME="fluentum-$CURRENT_NODE"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

print_status "Creating systemd service..."

sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Fluentum Testnet Node - $CURRENT_NODE
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

# Reload systemd and start service
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
sudo systemctl start "$SERVICE_NAME"

# Wait and check status
sleep 3
if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    print_success "Service started successfully"
else
    print_error "Failed to start service"
    sudo systemctl status "$SERVICE_NAME"
    exit 1
fi

# Show summary
echo ""
print_success "Deployment completed for $CURRENT_NODE!"
echo ""
echo "Node Information:"
echo "  Node Name: $CURRENT_NODE"
echo "  IP Address: $CURRENT_IP"
echo "  Home Directory: $FLUENTUM_HOME"
echo "  Service Name: $SERVICE_NAME"
echo ""
echo "Useful Commands:"
echo "  Check status: sudo systemctl status $SERVICE_NAME"
echo "  View logs: sudo journalctl -u $SERVICE_NAME -f"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Stop: sudo systemctl stop $SERVICE_NAME"
echo ""
echo "RPC Endpoint: http://$CURRENT_IP:26657"
echo "P2P Endpoint: $CURRENT_IP:26656" 