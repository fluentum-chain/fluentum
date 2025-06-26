# Search the Go module cache for legacy btcec imports

Write-Host "Searching Go module cache for legacy 'github.com/btcsuite/btcd/btcec' imports..."

$modCache = if ($env:GOPATH) { Join-Path $env:GOPATH 'pkg\mod' } else { Join-Path $env:USERPROFILE 'go\pkg\mod' }

Get-ChildItem -Recurse -Filter *.go $modCache | ForEach-Object {
    $matches = Select-String -Path $_.FullName -Pattern 'github.com/btcsuite/btcd/btcec"' -SimpleMatch
    if ($matches) {
        $matches | ForEach-Object { Write-Host ("$($_.Path):$($_.LineNumber): $($_.Line)") }
    }
}

Write-Host "Module cache search complete." 