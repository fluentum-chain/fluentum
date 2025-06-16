#!/bin/bash

# Check if required environment variables are set
if [ -z "$MONIKER" ]; then
    echo "Error: MONIKER environment variable is not set"
    exit 1
fi

if [ -z "$VALIDATOR_NAME" ]; then
    echo "Error: VALIDATOR_NAME environment variable is not set"
    exit 1
fi

# Set default chain ID if not provided
CHAIN_ID=${CHAIN_ID:-"fluentum-1"}

# Initialize node
echo "Initializing node with moniker: $MONIKER"
fluentumd init $MONIKER --chain-id $CHAIN_ID

# Download genesis if GENESIS_URL is provided
if [ ! -z "$GENESIS_URL" ]; then
    echo "Downloading genesis file from $GENESIS_URL"
    wget $GENESIS_URL -O ~/.fluentumd/config/genesis.json
else
    echo "Using local genesis file"
fi

# Configure node
echo "Configuring node..."
CONFIG_FILE=~/.fluentumd/config/config.toml

# Enable Prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' $CONFIG_FILE

# Configure seed nodes if provided
if [ ! -z "$SEED_NODES" ]; then
    sed -i "s|seed_nodes = \"\"|seed_nodes = \"$SEED_NODES\"|" $CONFIG_FILE
fi

# Configure persistent peers if provided
if [ ! -z "$PERSISTENT_PEERS" ]; then
    sed -i "s|persistent_peers = \"\"|persistent_peers = \"$PERSISTENT_PEERS\"|" $CONFIG_FILE
fi

# Configure RPC and P2P addresses
sed -i "s|laddr = \"tcp://127.0.0.1:26657\"|laddr = \"tcp://0.0.0.0:26657\"|" $CONFIG_FILE
sed -i "s|laddr = \"tcp://0.0.0.0:26656\"|laddr = \"tcp://0.0.0.0:26656\"|" $CONFIG_FILE

# Create validator keys
echo "Creating validator keys..."
fluentumd keys add $VALIDATOR_NAME --keyring-backend os

# Generate quantum-resistant keys
echo "Generating quantum-resistant keys..."
QUANTUM_KEY_FILE=~/.fluentumd/config/quantum_keys.json
fluentumd quantum gen-keys --output $QUANTUM_KEY_FILE

# Configure ZK prover if URL is provided
ZK_PROVER_URL=${ZK_PROVER_URL:-"https://zk.fluentum.net"}

# Create systemd service file
echo "Creating systemd service..."
cat > /etc/systemd/system/fluentumd.service << EOF
[Unit]
Description=Fluentum Validator Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=/usr/local/bin/fluentumd start \\
    --quantum.key-file=$QUANTUM_KEY_FILE \\
    --zk-prover-url=$ZK_PROVER_URL
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
echo "Enabling and starting service..."
sudo systemctl daemon-reload
sudo systemctl enable fluentumd
sudo systemctl start fluentumd

echo "Validator setup complete!"
echo "Check status with: sudo systemctl status fluentumd"
echo "View logs with: sudo journalctl -u fluentumd -f" 