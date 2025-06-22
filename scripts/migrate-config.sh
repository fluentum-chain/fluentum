#!/bin/bash

# Migration script for Fluentum blockchain from Tendermint to CometBFT
# This script uses confix to migrate configuration files

set -e

echo "Starting Fluentum blockchain migration from Tendermint to CometBFT..."

# Check if confix is installed
if ! command -v confix &> /dev/null; then
    echo "Installing confix..."
    go install github.com/cometbft/confix@latest
fi

# Set environment variables
export CMTHOME=${CMTHOME:-"$HOME/.cometbft"}
export TMHOME=${TMHOME:-"$HOME/.tendermint"}

echo "Using CMTHOME: $CMTHOME"
echo "Using TMHOME: $TMHOME"

# Create CometBFT directory if it doesn't exist
if [ ! -d "$CMTHOME" ]; then
    echo "Creating CometBFT home directory..."
    mkdir -p "$CMTHOME"
fi

# Migrate configuration if Tendermint config exists
if [ -d "$TMHOME" ]; then
    echo "Migrating configuration from Tendermint to CometBFT..."
    
    # Copy existing config files
    cp -r "$TMHOME"/* "$CMTHOME/" 2>/dev/null || true
    
    # Use confix to migrate the configuration
    echo "Running confix migration..."
    confix migrate --home "$CMTHOME" --target-version v0.38.6
    
    echo "Configuration migration completed!"
else
    echo "No existing Tendermint configuration found. Creating new CometBFT configuration..."
    
    # Initialize new CometBFT configuration
    fluentumd init --home "$CMTHOME"
fi

# Update environment variables in shell profile
SHELL_PROFILE=""
if [ -f "$HOME/.bashrc" ]; then
    SHELL_PROFILE="$HOME/.bashrc"
elif [ -f "$HOME/.zshrc" ]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [ -f "$HOME/.profile" ]; then
    SHELL_PROFILE="$HOME/.profile"
fi

if [ -n "$SHELL_PROFILE" ]; then
    echo "Updating shell profile: $SHELL_PROFILE"
    
    # Remove old TMHOME export if it exists
    sed -i '/export TMHOME/d' "$SHELL_PROFILE" 2>/dev/null || true
    
    # Add new CMTHOME export
    if ! grep -q "export CMTHOME" "$SHELL_PROFILE"; then
        echo "" >> "$SHELL_PROFILE"
        echo "# CometBFT home directory" >> "$SHELL_PROFILE"
        echo "export CMTHOME=\"$CMTHOME\"" >> "$SHELL_PROFILE"
    fi
fi

echo "Migration completed successfully!"
echo ""
echo "Next steps:"
echo "1. Run 'go mod tidy' to resolve dependencies"
echo "2. Build the application: 'make build'"
echo "3. Start the node: 'fluentumd start --home $CMTHOME'"
echo ""
echo "Note: The old Tendermint configuration is still available at $TMHOME"
echo "You can remove it after confirming everything works correctly." 