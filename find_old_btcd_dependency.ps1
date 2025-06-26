# Find which modules require github.com/btcsuite/btcd@v0.20.1-beta

Write-Host "Searching go mod graph for dependencies on github.com/btcsuite/btcd@v0.20.1-beta..."

go mod graph | Select-String "github.com/btcsuite/btcd@v0.20.1-beta" | ForEach-Object { Write-Host $_ }

Write-Host "Old btcd dependency chain search complete." 