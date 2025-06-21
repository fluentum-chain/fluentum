$content = Get-Content go.sum
$newContent = @()

foreach ($line in $content) {
    if (-not ($line -match "cosmossdk\.io/snapshots")) {
        $newContent += $line
    }
}

$newContent | Set-Content go.sum 