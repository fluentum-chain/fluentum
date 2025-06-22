# Alternative approach: Fix compilation errors using replace directives

Write-Host "=== Alternative Fix: Using Replace Directives ===" -ForegroundColor Cyan
Write-Host ""

# Check if we need to add replace directives
$goModContent = Get-Content go.mod -Raw
$hasReplace = $goModContent -match "replace"

if (-not $hasReplace) {
    Write-Host "Adding replace section to go.mod..." -ForegroundColor Yellow
    $goModContent = $goModContent + "`n`nreplace ("
} else {
    Write-Host "Replace section already exists, adding new directives..." -ForegroundColor Yellow
    $goModContent = $goModContent -replace "replace \(", "replace (`n"
}

# Add replace directives to fix the compilation errors
$replaceDirectives = @"
	github.com/cometbft/cometbft => github.com/tendermint/tendermint v0.35.9
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.2
"@

$goModContent = $goModContent -replace "\)", "$replaceDirectives`n)"

# Write the updated go.mod
$goModContent | Set-Content go.mod -NoNewline

Write-Host "Added replace directives to fix compilation errors" -ForegroundColor Green
Write-Host ""

# Clean and redownload modules
Write-Host "Cleaning module cache and redownloading..." -ForegroundColor Yellow
go clean -modcache
go mod download

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host "Fixed compilation errors using replace directives" -ForegroundColor Green
Write-Host "You can now try running the installation again" -ForegroundColor White
Write-Host ""
Write-Host "The replace directives will:" -ForegroundColor Yellow
Write-Host "1. Use Tendermint v0.35.9 instead of CometBFT" -ForegroundColor White
Write-Host "2. Use protobuf v1.5.2 instead of v1.5.3" -ForegroundColor White 