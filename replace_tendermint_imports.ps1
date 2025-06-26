# Replace all github.com/tendermint/tendermint imports with github.com/cometbft/cometbft in all .go files

Write-Host "Replacing all github.com/tendermint/tendermint imports with github.com/cometbft/cometbft..."

Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/tendermint/tendermint', 'github.com/cometbft/cometbft' | Set-Content $_.FullName
}

Write-Host "Import path replacement complete." 