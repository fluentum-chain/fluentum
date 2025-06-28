# PowerShell script to fix proto import paths
# Replace external proto imports with local ones

$files = Get-ChildItem -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*vendor*" }

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Replace external proto imports with local ones
    $content = $content -replace 'github\.com/fluentum-chain/fluentum/proto/tendermint', 'github.com/fluentum-chain/fluentum/proto/tendermint'
    
    # Only write if content changed
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        Write-Host "Fixed imports in: $($file.FullName)"
    }
}

Write-Host "Import fixes completed!" 