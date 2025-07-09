#!/bin/bash

# List of node hostnames or IPs (edit as needed)
NODES=("node1" "node2" "node3" "node4" "node5")
# Path to your Fluentum repo on each node
REMOTE_PATH="/opt/fluentum"
# Name of the systemd service (edit if different)
SERVICE_NAME="fluentumd"
# Path to the built binary (relative to REMOTE_PATH)
BINARY_PATH="fluentumd"

# Optional: Path to your SSH key
# SSH_KEY="~/.ssh/id_rsa"

echo "Building fluentumd binary locally..."
cd "$REMOTE_PATH" || { echo "Could not cd to $REMOTE_PATH"; exit 1; }
go build -o "$BINARY_PATH" ./cmd/fluentum || { echo "Build failed!"; exit 1; }
echo "Build successful."

for NODE in "${NODES[@]}"; do
    echo "----"
    echo "Deploying to $NODE..."

    # Copy the new binary to the node
    scp "$REMOTE_PATH/$BINARY_PATH" "$NODE:$REMOTE_PATH/$BINARY_PATH" || { echo "SCP failed for $NODE"; continue; }

    # Restart the service on the node
    ssh "$NODE" "sudo systemctl restart $SERVICE_NAME" || { echo "Failed to restart $SERVICE_NAME on $NODE"; continue; }

    # Optional: Check status and print last 10 log lines
    ssh "$NODE" "sudo systemctl status $SERVICE_NAME --no-pager"
    ssh "$NODE" "sudo journalctl -u $SERVICE_NAME -n 10 --no-pager"

    echo "$NODE updated and restarted."
done

echo "All nodes processed."