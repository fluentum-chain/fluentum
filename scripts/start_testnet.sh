#!/bin/bash

# Fluentum Testnet Startup Script
# This script helps you start a Fluentum node in testnet mode

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_MONIKER="fluentum-testnet-node"
DEFAULT_CHAIN_ID="fluentum-testnet-1"
DEFAULT_HOME_DIR="$HOME/.fluentum"
DEFAULT_SEEDS=""
DEFAULT_PERSISTENT_PEERS=""

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Fluentum Testnet Node Setup  ${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to check if fluentumd is installed
check_fluentumd() {
    if ! command -v fluentumd &> /dev/null; then
        print_error "fluentumd is not installed or not in PATH"
        echo "Please build and install fluentumd first:"
        echo "  make build"
        echo "  make install"
        exit 1
    fi
    print_status "Found fluentumd: $(fluentumd version)"
}

# Function to initialize node if needed
initialize_node() {
    local home_dir=$1
    local moniker=$2
    local chain_id=$3

    if [ ! -d "$home_dir/config" ]; then
        print_status "Initializing new node..."
        fluentumd init "$moniker" --chain-id "$chain_id" --home "$home_dir"
        print_status "Node initialized successfully"
    else
        print_status "Node already initialized at $home_dir"
    fi
}

# Function to configure the node
configure_node() {
    local home_dir=$1
    local moniker=$2
    local seeds=$3
    local persistent_peers=$4

    local config_file="$home_dir/config/config.toml"
    local app_config_file="$home_dir/config/app.toml"

    print_status "Configuring node..."

    # Update config.toml
    if [ -f "$config_file" ]; then
        # Set moniker
        sed -i "s/moniker = \".*\"/moniker = \"$moniker\"/" "$config_file"

        # Configure P2P
        sed -i 's/laddr = "tcp:\/\/127.0.0.1:26656"/laddr = "tcp:\/\/0.0.0.0:26656"/' "$config_file"
        
        # Set seeds if provided
        if [ ! -z "$seeds" ]; then
            sed -i "s/seeds = \"\"/seeds = \"$seeds\"/" "$config_file"
        fi

        # Set persistent peers if provided
        if [ ! -z "$persistent_peers" ]; then
            sed -i "s/persistent_peers = \"\"/persistent_peers = \"$persistent_peers\"/" "$config_file"
        fi

        # Configure RPC
        sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' "$config_file"

        # Configure consensus for testnet (faster block times)
        sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/' "$config_file"
        sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/' "$config_file"
        sed -i 's/create_empty_blocks = true/create_empty_blocks = true/' "$config_file"
        sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "10s"/' "$config_file"

        print_status "config.toml updated"
    fi

    # Update app.toml
    if [ -f "$app_config_file" ]; then
        # Enable API
        sed -i 's/enable = false/enable = true/' "$app_config_file"
        sed -i 's/swagger = false/swagger = true/' "$app_config_file"
        sed -i 's/address = "tcp:\/\/0.0.0.0:1317"/address = "tcp:\/\/0.0.0.0:1317"/' "$app_config_file"

        # Enable gRPC
        sed -i 's/enable = false/enable = true/' "$app_config_file"
        sed -i 's/address = "0.0.0.0:9090"/address = "0.0.0.0:9090"/' "$app_config_file"

        # Enable gRPC-Web
        sed -i 's/enable = false/enable = true/' "$app_config_file"
        sed -i 's/address = "0.0.0.0:9091"/address = "0.0.0.0:9091"/' "$app_config_file"

        print_status "app.toml updated"
    fi
}

# Function to create genesis account if needed
create_genesis_account() {
    local home_dir=$1
    local account_name=$2
    local coins=$3

    if [ ! -z "$account_name" ] && [ ! -z "$coins" ]; then
        print_status "Creating genesis account: $account_name with $coins"
        
        # Add key if it doesn't exist
        if ! fluentumd keys show "$account_name" --keyring-backend test --home "$home_dir" &> /dev/null; then
            fluentumd keys add "$account_name" --keyring-backend test --home "$home_dir" --output json --no-backup
        fi

        # Get address
        local address=$(fluentumd keys show "$account_name" -a --keyring-backend test --home "$home_dir")
        
        # Add genesis account
        fluentumd add-genesis-account "$address" "$coins" --home "$home_dir" --keyring-backend test
        
        print_status "Genesis account created: $address"
    fi
}

# Function to start the node
start_node() {
    local home_dir=$1
    local chain_id=$2
    local background=$3

    print_status "Starting Fluentum testnet node..."
    echo "Chain ID: $chain_id"
    echo "Home directory: $home_dir"
    echo "RPC endpoint: http://localhost:26657"
    echo "API endpoint: http://localhost:1317"
    echo "P2P endpoint: localhost:26656"
    echo ""

    if [ "$background" = "true" ]; then
        print_status "Starting node in background..."
        nohup fluentumd start \
            --home "$home_dir" \
            --chain-id "$chain_id" \
            --testnet \
            --api \
            --grpc \
            --grpc-web \
            > fluentum-testnet.log 2>&1 &
        
        local pid=$!
        echo $pid > fluentum-testnet.pid
        print_status "Node started in background (PID: $pid)"
        print_status "Logs: tail -f fluentum-testnet.log"
        print_status "Stop: kill $pid"
    else
        print_status "Starting node in foreground..."
        fluentumd start \
            --home "$home_dir" \
            --chain-id "$chain_id" \
            --testnet \
            --api \
            --grpc \
            --grpc-web
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -m, --moniker NAME        Node moniker (default: $DEFAULT_MONIKER)"
    echo "  -c, --chain-id ID         Chain ID (default: $DEFAULT_CHAIN_ID)"
    echo "  -h, --home DIR            Home directory (default: $DEFAULT_HOME_DIR)"
    echo "  -s, --seeds SEEDS         Comma-separated list of seed nodes"
    echo "  -p, --peers PEERS         Comma-separated list of persistent peers"
    echo "  -a, --account NAME        Genesis account name"
    echo "  -b, --background          Start node in background"
    echo "  --help                    Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Start with default settings"
    echo "  $0 -m my-node -c test-chain-1        # Start with custom moniker and chain ID"
    echo "  $0 -s node1:26656,node2:26656        # Start with seed nodes"
    echo "  $0 -a validator -b                   # Start with genesis account in background"
    echo ""
}

# Parse command line arguments
MONIKER="$DEFAULT_MONIKER"
CHAIN_ID="$DEFAULT_CHAIN_ID"
HOME_DIR="$DEFAULT_HOME_DIR"
SEEDS="$DEFAULT_SEEDS"
PERSISTENT_PEERS="$DEFAULT_PERSISTENT_PEERS"
GENESIS_ACCOUNT=""
GENESIS_COINS="1000000000ufluentum,1000000000stake"
BACKGROUND=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--moniker)
            MONIKER="$2"
            shift 2
            ;;
        -c|--chain-id)
            CHAIN_ID="$2"
            shift 2
            ;;
        -h|--home)
            HOME_DIR="$2"
            shift 2
            ;;
        -s|--seeds)
            SEEDS="$2"
            shift 2
            ;;
        -p|--peers)
            PERSISTENT_PEERS="$2"
            shift 2
            ;;
        -a|--account)
            GENESIS_ACCOUNT="$2"
            shift 2
            ;;
        -b|--background)
            BACKGROUND=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
print_header

# Check prerequisites
check_fluentumd

# Initialize node
initialize_node "$HOME_DIR" "$MONIKER" "$CHAIN_ID"

# Configure node
configure_node "$HOME_DIR" "$MONIKER" "$SEEDS" "$PERSISTENT_PEERS"

# Create genesis account if specified
if [ ! -z "$GENESIS_ACCOUNT" ]; then
    create_genesis_account "$HOME_DIR" "$GENESIS_ACCOUNT" "$GENESIS_COINS"
fi

# Start the node
start_node "$HOME_DIR" "$CHAIN_ID" "$BACKGROUND"

print_status "Setup complete!"
echo ""
echo "Node endpoints:"
echo "  RPC:     http://localhost:26657"
echo "  API:     http://localhost:1317"
echo "  gRPC:    localhost:9090"
echo "  gRPC-Web: localhost:9091"
echo "  P2P:     localhost:26656"
echo ""
echo "Useful commands:"
echo "  fluentumd status --home $HOME_DIR"
echo "  fluentumd query bank balances --home $HOME_DIR"
echo "  fluentumd tendermint show-node-id --home $HOME_DIR"
echo "" 