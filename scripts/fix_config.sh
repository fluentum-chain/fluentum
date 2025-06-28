#!/bin/bash

# Fix config file line endings for Linux server
# This script converts Windows line endings to Unix line endings

echo "Fixing config file line endings..."

# Check if config file exists
if [ ! -f "config/config.toml" ]; then
    echo "Error: config/config.toml not found"
    exit 1
fi

# Convert Windows line endings to Unix line endings
dos2unix config/config.toml 2>/dev/null || sed -i 's/\r$//' config/config.toml

echo "Config file line endings fixed!" 