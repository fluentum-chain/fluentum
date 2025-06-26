# Force remove any line containing the legacy btcec replace directive from go.mod, then tidy modules

Write-Host "Force removing any legacy btcec replace directive from go.mod..."
(Get-Content go.mod) | Where-Object { $_ -notmatch 'replace github.com/btcsuite/btcd/btcec' } | Set-Content go.mod

Write-Host "Tidying up modules..."
go mod tidy

Write-Host "All legacy btcec replace directives forcibly removed and modules tidied." 