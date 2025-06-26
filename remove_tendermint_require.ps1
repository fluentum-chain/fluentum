# Remove the require directive for github.com/tendermint/tendermint from go.mod, then tidy modules

Write-Host "Removing require directive for github.com/tendermint/tendermint..."
go mod edit -droprequire github.com/tendermint/tendermint

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "Tendermint require directive removed and modules tidied." 