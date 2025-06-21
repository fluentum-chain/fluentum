# PowerShell script to clean up go.sum file by removing incomplete entries

$goSumPath = "go.sum"
$goModPath = "go.mod"
$backupPath = "go.sum.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"

Write-Host "=== Go.sum Cleanup Script ===" -ForegroundColor Cyan
Write-Host ""

# Check if files exist
if (!(Test-Path $goSumPath)) {
    Write-Host "Error: go.sum not found in current directory." -ForegroundColor Red
    exit 1
}
if (!(Test-Path $goModPath)) {
    Write-Host "Error: go.mod not found in current directory." -ForegroundColor Red
    exit 1
}

# Create backup
Write-Host "Creating backup of go.sum..." -ForegroundColor Yellow
Copy-Item $goSumPath $backupPath
Write-Host "Backup created: $backupPath" -ForegroundColor Green

# Read current go.sum
$lines = Get-Content $goSumPath
$originalCount = $lines.Count
Write-Host "Original go.sum has $originalCount lines" -ForegroundColor White

# Analyze entries to find incomplete pairs
$moduleVersions = @{ }
$incompleteEntries = @()

foreach ($line in $lines) {
    $parts = $line -split '\s+'
    if ($parts.Count -eq 3) {
        $mod = $parts[0]
        $ver = $parts[1]
        $key = if ($ver -like "*/go.mod") { "$mod $($ver -replace '/go.mod', '')" } else { "$mod $ver" }
        
        if (-not $moduleVersions.ContainsKey($key)) {
            $moduleVersions[$key] = @{ zip = $false; mod = $false; lines = @() }
        }
        
        if ($ver -like "*/go.mod") {
            $moduleVersions[$key].mod = $true
        } else {
            $moduleVersions[$key].zip = $true
        }
        
        $moduleVersions[$key].lines += $line
    }
}

# Find incomplete entries
foreach ($entry in $moduleVersions.GetEnumerator()) {
    if (-not ($entry.Value.zip -and $entry.Value.mod)) {
        $incompleteEntries += $entry.Value.lines
    }
}

Write-Host "Found $($incompleteEntries.Count) incomplete entries" -ForegroundColor Yellow

if ($incompleteEntries.Count -gt 0) {
    Write-Host ""
    Write-Host "Incomplete entries to be removed:" -ForegroundColor Red
    foreach ($entry in $incompleteEntries) {
        Write-Host "  $entry" -ForegroundColor Red
    }
    
    # Remove incomplete entries
    $cleanLines = $lines | Where-Object { $incompleteEntries -notcontains $_ }
    
    # Write cleaned go.sum
    $cleanLines | Set-Content $goSumPath
    Write-Host ""
    Write-Host "Cleaned go.sum written with $($cleanLines.Count) lines" -ForegroundColor Green
    Write-Host "Removed $($originalCount - $cleanLines.Count) incomplete entries" -ForegroundColor Green
} else {
    Write-Host "No incomplete entries found. go.sum is already clean." -ForegroundColor Green
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host "Original lines: $originalCount" -ForegroundColor White
Write-Host "Incomplete entries: $($incompleteEntries.Count)" -ForegroundColor Yellow
Write-Host "Clean lines: $($cleanLines.Count)" -ForegroundColor Green
Write-Host "Backup file: $backupPath" -ForegroundColor White

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Run 'go mod tidy' to regenerate missing entries" -ForegroundColor White
Write-Host "2. Verify the build works correctly" -ForegroundColor White
Write-Host "3. If everything works, you can delete the backup file" -ForegroundColor White
Write-Host "4. If there are issues, restore from backup: Copy-Item '$backupPath' 'go.sum'" -ForegroundColor White

Write-Host ""
Write-Host "Would you like to run 'go mod tidy' now? (y/n)" -ForegroundColor Yellow
$response = Read-Host

if ($response -eq 'y' -or $response -eq 'Y') {
    Write-Host "Running 'go mod tidy'..." -ForegroundColor Yellow
    try {
        go mod tidy
        Write-Host "go mod tidy completed successfully!" -ForegroundColor Green
    } catch {
        Write-Host "Error running go mod tidy: $_" -ForegroundColor Red
        Write-Host "You can restore the backup with: Copy-Item '$backupPath' 'go.sum'" -ForegroundColor Yellow
    }
} else {
    Write-Host "Skipping go mod tidy. Run it manually when ready." -ForegroundColor Yellow
} 