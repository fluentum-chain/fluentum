# Search all go.mod files for references to v0.20.1-beta

Write-Host "Searching all go.mod files for references to v0.20.1-beta..."

Get-ChildItem -Recurse -Filter go.mod | ForEach-Object {
    $matches = Select-String -Path $_.FullName -Pattern 'v0.20.1-beta' -SimpleMatch
    if ($matches) {
        $matches | ForEach-Object { Write-Host ("$($_.Path):$($_.LineNumber): $($_.Line)") }
    }
}

Write-Host "Search for v0.20.1-beta in go.mod files complete." 