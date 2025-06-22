# Cleanup Dependencies Script for Fluentum
# This script helps prepare the go.mod file for go mod tidy on the server

Write-Host "=== Fluentum Dependency Cleanup Script ===" -ForegroundColor Green
Write-Host "This script will help prepare dependencies for go mod tidy" -ForegroundColor Yellow

# Backup current go.mod and go.sum
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$backupDir = "backup_$timestamp"

if (!(Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir
}

Copy-Item "go.mod" "$backupDir/go.mod.backup"
Copy-Item "go.sum" "$backupDir/go.sum.backup"

Write-Host "Backed up current files to $backupDir" -ForegroundColor Cyan

# Remove go.sum to force regeneration
if (Test-Path "go.sum") {
    Remove-Item "go.sum"
    Write-Host "Removed go.sum to force regeneration" -ForegroundColor Yellow
}

# Check for any remaining issues in go.mod
Write-Host "`nChecking go.mod for potential issues..." -ForegroundColor Cyan

$goModContent = Get-Content "go.mod" -Raw

# Check for duplicate require blocks
$requireBlocks = ([regex]::Matches($goModContent, 'require\s*\(')).Count
if ($requireBlocks -gt 1) {
    Write-Host "Warning: Found $requireBlocks require blocks in go.mod" -ForegroundColor Yellow
    Write-Host "This may cause issues. Consider merging them into a single require block." -ForegroundColor Yellow
}

# Check for replace directives that might conflict
$replaceDirectives = [regex]::Matches($goModContent, 'replace\s+.*=>.*').Count
Write-Host "Found $replaceDirectives replace directives" -ForegroundColor Cyan

# Check for specific problematic patterns
if ($goModContent -match 'github\.com/cometbft/cometbft') {
    Write-Host "Warning: Found CometBFT references in go.mod" -ForegroundColor Yellow
    Write-Host "These should be replaced with Tendermint v0.35.9 for compatibility" -ForegroundColor Yellow
}

# Verify Cosmos SDK version
if ($goModContent -match 'github\.com/cosmos/cosmos-sdk v0\.47\.5') {
    Write-Host "✓ Cosmos SDK v0.47.5 found - compatible with Tendermint v0.35.9" -ForegroundColor Green
} else {
    Write-Host "Warning: Cosmos SDK version may not be compatible" -ForegroundColor Yellow
}

Write-Host "`n=== Next Steps ===" -ForegroundColor Green
Write-Host "1. Run 'go mod tidy' on the server to resolve dependencies" -ForegroundColor White
Write-Host "2. If issues persist, check the backup files in $backupDir" -ForegroundColor White
Write-Host "3. You may need to manually adjust replace directives" -ForegroundColor White

Write-Host "`n=== Expected Issues and Solutions ===" -ForegroundColor Green
Write-Host "If go mod tidy fails with CometBFT errors:" -ForegroundColor Yellow
Write-Host "  - Ensure replace directive: github.com/cometbft/cometbft => github.com/tendermint/tendermint v0.35.9" -ForegroundColor White
Write-Host "  - Verify Cosmos SDK v0.47.5 is used (not v0.50.0)" -ForegroundColor White

Write-Host "`nIf go mod tidy fails with missing packages:" -ForegroundColor Yellow
Write-Host "  - Check that all imports in cmd/fluentum/main.go use correct package paths" -ForegroundColor White
Write-Host "  - Ensure fluentum/app/encoding.go has correct imports" -ForegroundColor White

Write-Host "`n=== Cleanup Complete ===" -ForegroundColor Green
Write-Host "Ready to run 'go mod tidy' on the server" -ForegroundColor Cyan 