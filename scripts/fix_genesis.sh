#!/bin/bash

# Quick Fix for Genesis File
# This script fixes the genesis file format issue

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "=========================================="
echo "    Genesis File Fix Script"
echo "=========================================="
echo ""

# Check if genesis file exists
GENESIS_FILE="/opt/fluentum/config/genesis.json"

if [ ! -f "$GENESIS_FILE" ]; then
    print_error "Genesis file not found at $GENESIS_FILE"
    exit 1
fi

print_status "Found genesis file: $GENESIS_FILE"

# Create backup
BACKUP_FILE="$GENESIS_FILE.backup.$(date +%Y%m%d_%H%M%S)"
cp "$GENESIS_FILE" "$BACKUP_FILE"
print_success "Created backup: $BACKUP_FILE"

# Fix the genesis file
print_status "Fixing genesis file..."

# Use sed to fix all numeric fields (remove quotes)
sed -i 's/"max_age_duration": "172800000000000"/"max_age_duration": 172800000000000/' "$GENESIS_FILE"
sed -i 's/"max_bytes": "22020096"/"max_bytes": 22020096/' "$GENESIS_FILE"
sed -i 's/"max_gas": "-1"/"max_gas": -1/' "$GENESIS_FILE"
sed -i 's/"time_iota_ms": "1000"/"time_iota_ms": 1000/' "$GENESIS_FILE"
sed -i 's/"max_age_num_blocks": "100000"/"max_age_num_blocks": 100000/' "$GENESIS_FILE"
sed -i 's/"max_bytes": "1048576"/"max_bytes": 1048576/' "$GENESIS_FILE"
sed -i 's/"initial_height": "1"/"initial_height": 1/' "$GENESIS_FILE"

# Also fix any other potential numeric fields that might be strings
sed -i 's/"max_bytes": "\([0-9]*\)"/"max_bytes": \1/g' "$GENESIS_FILE"
sed -i 's/"max_gas": "\(-?[0-9]*\)"/"max_gas": \1/g' "$GENESIS_FILE"
sed -i 's/"time_iota_ms": "\([0-9]*\)"/"time_iota_ms": \1/g' "$GENESIS_FILE"
sed -i 's/"max_age_num_blocks": "\([0-9]*\)"/"max_age_num_blocks": \1/g' "$GENESIS_FILE"
sed -i 's/"initial_height": "\([0-9]*\)"/"initial_height": \1/g' "$GENESIS_FILE"

print_success "Genesis file fixed!"

# Validate the JSON
if jq empty "$GENESIS_FILE" 2>/dev/null; then
    print_success "Genesis file is valid JSON"
else
    print_error "Genesis file is not valid JSON"
    exit 1
fi

# Show the fixed content
print_status "Fixed genesis file content:"
echo "----------------------------------------"
jq '.consensus_params.evidence' "$GENESIS_FILE"
echo "----------------------------------------"

# Stop the service if it's running
if systemctl is-active --quiet fluentum-testnet.service; then
    print_status "Stopping fluentum-testnet service..."
    sudo systemctl stop fluentum-testnet.service
fi

# Start the service
print_status "Starting fluentum-testnet service..."
sudo systemctl start fluentum-testnet.service

# Wait a moment and check status
sleep 3
if systemctl is-active --quiet fluentum-testnet.service; then
    print_success "Service started successfully!"
    echo ""
    print_status "Service status:"
    sudo systemctl status fluentum-testnet.service --no-pager
else
    print_error "Service failed to start. Checking logs..."
    sudo journalctl -u fluentum-testnet.service --no-pager -n 10
fi

echo ""
print_success "Genesis file fix completed!"
echo ""
echo "If the service is now running, you can test it with:"
echo "  curl http://localhost:26657/status"
echo ""
echo "To view logs:"
echo "  sudo journalctl -u fluentum-testnet.service -f" 