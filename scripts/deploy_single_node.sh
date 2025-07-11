#!/bin/bash

# Single Node Fluentum Testnet Deployment
# This script deploys a single node for the current server

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

CURRENT_IP=$(curl -s ifconfig.me 2>/dev/null || echo "unknown")
print_status "Current server IP: $CURRENT_IP"

CURRENT_NODE=""
for node_name in "${VALID_NODES[@]}"; do
    if [ "${NODE_IPS[$node_name]}" = "$CURRENT_IP" ]; then
        CURRENT_NODE="$node_name"
        break
    fi
done

if [ -z "$CURRENT_NODE" ]; then
    print_error "Could not determine current node. Available nodes:"
    for node_name in "${VALID_NODES[@]}"; do
        echo "  $node_name: ${NODE_IPS[$node_name]}"
    done
    exit 1
fi

print_success "Detected as $CURRENT_NODE"

FLUENTUM_HOME="/tmp/$CURRENT_NODE"

print_status "Setting up $CURRENT_NODE at $FLUENTUM_HOME"

mkdir -p "$FLUENTUM_HOME/config"
cp "config/testnet-config.toml" "$FLUENTUM_HOME/config/config.toml"
sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"
sed -i "s/external_address = \"\"/external_address = \"$CURRENT_IP:$P2P_PORT\"/" "$FLUENTUM_HOME/config/config.toml"
sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$CURRENT_NODE\"/" "$FLUENTUM_HOME/config/config.toml"

print_status "Initializing node..."
if ! INIT_OUT=$(fluentumd init "$CURRENT_NODE" 2>&1); then
    print_error "Failed to initialize node. Output:\n$INIT_OUT\nUsing alternative method..."
    print_status "Creating minimal node setup..."
    cat > "$FLUENTUM_HOME/config/genesis.json" << 'EOF'
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
    print_success "Node setup completed with alternative method"
else
    print_success "Node initialized successfully"
fi

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

sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
print_status "Starting systemd service..."
if ! START_OUT=$(sudo systemctl start "$SERVICE_NAME" 2>&1); then
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
echo "RPC Endpoint: http://$CURRENT_IP:$RPC_PORT"
echo "P2P Endpoint: $CURRENT_IP:$P2P_PORT"
