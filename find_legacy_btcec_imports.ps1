# Search for all legacy btcec imports in the codebase

Write-Host "Searching for legacy 'github.com/btcsuite/btcd/btcec' imports in .go files..."

Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
    $matches = Select-String -Path $_.FullName -Pattern 'github.com/btcsuite/btcd/btcec"' -SimpleMatch
    if ($matches) {
        $matches | ForEach-Object { Write-Host ("$($_.Path):$($_.LineNumber): $($_.Line)") }
    }
}

Write-Host "Search complete." 