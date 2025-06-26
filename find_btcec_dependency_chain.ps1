# Find which modules require legacy github.com/btcsuite/btcd/btcec

Write-Host "Searching go mod graph for legacy btcec dependencies..."

go mod graph | Select-String "github.com/btcsuite/btcd/btcec" | ForEach-Object { Write-Host $_ }

Write-Host "Dependency chain search complete." 