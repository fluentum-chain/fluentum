#!/bin/bash

echo "=== Debugging Genesis File and State ==="

# Check if genesis file exists
if [ -f "/opt/fluentum/config/genesis.json" ]; then
    echo "✓ Genesis file exists"
    echo "=== Genesis File Content ==="
    cat /opt/fluentum/config/genesis.json | head -30
    echo "=== End Genesis File ==="
else
    echo "✗ Genesis file does not exist"
fi

# Check for any persistent state files
echo ""
echo "=== Checking for Persistent State ==="
if [ -d "/opt/fluentum/data" ]; then
    echo "Data directory contents:"
    ls -la /opt/fluentum/data/
    
    # Check for state.db or similar files
    if [ -f "/opt/fluentum/data/state.db" ]; then
        echo "⚠️  Found state.db - this might contain old state"
    fi
    
    if [ -f "/opt/fluentum/data/application.db" ]; then
        echo "⚠️  Found application.db - this might contain old state"
    fi
else
    echo "Data directory does not exist"
fi

# Check for any Tendermint state files
echo ""
echo "=== Checking Tendermint State ==="
if [ -f "/opt/fluentum/data/priv_validator_state.json" ]; then
    echo "Found priv_validator_state.json:"
    cat /opt/fluentum/data/priv_validator_state.json
fi

# Check for any other potential state files
echo ""
echo "=== Checking for Other State Files ==="
find /opt/fluentum -name "*.db" -o -name "*.state" -o -name "state.json" 2>/dev/null

echo ""
echo "=== Debug Complete ===" 