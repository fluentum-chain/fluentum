$lines = Get-Content go.sum
$seen = @{}
$unique = @()
foreach ($line in $lines) {
    if (-not $seen.ContainsKey($line)) {
        $unique += $line
        $seen[$line] = $true
    }
}
$unique | Set-Content go.sum 