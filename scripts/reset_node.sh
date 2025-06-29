#!/bin/bash

echo "=== Node Reset Script ==="
echo "This script will reset the node data and reinitialize it properly"

# Stop any running processes
echo "Stopping any running fluentumd processes..."
pkill -f fluentumd || true
sleep 2

# Backup current data (optional)
echo "Creating backup of current data..."
if [ -d "/opt/fluentum" ]; then
    sudo cp -r /opt/fluentum /opt/fluentum.backup.$(date +%Y%m%d_%H%M%S) || echo "Backup failed, continuing..."
fi

# Remove existing data
echo "Removing existing node data..."
sudo rm -rf /opt/fluentum/data
sudo rm -rf /opt/fluentum/config/priv_validator_state.json
sudo rm -rf /opt/fluentum/config/priv_validator_key.json

# Reinitialize the node
echo "Reinitializing the node..."
fluentumd init fluentum-node1 --home /opt/fluentum --chain-id fluentum-testnet-1

# Generate genesis file if it doesn't exist
if [ ! -f "/opt/fluentum/config/genesis.json" ]; then
    echo "Generating genesis file..."
    fluentumd init-genesis --home /opt/fluentum --chain-id fluentum-testnet-1
fi

# Set proper permissions
echo "Setting proper permissions..."
sudo chown -R ktang:ktang /opt/fluentum || sudo chown -R $USER:$USER /opt/fluentum

# Update config for testnet
echo "Updating configuration for testnet..."
cat > /opt/fluentum/config/config.toml << EOF
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

proxy_app = "tcp://127.0.0.1:26658"
moniker = "fluentum-node1"
fast_sync = true

[consensus]
timeout_commit = "1s"
timeout_propose = "1s"
create_empty_blocks = true
create_empty_blocks_interval = "10s"

[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""
seeds = ""
max_num_inbound_peers = 40
max_num_outbound_peers = 10

[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = []
cors_allowed_methods = ["HEAD", "GET", "POST"]
cors_allowed_headers = ["Origin", "X-Requested-With", "Content-Type", "Accept"]

[mempool]
recheck = true
broadcast = true
wal_dir = "data/mempool.wal"

[instrumentation]
prometheus = false
prometheus_listen_addr = ":26660"
max_open_connections = 3
namespace = "tendermint"
EOF

echo "=== Node reset completed ==="
echo "You can now start the node with:"
echo "fluentumd start --home /opt/fluentum --moniker fluentum-node1 --chain-id fluentum-testnet-1 --testnet --log_level info" 