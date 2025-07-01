#!/bin/bash

# Fluentum Public Testnet Setup Script
# This script sets up a testnet node on one of the 4 servers

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Testnet configuration
TESTNET_CHAIN_ID="fluentum-testnet-1"
TESTNET_HOME="/opt/fluentum"
TESTNET_USER="fluentum"

# Server configurations
declare -A SERVERS=(
    ["fluentum-node1"]="34.44.129.207"
    ["fluentum-node2"]="34.44.82.114"
    ["fluentum-node3"]="34.68.180.153"
    ["fluentum-node4"]="34.72.252.153"
)

# Parse command line arguments
NODE_NAME=${1:-"fluentum-node1"}
NODE_INDEX=${2:-1}

# Validate node name
if [[ ! ${SERVERS[$NODE_NAME]+_} ]]; then
    print_error "Invalid node name: $NODE_NAME"
    echo "Valid options: ${!SERVERS[@]}"
    exit 1
fi

SERVER_IP=${SERVERS[$NODE_NAME]}
P2P_PORT=26656
RPC_PORT=26657
API_PORT=1317

print_status "Setting up Fluentum Testnet Node"
echo "Node Name: $NODE_NAME"
echo "Server IP: $SERVER_IP"
echo "P2P Port: $P2P_PORT"
echo "RPC Port: $RPC_PORT"
echo "API Port: $API_PORT"
echo "Chain ID: $TESTNET_CHAIN_ID"
echo ""

# Check if running as root
if [[ $EUID -eq 0 ]]; then
    print_error "This script should not be run as root"
    exit 1
fi

# Check if fluentumd binary exists
if ! command -v ./build/fluentumd &> /dev/null; then
    print_error "fluentumd binary not found. Please build the project first:"
    echo "  make build"
    exit 1
fi

# Create testnet directory structure
print_status "Creating testnet directory structure..."
sudo mkdir -p $TESTNET_HOME
sudo chown $USER:$USER $TESTNET_HOME
mkdir -p $TESTNET_HOME/config
mkdir -p $TESTNET_HOME/data
mkdir -p $TESTNET_HOME/logs

# Initialize the node
print_status "Initializing node..."
if ./build/fluentumd init "$NODE_NAME" --chain-id "$TESTNET_CHAIN_ID" --home "$TESTNET_HOME"; then
    print_success "Node initialized successfully"
else
    print_warning "Node initialization failed, but continuing..."
fi

# Generate node key if not exists
if [ ! -f "$TESTNET_HOME/config/node_key.json" ]; then
    print_status "Generating node key..."
    ./build/fluentumd gen-node-key --home "$TESTNET_HOME"
    print_success "Node key generated"
fi

# Generate validator key if not exists
if [ ! -f "$TESTNET_HOME/config/priv_validator_key.json" ]; then
    print_status "Generating validator key..."
    ./build/fluentumd gen-validator-key --home "$TESTNET_HOME"
    print_success "Validator key generated"
fi

# Create testnet configuration
print_status "Creating testnet configuration..."
cat > "$TESTNET_HOME/config/config.toml" << EOF
# Fluentum Testnet Configuration
chain_id = "$TESTNET_CHAIN_ID"
moniker = "$NODE_NAME"

# Database backend: goleveldb (compatible with Tendermint)
db_backend = "goleveldb"
db_dir = "data"

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:$P2P_PORT"
external_address = "$SERVER_IP:$P2P_PORT"
seeds = ""
persistent_peers = ""

# RPC Configuration
[rpc]
laddr = "tcp://0.0.0.0:$RPC_PORT"
cors_allowed_origins = ["*"]
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["*"]
max_open_connections = 900
unsafe = false

# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:$API_PORT"

# Consensus Configuration (optimized for testnet)
[consensus]
timeout_propose = "1s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "1s"
create_empty_blocks = true
create_empty_blocks_interval = "10s"

# Mempool Configuration
[mempool]
version = "v0"
recheck = true
broadcast = true
size = 5000
max_txs_bytes = 1073741824
cache_size = 10000

# State Sync Configuration
[statesync]
enable = true
temp_dir = "/tmp/fluentum-statesync"

# Instrumentation
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
namespace = "tendermint"
EOF

print_success "Testnet configuration created"

