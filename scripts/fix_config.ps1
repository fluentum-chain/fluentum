# Fix config file line endings for Linux server
# This script converts Windows line endings to Unix line endings

Write-Host "Fixing config file line endings..." -ForegroundColor Green

# Read the config file with Windows line endings
$configContent = Get-Content "config/config.toml" -Raw

# Convert to Unix line endings
$configContent = $configContent -replace "`r`n", "`n"

# Write back with Unix line endings
[System.IO.File]::WriteAllText("config/config.toml", $configContent, [System.Text.Encoding]::UTF8)

Write-Host "Config file line endings fixed!" -ForegroundColor Green 