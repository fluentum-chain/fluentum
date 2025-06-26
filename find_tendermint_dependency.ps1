# Find which modules require github.com/tendermint/tendermint

Write-Host "Searching go mod graph for dependencies on github.com/tendermint/tendermint..."

go mod graph | Select-String "github.com/tendermint/tendermint" | ForEach-Object { Write-Host $_ }

Write-Host "Tendermint dependency chain search complete." 