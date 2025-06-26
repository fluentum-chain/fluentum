# Find which modules require legacy github.com/btcsuite/btcd/btcec (not /v2)

Write-Host "Searching go mod graph for legacy btcec (not /v2) dependencies..."

go mod graph | Select-String "github.com/btcsuite/btcd/btcec " | ForEach-Object { Write-Host $_ }

Write-Host "Legacy btcec dependency chain search complete." 