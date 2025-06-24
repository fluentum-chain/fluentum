#!/bin/bash

# Fix proto import paths
# This script fixes import paths in proto files to use relative paths

set -e

# Change to the project root directory
cd "$(dirname "$0")/.."

echo "Fixing proto import paths..."

# Function to fix imports in a proto file
fix_imports() {
    local file="$1"
    if [ -f "$file" ]; then
        echo "Fixing imports in $file..."
        
        # Replace tendermint/ imports with relative imports
        sed -i 's|import "tendermint/|import "|g' "$file"
        
        # Replace tendermint. type references with relative references
        sed -i 's|tendermint\.types\.|types\.|g' "$file"
        sed -i 's|tendermint\.crypto\.|crypto\.|g' "$file"
        sed -i 's|tendermint\.p2p\.|p2p\.|g' "$file"
        sed -i 's|tendermint\.state\.|state\.|g' "$file"
        sed -i 's|tendermint\.version\.|version\.|g' "$file"
        sed -i 's|tendermint\.consensus\.|consensus\.|g' "$file"
        sed -i 's|tendermint\.privval\.|privval\.|g' "$file"
        sed -i 's|tendermint\.blockchain\.|blockchain\.|g' "$file"
        sed -i 's|tendermint\.libs\.|libs\.|g' "$file"
        sed -i 's|tendermint\.mempool\.|mempool\.|g' "$file"
        sed -i 's|tendermint\.store\.|store\.|g' "$file"
        sed -i 's|tendermint\.statesync\.|statesync\.|g' "$file"
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
echo "2. Or run: ./scripts/generate_proto_buf.sh" 