# Remove go.mod and go.sum if they exist
Remove-Item -Force go.mod, go.sum -ErrorAction SilentlyContinue

Write-Host "[1/7] Removed go.mod and go.sum."

# Re-initialize Go module
$MODULE_NAME = "github.com/fluentum-chain/fluentum"
go mod init $MODULE_NAME
Write-Host "[2/7] Initialized new Go module: $MODULE_NAME."

# Add replace directives for Tendermint/CometBFT and Cosmos SDK
go mod edit -replace "github.com/tendermint/tendermint=github.com/cometbft/cometbft@v0.38.6"
go mod edit -replace "github.com/tendermint/tendermint-db=github.com/cometbft/cometbft-db@v1.0.4"
Write-Host "[3/7] Added replace directives."

# Add the Cosmos SDK and CometBFT dependencies
go get github.com/cosmos/cosmos-sdk@v0.50.6
go get github.com/cometbft/cometbft@v0.38.6
Write-Host "[4/7] Added Cosmos SDK and CometBFT dependencies."

# Update all import paths from cosmossdk.io to github.com/cosmos/cosmos-sdk
Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'cosmossdk\.io', 'github.com/cosmos/cosmos-sdk' | Set-Content $_.FullName
}
Write-Host "[5/7] Updated all import paths from cosmossdk.io to github.com/cosmos/cosmos-sdk."

# Tidy up your modules
go mod tidy
Write-Host "[6/7] Ran go mod tidy."

# Verify your dependencies
go list -m all | Select-String -Pattern 'cosmos|cometbft|tendermint'
Write-Host "[7/7] Dependency check complete."

# Double-check for any remaining cosmossdk.io imports
$remaining = Get-ChildItem -Recurse -Filter *.go | Select-String "cosmossdk.io"
if ($remaining) {
    Write-Host "Some cosmossdk.io imports remain:"
    $remaining | ForEach-Object { Write-Host $_ }
} else {
    Write-Host "No cosmossdk.io imports remain."
} 