#!/bin/bash

# Fluentum Node Deployment Script
# Version: 2.0.0
# Description: Deploys a Fluentum node with quantum signing and AI validation features

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
NODE_ID=""
NETWORK="mainnet"
VERSION="v0.2.0-alpha"
FEATURES="quantum_signing,ai_validation"
RPC_PORT=26657
P2P_PORT=26656
API_PORT=1317
GRPC_PORT=9090
HOME_DIR="$HOME/.fluentum"
CONFIG_DIR="$HOME_DIR/config"
DATA_DIR="$HOME_DIR/data"
LOG_LEVEL="info"

# Display usage information
function show_usage() {
    echo -e "${YELLOW}Usage:${NC} $0 [options]"
    echo "Options:"
    echo "  -i, --node-id     Node identifier (required)"
    echo "  -n, --network     Network to join (mainnet|testnet|devnet), default: $NETWORK"
    echo "  -v, --version     Fluentum version to install, default: $VERSION"
    echo "  -f, --features    Comma-separated list of features to enable, default: $FEATURES"
    echo "  -h, --help        Show this help message"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -i|--node-id)
            NODE_ID="$2"
            shift # past argument
            shift # past value
            ;;
        -n|--network)
            NETWORK="$2"
            shift
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift
            shift
            ;;
        -f|--features)
            FEATURES="$2"
            shift
            shift
            ;;
        -h|--help)
            show_usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            ;;
    esac
done

# Validate required parameters
if [ -z "$NODE_ID" ]; then
    echo -e "${RED}Error: Node ID is required${NC}"
    show_usage
fi

# Display deployment information
echo -e "${GREEN}=== Fluentum Node Deployment ===${NC}"
echo -e "Node ID:    ${YELLOW}$NODE_ID${NC}"
echo -e "Network:    ${YELLOW}$NETWORK${NC}"
echo -e "Version:    ${YELLOW}$VERSION${NC}"
echo -e "Features:   ${YELLOW}$FEATURES${NC}"
echo -e "Home Dir:   ${YELLOW}$HOME_DIR${NC}"
echo ""

# Install system dependencies
echo -e "${GREEN}[1/8] Installing system dependencies...${NC}"
sudo apt-get update
sudo apt-get install -y git build-essential make jq curl wget lz4 unzip

# Install Go
echo -e "${GREEN}[2/8] Installing Go...${NC}"
if ! command -v go &> /dev/null; then
    wget https://dl.google.com/go/go1.21.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    echo 'export GOPATH=$HOME/go' >> ~/.profile
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.profile
    source ~/.profile
    rm go1.21.0.linux-amd64.tar.gz
fi

# Clone or update the repository
echo -e "${GREEN}[3/8] Setting up Fluentum repository...${NC}"
if [ -d "$HOME/fluentum" ]; then
    cd $HOME/fluentum
    git fetch --all
    git checkout $VERSION
    git pull origin $VERSION
else
    git clone https://github.com/fluentum-chain/fluentum.git $HOME/fluentum
    cd $HOME/fluentum
    git checkout $VERSION
fi

# Build and install
echo -e "${GREEN}[4/8] Building Fluentum...${NC}"
cd $HOME/fluentum
make install

# Initialize the node
echo -e "${GREEN}[5/8] Initializing node...${NC}"
fluentumd init "$NODE_ID" --chain-id $NETWORK --home $HOME_DIR

# Download genesis and address book
echo -e "${GREEN}[6/8] Configuring node...${NC}"
case $NETWORK in
    mainnet)
        GENESIS_URL="https://raw.githubusercontent.com/fluentum-chain/networks/main/$NETWORK/genesis.json"
        ADDRBOOK_URL="https://raw.githubusercontent.com/fluentum-chain/networks/main/$NETWORK/addrbook.json"
        ;;
    testnet)
        GENESIS_URL="https://raw.githubusercontent.com/fluentum-chain/networks/main/$NETWORK/genesis.json"
        ADDRBOOK_URL="https://raw.githubusercontent.com/fluentum-chain/networks/main/$NETWORK/addrbook.json"
        ;;
    *)
        # For devnet or custom networks, use default genesis
        cp $HOME/fluentum/networks/devnet/genesis.json $CONFIG_DIR/
        ;;
esac

if [ -n "$GENESIS_URL" ]; then
    curl -s $GENESIS_URL > $CONFIG_DIR/genesis.json
    curl -s $ADDRBOOK_URL > $CONFIG_DIR/addrbook.json
fi

# Configure node
echo -e "${GREEN}[7/8] Configuring features...${NC}"

# Create or update config.toml with features
cat > $CONFIG_DIR/config.toml << EOF
# Fluentum Node Configuration

[features]
enabled = true
auto_reload = true
check_compatibility = true
min_node_version = "1.0.0"

# Quantum Signing Configuration
[features.quantum_signing]
enabled = true
dilithium_mode = 3
quantum_headers = true
enable_metrics = true
max_latency_ms = 1000

# AI Validation Configuration
[features.ai_validation]
enabled = true
model_path = ""
use_gpu = true
max_batch_size = 32
confidence_threshold = 0.9
enable_logging = true

# Node Configuration
[p2p]
laddr = "tcp://0.0.0.0:${P2P_PORT}"

[rpc]
laddr = "tcp://0.0.0.0:${RPC_PORT}"

[api]
address = "tcp://0.0.0.0:${API_PORT}"

[grpc]
address = "0.0.0.0:${GRPC_PORT}"

[telemetry]
enabled = true
service-name = "$NODE_ID"

[consensus]
timeout_commit = "5s"
timeout_propose = "3s"
EOF

# Create systemd service
echo -e "${GREEN}[8/8] Setting up systemd service...${NC}"
sudo tee /etc/systemd/system/fluentumd.service > /dev/null << EOF
[Unit]
Description=Fluentum Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=$(which fluentumd) start --home $HOME_DIR --log_level $LOG_LEVEL
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
sudo systemctl daemon-reload
sudo systemctl enable fluentumd
sudo systemctl restart fluentumd

# Display node information
echo -e "${GREEN}=== Deployment Complete ===${NC}"
echo -e "Node ID:     ${YELLOW}$NODE_ID${NC}"
echo -e "Network:     ${YELLOW}$NETWORK${NC}"
echo -e "Version:     ${YELLOW}$VERSION${NC}"
echo -e "Features:    ${YELLOW}$FEATURES${NC}"
echo -e "RPC:         ${YELLOW}http://$(curl -s ifconfig.me):$RPC_PORT${NC}"
echo -e "API:         ${YELLOW}http://$(curl -s ifconfig.me):$API_PORT${NC}"
echo -e "P2P:         ${YELLOW}$(curl -s ifconfig.me):$P2P_PORT${NC}"
echo -e "Home Dir:    ${YELLOW}$HOME_DIR${NC}"
echo -e "Logs:        ${YELLOW}journalctl -u fluentumd -f${NC}"

# Check node status
echo -e "\n${GREEN}Checking node status...${NC}"
sleep 5
systemctl status fluentumd --no-pager

# Display sync status
echo -e "\n${GREEN}Sync Status:${NC}"
curl -s http://localhost:${RPC_PORT}/status | jq -r '.result.sync_info'

echo -e "\n${GREEN}Deployment completed successfully!${NC}"
