#!/bin/bash

# Fluentum Node Initialization and Startup Script
# This script helps initialize and start a Fluentum node

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

# Default values
DEFAULT_HOME="/tmp/fluentum-new-test"
DEFAULT_MONIKER="fluentum-node"
DEFAULT_CHAIN_ID="fluentum-mainnet-1"
DEFAULT_TESTNET=false

# Parse command line arguments
HOME_DIR=${1:-$DEFAULT_HOME}
MONIKER=${2:-$DEFAULT_MONIKER}
CHAIN_ID=${3:-$DEFAULT_CHAIN_ID}
TESTNET=${4:-$DEFAULT_TESTNET}

print_status "Fluentum Node Setup Script"
echo "Home Directory: $HOME_DIR"
echo "Moniker: $MONIKER"
echo "Chain ID: $CHAIN_ID"
echo "Testnet Mode: $TESTNET"
echo ""

# Check if fluentumd binary exists
if ! command -v ./build/fluentumd &> /dev/null; then
    print_error "fluentumd binary not found. Please build the project first:"
    echo "  make build"
    exit 1
fi

print_status "Step 1: Initializing node..."
if ./build/fluentumd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"; then
    print_success "Node initialized successfully"
else
    print_warning "Node initialization failed, but continuing..."
fi

print_status "Step 2: Checking configuration..."
if [ -f "$HOME_DIR/config/config.toml" ]; then
    print_success "Configuration file found"
else
    print_warning "Configuration file not found, creating minimal config..."
    
    # Create minimal config.toml
    mkdir -p "$HOME_DIR/config"
    cat > "$HOME_DIR/config/config.toml" << EOF
# Fluentum Node Configuration
chain_id = "$CHAIN_ID"
moniker = "$MONIKER"

# Database backend: goleveldb (compatible with Tendermint)
db_backend = "goleveldb"
db_dir = "data"

[p2p]
laddr = "tcp://0.0.0.0:26656"

[rpc]
laddr = "tcp://0.0.0.0:26657"

[consensus]
timeout_commit = "5s"
timeout_propose = "3s"
EOF
    print_success "Minimal configuration created"
fi

print_status "Step 3: Checking node key..."
if [ -f "$HOME_DIR/config/node_key.json" ]; then
    print_success "Node key found"
else
    print_warning "Node key not found, generating..."
    if ./build/fluentumd gen-node-key --home "$HOME_DIR"; then
        print_success "Node key generated"
    else
        print_error "Failed to generate node key"
        exit 1
    fi
fi

print_status "Step 4: Checking validator key..."
if [ -f "$HOME_DIR/config/priv_validator_key.json" ]; then
    print_success "Validator key found"
else
    print_warning "Validator key not found, generating..."
    if ./build/fluentumd gen-validator-key --home "$HOME_DIR"; then
        print_success "Validator key generated"
    else
        print_error "Failed to generate validator key"
        exit 1
    fi
fi

print_status "Step 5: Checking genesis file..."
if [ -f "$HOME_DIR/config/genesis.json" ]; then
    print_success "Genesis file found"
else
    print_warning "Genesis file not found, creating minimal genesis..."
    
    # Create minimal genesis.json
    cat > "$HOME_DIR/config/genesis.json" << EOF
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
    print_success "Minimal genesis file created"
fi

print_status "Step 6: Starting node..."
echo ""
echo "Starting Fluentum node with the following configuration:"
echo "  Home Directory: $HOME_DIR"
echo "  Moniker: $MONIKER"
echo "  Chain ID: $CHAIN_ID"
echo "  RPC Endpoint: http://localhost:26657"
echo "  P2P Endpoint: localhost:26656"
echo ""

# Build the start command
START_CMD="./build/fluentumd start --home $HOME_DIR --moniker $MONIKER --chain-id $CHAIN_ID"

if [ "$TESTNET" = "true" ]; then
    START_CMD="$START_CMD --testnet"
    echo "  Testnet Mode: Enabled"
fi

echo "Command: $START_CMD"
echo ""

# Ask user if they want to start the node
read -p "Do you want to start the node now? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_status "Starting Fluentum node..."
    eval $START_CMD
else
    print_status "Node setup complete. To start the node manually, run:"
    echo "  $START_CMD"
fi

print_success "Setup complete!" 