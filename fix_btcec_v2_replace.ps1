# Update the replace directive for btcec/v2 to the correct version v2.3.2, then tidy modules

Write-Host "Updating replace directive for github.com/btcsuite/btcd/btcec/v2 to v2.3.2..."
go mod edit -replace github.com/btcsuite/btcd/btcec/v2=github.com/btcsuite/btcd/btcec/v2@v2.3.2

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "btcec/v2 replace directive updated and modules tidied." 