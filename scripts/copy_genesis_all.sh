#!/bin/bash

# Usage: ./copy_genesis_all.sh /opt/fluentum/config/genesis.json
# Copies the given genesis.json to all Fluentum nodes

# List of node IPs (edit as needed)
NODES=(
  "34.44.82.114"   # fluentum-node2
  "34.68.180.153"  # fluentum-node3
  "34.72.252.153"  # fluentum-node4
)

USER="ktang"
GENESIS_SRC="${1:-/opt/fluentum/config/genesis.json}"
GENESIS_DEST="/opt/fluentum/config/genesis.json"

if [ ! -f "$GENESIS_SRC" ]; then
  echo "[ERROR] Source genesis file not found: $GENESIS_SRC"
  exit 1
fi

for NODE in "${NODES[@]}"; do
  echo "\n[INFO] Copying $GENESIS_SRC to $USER@$NODE:$GENESIS_DEST ..."
  scp "$GENESIS_SRC" "$USER@$NODE:~/genesis.json"
  if [ $? -eq 0 ]; then
    echo "[INFO] Moving genesis.json to $GENESIS_DEST on $NODE ..."
    ssh "$USER@$NODE" "sudo mv ~/genesis.json $GENESIS_DEST && sudo chown root:root $GENESIS_DEST && sudo chmod 644 $GENESIS_DEST"
    if [ $? -eq 0 ]; then
      echo "[SUCCESS] Copied and moved genesis.json on $NODE"
    else
      echo "[ERROR] Failed to move genesis.json to $GENESIS_DEST on $NODE"
    fi
  else
    echo "[ERROR] scp failed for $NODE"
  fi
  echo "--------------------------------------"
done

echo "\n[INFO] Genesis file copied to all nodes." 