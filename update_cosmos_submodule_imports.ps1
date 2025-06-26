# Update all Cosmos SDK submodule imports to new cosmossdk.io paths

Write-Host "Updating github.com/cosmos/cosmos-sdk/core/store to cosmossdk.io/core/store..."
Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/cosmos/cosmos-sdk/core/store', 'cosmossdk.io/core/store' | Set-Content $_.FullName
}

Write-Host "Updating github.com/cosmos/cosmos-sdk/core to cosmossdk.io/core..."
Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/cosmos/cosmos-sdk/core', 'cosmossdk.io/core' | Set-Content $_.FullName
}

Write-Host "Updating github.com/cosmos/cosmos-sdk/log to cosmossdk.io/log..."
Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/cosmos/cosmos-sdk/log', 'cosmossdk.io/log' | Set-Content $_.FullName
}

Write-Host "Updating github.com/cosmos/cosmos-sdk/store/types to cosmossdk.io/store/types..."
Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    (Get-Content $_.FullName) -replace 'github.com/cosmos/cosmos-sdk/store/types', 'cosmossdk.io/store/types' | Set-Content $_.FullName
}

Write-Host "Cosmos SDK submodule import path updates complete." 