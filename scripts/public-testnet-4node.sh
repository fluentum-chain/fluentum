#!/bin/bash

set -e

# Node configurations with specific IP addresses
declare -A NODE_CONFIGS=(
  ["fluentum-node1"]="34.44.129.207"
  ["fluentum-node3"]="34.44.82.114"
  ["fluentum-node4"]="34.68.180.153"
  ["fluentum-node5"]="34.72.252.153"
)

CHAIN_ID="fluentum-testnet-1"
CONFIG_TEMPLATE="config/testnet-config.toml"

# Setup each node with its specific configuration
for node_name in "${!NODE_CONFIGS[@]}"; do
  ip_address="${NODE_CONFIGS[$node_name]}"
  export FLUENTUM_HOME="/tmp/$node_name"
  
  echo -e "\n=== Setting up $node_name at $FLUENTUM_HOME (IP: $ip_address) ==="
  
  # Create node directory
  mkdir -p "$FLUENTUM_HOME/config"
  
  # Copy config template
  cp "$CONFIG_TEMPLATE" "$FLUENTUM_HOME/config/config.toml"
  
  # Update backend to pebble
  sed -i 's/backend = "goleveldb"/backend = "pebble"/' "$FLUENTUM_HOME/config/config.toml"
  
  # Update external_address with the specific IP
  sed -i "s/external_address = \"\"/external_address = \"$ip_address:26656\"/" "$FLUENTUM_HOME/config/config.toml"
  
  # Update moniker
  sed -i "s/moniker = \"fluentum-testnet-node\"/moniker = \"$node_name\"/" "$FLUENTUM_HOME/config/config.toml"
  
  # Initialize the node (without --testnet flag)
  fluentumd init "$node_name" --chain-id $CHAIN_ID --home "$FLUENTUM_HOME"
  
  # Show configuration
  ls -la "$FLUENTUM_HOME"
  ls -la "$FLUENTUM_HOME/config/"
  echo "Backend configuration:"
  grep -A 2 -B 2 "backend" "$FLUENTUM_HOME/config/config.toml"
  echo "External address configuration:"
  grep "external_address" "$FLUENTUM_HOME/config/config.toml"
  
  echo "$node_name config complete."
done

echo -e "\nAll nodes configured with specific IP addresses:"
for node_name in "${!NODE_CONFIGS[@]}"; do
  echo "  $node_name: ${NODE_CONFIGS[$node_name]}"
done

echo -e "\nTo start a specific node, run:"
echo "fluentumd start --home /tmp/fluentum-node1"
echo "fluentumd start --home /tmp/fluentum-node3"
echo "fluentumd start --home /tmp/fluentum-node4"
echo "fluentumd start --home /tmp/fluentum-node5" 