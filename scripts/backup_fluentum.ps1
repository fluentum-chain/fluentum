# Fluentum Backup Script (PowerShell)
# This script creates backups of critical Fluentum files on Windows

param(
    [string]$BackupDir = "C:\backup\fluentum",
    [string]$FluentumHome = "$env:USERPROFILE\.fluentum",
    [string]$LogFile = "C:\logs\fluentum_backup.log"
)

# Create log directory if it doesn't exist
$LogDir = Split-Path $LogFile -Parent
if (!(Test-Path $LogDir)) {
    New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
}

# Logging function
function Write-Log {
    param([string]$Message)
    $Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $LogMessage = "$Timestamp - $Message"
    Write-Host $LogMessage
    Add-Content -Path $LogFile -Value $LogMessage
}

# Error handling
function Write-ErrorAndExit {
    param([string]$Message)
    Write-Log "ERROR: $Message"
    exit 1
}

# Check if Fluentum home directory exists
if (!(Test-Path $FluentumHome)) {
    Write-ErrorAndExit "Fluentum home directory not found: $FluentumHome"
}

# Create backup directory
Write-Log "Creating backup directory: $BackupDir"
if (!(Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

$Date = Get-Date -Format "yyyyMMdd_HHmmss"
Write-Log "🔄 Creating Fluentum backup at $Date"

# Function to backup file with error handling
function Backup-File {
    param(
        [string]$Source,
        [string]$Dest,
        [string]$Description
    )
    
    if (Test-Path $Source) {
        try {
            Copy-Item -Path $Source -Destination $Dest -Force
            Write-Log "✅ Backed up $Description"
        }
        catch {
            Write-ErrorAndExit "Failed to backup $Description : $_"
        }
    }
    else {
        Write-Log "⚠️  Warning: $Description not found at $Source"
    }
}

# Backup critical files
Backup-File "$FluentumHome\config\genesis.json" "$BackupDir\genesis.json.$Date" "genesis file"
Backup-File "$FluentumHome\config\priv_validator_key.json" "$BackupDir\priv_validator_key.json.$Date" "private validator key"
Backup-File "$FluentumHome\config\node_key.json" "$BackupDir\node_key.json.$Date" "node key"
Backup-File "$FluentumHome\data\priv_validator_state.json" "$BackupDir\priv_validator_state.json.$Date" "validator state"
Backup-File "$FluentumHome\config\config.toml" "$BackupDir\config.toml.$Date" "configuration file"

# Set proper permissions for sensitive files
Write-Log "Setting proper permissions for sensitive files"
try {
    Get-ChildItem -Path "$BackupDir\*.json.$Date" | ForEach-Object {
        $acl = Get-Acl $_.FullName
        $acl.SetAccessRuleProtection($true, $false)
        $rule = New-Object System.Security.AccessControl.FileSystemAccessRule("$env:USERNAME", "FullControl", "Allow")
        $acl.AddAccessRule($rule)
        Set-Acl $_.FullName $acl
    }
}
catch {
    Write-Log "Warning: Could not set permissions on some files: $_"
}

# Create checksums for integrity verification
Write-Log "Creating checksums for integrity verification"
try {
    $Checksums = @()
    Get-ChildItem -Path "$BackupDir\*.$Date" | ForEach-Object {
        $hash = Get-FileHash -Path $_.FullName -Algorithm SHA256
        $Checksums += "$($hash.Hash)  $($_.Name)"
    }
    $Checksums | Out-File -FilePath "$BackupDir\checksums.$Date" -Encoding UTF8
}
catch {
    Write-ErrorAndExit "Failed to create checksums: $_"
}

# Create backup manifest
$Manifest = @"
Fluentum Backup Manifest
========================
Backup Date: $(Get-Date)
Backup Directory: $BackupDir
Fluentum Home: $FluentumHome

Files Backed Up:
$((Get-ChildItem -Path "$BackupDir\*.$Date" | Where-Object { $_.Name -notmatch "checksums|manifest" } | ForEach-Object { "  $($_.Name) - $($_.Length) bytes" }) -join "`n")

Checksums:
$($Checksums -join "`n")

Backup completed successfully.
"@

$Manifest | Out-File -FilePath "$BackupDir\manifest.$Date" -Encoding UTF8

# Clean old backups (keep last 7 days)
Write-Log "Cleaning old backups (keeping last 7 days)"
try {
    $CutoffDate = (Get-Date).AddDays(-7).ToString("yyyyMMdd")
    Get-ChildItem -Path "$BackupDir\*" | Where-Object { 
        $_.Name -match "\.$CutoffDate" 
    } | Remove-Item -Force
}
catch {
    Write-Log "Warning: Could not clean old backups: $_"
}

# Calculate backup size
$BackupSize = (Get-ChildItem -Path $BackupDir -Recurse | Measure-Object -Property Length -Sum).Sum
$BackupSizeMB = [math]::Round($BackupSize / 1MB, 2)
$BackupCount = (Get-ChildItem -Path "$BackupDir\*.$Date" | Measure-Object).Count

Write-Log "✅ Backup completed successfully"
Write-Log "📊 Backup location: $BackupDir"
Write-Log "📊 Backup size: $BackupSizeMB MB"
Write-Log "📊 Files backed up: $BackupCount"
Write-Log "📊 Manifest: $BackupDir\manifest.$Date"
Write-Log "📊 Checksums: $BackupDir\checksums.$Date"

# Verify backup integrity
Write-Log "Verifying backup integrity..."
try {
    $ChecksumFile = "$BackupDir\checksums.$Date"
    if (Test-Path $ChecksumFile) {
        $ChecksumContent = Get-Content $ChecksumFile
        $AllValid = $true
        
        foreach ($line in $ChecksumContent) {
            if ($line -match "^([a-fA-F0-9]{64})\s+(.+)$") {
                $ExpectedHash = $matches[1]
                $FileName = $matches[2]
                $FilePath = Join-Path $BackupDir $FileName
                
                if (Test-Path $FilePath) {
                    $ActualHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash
                    if ($ActualHash -ne $ExpectedHash) {
                        Write-Log "❌ Checksum mismatch for $FileName"
                        $AllValid = $false
                    }
                }
            }
        }
        
        if ($AllValid) {
            Write-Log "✅ Backup integrity verified"
        }
        else {
            Write-Log "❌ Backup integrity check failed"
            exit 1
        }
    }
}
catch {
    Write-Log "❌ Backup integrity check failed: $_"
    exit 1
}

Write-Log "🎉 Fluentum backup completed successfully!" 