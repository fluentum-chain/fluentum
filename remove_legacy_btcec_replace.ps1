# Remove the replace directive for legacy btcec from go.mod, then tidy modules

Write-Host "Removing replace directive for github.com/btcsuite/btcd/btcec..."
go mod edit -dropreplace github.com/btcsuite/btcd/btcec

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "Legacy btcec replace directive removed and modules tidied." 