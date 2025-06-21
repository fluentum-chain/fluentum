#!/bin/bash

# Fix Permissions Script for Fluentum Core
# This script makes the build scripts executable

set -e

echo "=========================================="
echo "    Fixing Script Permissions"
echo "=========================================="
echo ""

# Make build scripts executable
echo "Making build scripts executable..."
chmod +x scripts/build.sh
chmod +x scripts/fix_permissions.sh

# Verify permissions
echo "Verifying permissions..."
ls -la scripts/build.sh
ls -la scripts/fix_permissions.sh

echo ""
echo "✅ Permissions fixed successfully!"
echo ""
echo "You can now run:"
echo "  ./scripts/build.sh"
echo "  make build"
echo "" 