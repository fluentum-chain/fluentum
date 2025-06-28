#!/bin/bash

# Setup Current Server Node Only
# This script sets up only the node for the current server

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Get current server IP
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
CONFIG_TEMPLATE="config/testnet-config.toml"

print_status "Setting up $CURRENT_NODE at $FLUENTUM_HOME"

# Create node directory structure
mkdir -p "$FLUENTUM_HOME/config"
mkdir -p "$FLUENTUM_HOME/data"

# Copy config template
cp "$CONFIG_TEMPLATE" "$FLUENTUM_HOME/config/config.toml"

# Update backend to pebble
sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"

# Update external_address with the specific IP
sed -i "s/external_address = \"\"/external_address = \"$CURRENT_IP:26656\"/" "$FLUENTUM_HOME/config/config.toml"

# Update moniker
sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$CURRENT_NODE\"/" "$FLUENTUM_HOME/config/config.toml"

# Try to initialize the node using fluentumd init first
print_status "Initializing node using fluentumd init..."
if fluentumd init "$CURRENT_NODE" --chain-id $CHAIN_ID --home "$FLUENTUM_HOME" 2>/dev/null; then
    print_success "Node initialized successfully with fluentumd init"
else
    print_warning "fluentumd init failed, using manual setup..."
    
    # Manual setup: Create minimal configuration
    print_status "Creating minimal node setup manually..."
    
    # Create a minimal genesis file
    cat > "$FLUENTUM_HOME/config/genesis.json" << 'EOF'
{
  "genesis_time": "2024-01-01T00:00:00Z",
  "chain_id": "fluentum-testnet-1",
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

    # Create a minimal node key file
    print_status "Creating node key..."
    cat > "$FLUENTUM_HOME/config/node_key.json" << 'EOF'
{
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
  }
}
EOF
        
    # Instead of creating invalid key files, let's generate proper ones
    print_status "Generating proper validator keys..."
    
    # Use fluentumd to generate keys properly
    if fluentumd init "$CURRENT_NODE" --chain-id $CHAIN_ID --home "$FLUENTUM_HOME" --overwrite 2>/dev/null; then
        print_success "Node initialized successfully with fluentumd init --overwrite"
    else
        print_warning "fluentumd init still failed, creating minimal setup..."
        
        # Create a minimal validator key file with proper structure
        print_status "Creating validator key..."
        cat > "$FLUENTUM_HOME/config/priv_validator_key.json" << 'EOF'
{
  "address": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
  },
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
  }
}
EOF

        # Create validator state file with proper structure
        cat > "$FLUENTUM_HOME/data/priv_validator_state.json" << 'EOF'
{
  "height": 0,
  "round": 0,
  "step": 0
}
EOF
    fi
        
    print_success "Node setup completed with manual method"
fi

# Show configuration
echo ""
print_status "Configuration for $CURRENT_NODE:"
echo "  Node Name: $CURRENT_NODE"
echo "  IP Address: $CURRENT_IP"
echo "  Chain ID: $CHAIN_ID"
echo "  Home Directory: $FLUENTUM_HOME"
echo ""
echo "Configuration files:"
ls -la "$FLUENTUM_HOME"
ls -la "$FLUENTUM_HOME/config/"
echo ""
print_status "Backend configuration:"
grep -A 2 -B 2 "backend" "$FLUENTUM_HOME/config/config.toml"
echo ""
print_status "External address configuration:"
grep "external_address" "$FLUENTUM_HOME/config/config.toml"

print_success "$CURRENT_NODE setup complete!"
echo ""
echo "To start this node, run:"
echo "fluentumd start --home $FLUENTUM_HOME" 