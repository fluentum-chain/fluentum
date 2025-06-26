# Add replace directives for btcec/v2 and legacy btcec, then tidy modules

Write-Host "Adding replace directive for github.com/btcsuite/btcd/btcec/v2..."
go mod edit -replace github.com/btcsuite/btcd/btcec/v2=github.com/btcsuite/btcd/btcec/v2@v2.0.1

Write-Host "Adding replace directive for github.com/btcsuite/btcd/btcec (legacy)..."
go mod edit -replace github.com/btcsuite/btcd/btcec=github.com/btcsuite/btcd/btcec@v1.0.3

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "btcec replace directives applied and modules tidied." 