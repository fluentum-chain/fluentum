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

# Remove ALL existing data completely
echo "Removing ALL existing node data..."
sudo rm -rf /opt/fluentum/data
sudo rm -rf /opt/fluentum/config/priv_validator_state.json
sudo rm -rf /opt/fluentum/config/priv_validator_key.json
sudo rm -rf /opt/fluentum/config/node_key.json
sudo rm -rf /opt/fluentum/config/genesis.json

# Also remove any application database files
echo "Removing application database files..."
sudo rm -rf /opt/fluentum/data/application.db
sudo rm -rf /opt/fluentum/data/blockstore.db
sudo rm -rf /opt/fluentum/data/state.db
sudo rm -rf /opt/fluentum/data/tx_index.db
sudo rm -rf /opt/fluentum/data/evidence.db
sudo rm -rf /opt/fluentum/data/application
sudo rm -rf /opt/fluentum/data/blockstore
sudo rm -rf /opt/fluentum/data/state
sudo rm -rf /opt/fluentum/data/tx_index
sudo rm -rf /opt/fluentum/data/evidence

# Recreate directories
echo "Recreating directories..."
sudo mkdir -p /opt/fluentum/data
sudo mkdir -p /opt/fluentum/config

# Reinitialize the node completely
echo "Reinitializing the node completely..."
fluentumd init fluentum-node1 --home /opt/fluentum --chain-id fluentum-testnet-1

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