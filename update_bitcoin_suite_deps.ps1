# Update all btcsuite dependencies to their latest versions, then tidy modules

Write-Host "Updating btcsuite dependencies to latest versions..."
go get github.com/btcsuite/btcd@latest
go get github.com/btcsuite/btcutil@latest
go get github.com/btcsuite/btclog@latest
go get github.com/btcsuite/btcutil/bloom@latest
go get github.com/btcsuite/btcd/chaincfg/chainhash@latest

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "btcsuite dependencies updated and modules tidied." 