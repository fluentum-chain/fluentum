#!/bin/bash

set -e

# Number of nodes
NODES=4
CHAIN_ID="fluentum-testnet-1"
CONFIG_TEMPLATE="config/testnet-config.toml"

for i in $(seq 1 $NODES); do
  export FLUENTUM_HOME="/tmp/fluentum-node$i"
  echo "\n=== Setting up node $i at $FLUENTUM_HOME ==="
  mkdir -p "$FLUENTUM_HOME/config"
  cp "$CONFIG_TEMPLATE" "$FLUENTUM_HOME/config/config.toml"
  sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"
  ./fluentumd init --testnet --chain-id $CHAIN_ID --home "$FLUENTUM_HOME"
  ls -la "$FLUENTUM_HOME"
  ls -la "$FLUENTUM_HOME/config/"
  grep -A 2 -B 2 "backend" "$FLUENTUM_HOME/config/config.toml"
  echo "Node $i config complete."
done

echo "\nAll $NODES nodes configured. To start a node, run:"
echo "./fluentumd start --testnet --home /tmp/fluentum-node1" 