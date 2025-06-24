#!/bin/bash

# Fix proto import paths - Final version
# This script fixes import paths in proto files to work with buf

set -e

# Change to the project root directory
cd "$(dirname "$0")/.."

echo "Fixing proto import paths for buf compatibility..."

# Function to fix imports in a proto file
fix_imports() {
    local file="$1"
    if [ -f "$file" ]; then
        echo "Fixing imports in $file..."
        
        # Replace relative imports with absolute tendermint imports
        sed -i 's|import "crypto/|import "tendermint/crypto/|g' "$file"
        sed -i 's|import "types/|import "tendermint/types/|g' "$file"
        sed -i 's|import "p2p/|import "tendermint/p2p/|g' "$file"
        sed -i 's|import "state/|import "tendermint/state/|g' "$file"
        sed -i 's|import "version/|import "tendermint/version/|g' "$file"
        sed -i 's|import "consensus/|import "tendermint/consensus/|g' "$file"
        sed -i 's|import "privval/|import "tendermint/privval/|g' "$file"
        sed -i 's|import "blockchain/|import "tendermint/blockchain/|g' "$file"
        sed -i 's|import "libs/|import "tendermint/libs/|g' "$file"
        sed -i 's|import "mempool/|import "tendermint/mempool/|g' "$file"
        sed -i 's|import "store/|import "tendermint/store/|g' "$file"
        sed -i 's|import "statesync/|import "tendermint/statesync/|g' "$file"
        
        # Replace relative type references with absolute tendermint references
        sed -i 's|types\.|tendermint.types.|g' "$file"
        sed -i 's|crypto\.|tendermint.crypto.|g' "$file"
        sed -i 's|p2p\.|tendermint.p2p.|g' "$file"
        sed -i 's|state\.|tendermint.state.|g' "$file"
        sed -i 's|version\.|tendermint.version.|g' "$file"
        sed -i 's|consensus\.|tendermint.consensus.|g' "$file"
        sed -i 's|privval\.|tendermint.privval.|g' "$file"
        sed -i 's|blockchain\.|tendermint.blockchain.|g' "$file"
        sed -i 's|libs\.|tendermint.libs.|g' "$file"
        sed -i 's|mempool\.|tendermint.mempool.|g' "$file"
        sed -i 's|store\.|tendermint.store.|g' "$file"
        sed -i 's|statesync\.|tendermint.statesync.|g' "$file"
    fi
}

# Find all proto files and fix their imports
find proto/tendermint -name "*.proto" -type f | while read -r file; do
    fix_imports "$file"
done

echo "Proto import paths fixed successfully!"
echo ""
echo "Next steps:"
echo "1. Run: cd proto && buf generate"
echo "2. Or run: make proto-gen" 