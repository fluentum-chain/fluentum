#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# 1. Remove existing go.mod and go.sum
rm -f go.mod go.sum

echo "[1/7] Removed go.mod and go.sum."

# 2. Re-initialize your Go module
MODULE_NAME="github.com/fluentum-chain/fluentum"
go mod init "$MODULE_NAME"
echo "[2/7] Initialized new Go module: $MODULE_NAME."

# 3. Add replace directives for Tendermint/CometBFT and Cosmos SDK
REPLACES=(
  "github.com/tendermint/tendermint=github.com/cometbft/cometbft@v0.38.6"
  "github.com/tendermint/tendermint-db=github.com/cometbft/cometbft-db@v1.0.4"
)
for rep in "${REPLACES[@]}"; do
  go mod edit -replace "$rep"
done
echo "[3/7] Added replace directives."

# 4. Add the Cosmos SDK and CometBFT dependencies
go get github.com/cosmos/cosmos-sdk@v0.50.6
go get github.com/cometbft/cometbft@v0.38.6
echo "[4/7] Added Cosmos SDK and CometBFT dependencies."

# 5. Update all import paths from cosmossdk.io to github.com/cosmos/cosmos-sdk
find . -name "*.go" -exec sed -i 's|cosmossdk.io|github.com/cosmos/cosmos-sdk|g' {} \;
echo "[5/7] Updated all import paths from cosmossdk.io to github.com/cosmos/cosmos-sdk."

# 6. Tidy up your modules
go mod tidy
echo "[6/7] Ran go mod tidy."

# 7. Verify your dependencies
go list -m all | grep -E 'cosmos|cometbft|tendermint'

echo "[7/7] Dependency check complete."

echo "\nScript complete. If you see no errors above, your project should now use the correct Cosmos SDK imports and dependencies."

# 8. Double-check for any remaining cosmossdk.io imports
grep -r "cosmossdk.io" . --include="*.go" || echo "No cosmossdk.io imports remain." 