# Create genesis file if not exists
if [ ! -f "$TESTNET_HOME/config/genesis.json" ]; then
    print_status "Creating genesis file..."
    cat > "$TESTNET_HOME/config/genesis.json" << EOF
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "$TESTNET_CHAIN_ID",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": 22020096,
      "max_gas": -1,
      "time_iota_ms": 1000
    },
    "evidence": {
      "max_age_num_blocks": 100000,
      "max_age_duration": 172800000000000,
      "max_bytes": 1048576
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
    print_success "Genesis file created"
fi

# Create systemd service file
print_status "Creating systemd service..."
sudo tee /etc/systemd/system/fluentum-testnet.service > /dev/null << EOF
[Unit]
Description=Fluentum Testnet Node - $NODE_NAME
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/build/fluentumd start \\
    --home $TESTNET_HOME \\
    --moniker $NODE_NAME \\
    --chain-id $TESTNET_CHAIN_ID \\
    --testnet \\
    --log_level info
Restart=on-failure
RestartSec=3
LimitNOFILE=65536
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Create update peers script
print_status "Creating peer update script..."
cat > "$TESTNET_HOME/update_peers.sh" << 'EOF'
#!/bin/bash

# Update persistent peers for testnet
TESTNET_HOME="/opt/fluentum"
CONFIG_FILE="$TESTNET_HOME/config/config.toml"

# Get all server IPs and P2P ports
declare -A SERVERS=(
    ["fluentum-node1"]="34.44.129.207:26656"
    ["fluentum-node2"]="34.44.82.114:26656"
    ["fluentum-node3"]="34.68.180.153:26656"
    ["fluentum-node4"]="34.72.252.153:26656"
)

# Build persistent peers string (exclude current node)
CURRENT_NODE=$(grep "moniker" $CONFIG_FILE | cut -d'"' -f2)
PERSISTENT_PEERS=""

for NODE in "${!SERVERS[@]}"; do
    if [ "$NODE" != "$CURRENT_NODE" ]; then
        if [ -n "$PERSISTENT_PEERS" ]; then
            PERSISTENT_PEERS="$PERSISTENT_PEERS,"
        fi
        PERSISTENT_PEERS="$PERSISTENT_PEERS${SERVERS[$NODE]}"
    fi
done

# Update config file
sed -i "s|persistent_peers = \"\"|persistent_peers = \"$PERSISTENT_PEERS\"|" $CONFIG_FILE

echo "Updated persistent peers: $PERSISTENT_PEERS"
EOF

chmod +x "$TESTNET_HOME/update_peers.sh"

# Create start script
print_status "Creating start script..."
cat > "$TESTNET_HOME/start_node.sh" << EOF
#!/bin/bash

# Start Fluentum testnet node
echo "Starting Fluentum testnet node: $NODE_NAME"
echo "Chain ID: $TESTNET_CHAIN_ID"
echo "Home: $TESTNET_HOME"
echo "P2P: $SERVER_IP:$P2P_PORT"
echo "RPC: http://$SERVER_IP:$RPC_PORT"
echo "API: http://$SERVER_IP:$API_PORT"
echo ""

# Update peers before starting
$TESTNET_HOME/update_peers.sh

# Start the node
$(pwd)/build/fluentumd start \\
    --home $TESTNET_HOME \\
    --moniker $NODE_NAME \\
    --chain-id $TESTNET_CHAIN_ID \\
    --testnet \\
    --log_level info
EOF

chmod +x "$TESTNET_HOME/start_node.sh"

# Reload systemd and enable service
print_status "Enabling systemd service..."
sudo systemctl daemon-reload
sudo systemctl enable fluentum-testnet.service

print_success "Testnet node setup complete!"
echo ""
echo "Node Information:"
echo "  Name: $NODE_NAME"
echo "  IP: $SERVER_IP"
echo "  P2P Port: $P2P_PORT"
echo "  RPC Port: $RPC_PORT"
echo "  API Port: $API_PORT"
echo "  Chain ID: $TESTNET_CHAIN_ID"
echo "  Home Directory: $TESTNET_HOME"
echo ""
echo "To start the node:"
echo "  sudo systemctl start fluentum-testnet.service"
echo "  sudo systemctl status fluentum-testnet.service"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u fluentum-testnet.service -f"
echo ""
echo "To start manually:"
echo "  $TESTNET_HOME/start_node.sh"
echo ""
echo "Configuration files:"
echo "  Config: $TESTNET_HOME/config/config.toml"
echo "  Genesis: $TESTNET_HOME/config/genesis.json"
echo "  Node Key: $TESTNET_HOME/config/node_key.json"
echo "  Validator Key: $TESTNET_HOME/config/priv_validator_key.json" 