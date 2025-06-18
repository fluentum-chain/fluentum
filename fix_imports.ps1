# PowerShell script to fix external Tendermint imports
# This script replaces all github.com/tendermint/tendermint imports with github.com/fluentum-chain/fluentum

Write-Host "Fixing external Tendermint imports..."

# Get all Go files
$goFiles = Get-ChildItem -Path . -Filter "*.go" -Recurse

$totalFiles = $goFiles.Count
$processedFiles = 0

foreach ($file in $goFiles) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Replace external Tendermint imports with local Fluentum imports
    $content = $content -replace 'github\.com/tendermint/tendermint/', 'github.com/fluentum-chain/fluentum/'
    
    # Only write if content changed
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        Write-Host "Fixed imports in: $($file.FullName)"
        $processedFiles++
    }
}

Write-Host "Processed $processedFiles out of $totalFiles files"
Write-Host "Import fix completed!" 