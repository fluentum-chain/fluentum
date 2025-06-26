# Update the replace directive for legacy btcec to the correct version v1.0.4, then tidy modules

Write-Host "Updating replace directive for github.com/btcsuite/btcd/btcec to v1.0.4..."
go mod edit -replace github.com/btcsuite/btcd/btcec=github.com/btcsuite/btcd/btcec@v1.0.4

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "btcec (legacy) replace directive updated and modules tidied." 