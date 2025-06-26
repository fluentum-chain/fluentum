# Remove the replace directive for github.com/btcsuite/btcd/chaincfg/chainhash from go.mod, then tidy modules

Write-Host "Removing replace directive for github.com/btcsuite/btcd/chaincfg/chainhash..."
go mod edit -dropreplace github.com/btcsuite/btcd/chaincfg/chainhash

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "chaincfg/chainhash replace directive removed and modules tidied." 