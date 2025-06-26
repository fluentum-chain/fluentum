# Add replace directives to fix ambiguous and legacy module paths

Write-Host "Adding replace directive for github.com/btcsuite/btcd/chaincfg/chainhash..."
go mod edit -replace github.com/btcsuite/btcd/chaincfg/chainhash=github.com/btcsuite/btcd@v0.20.1-beta

Write-Host "Adding replace directive for github.com/tendermint/tendermint..."
go mod edit -replace github.com/tendermint/tendermint=github.com/cometbft/cometbft@v0.38.6

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "Replace directives applied and modules tidied." 