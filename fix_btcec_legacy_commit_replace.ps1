# Update the replace directive for legacy btcec to the latest commit hash, then tidy modules

Write-Host "Updating replace directive for github.com/btcsuite/btcd/btcec to commit b3f1a3a..."
go mod edit -replace github.com/btcsuite/btcd/btcec=github.com/btcsuite/btcd/btcec@b3f1a3a

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "btcec (legacy) replace directive updated to commit and modules tidied." 