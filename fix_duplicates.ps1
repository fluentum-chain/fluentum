# PowerShell script to check for duplicate, malformed, and other inconsistencies in go.sum and go.mod

$goSumPath = "go.sum"
$goModPath = "go.mod"

if (!(Test-Path $goSumPath)) {
    Write-Host "go.sum not found in current directory."
    exit 1
}
if (!(Test-Path $goModPath)) {
    Write-Host "go.mod not found in current directory."
    exit 1
}

# Read all lines
$lines = Get-Content $goSumPath
$goModContent = Get-Content $goModPath

# Check for duplicate lines
$duplicates = $lines | Group-Object | Where-Object { $_.Count -gt 1 }
if ($duplicates) {
    Write-Host "Duplicate lines found in go.sum:" -ForegroundColor Yellow
    foreach ($dup in $duplicates) {
        Write-Host ("  " + $dup.Name)
    }
} else {
    Write-Host "No duplicate lines found in go.sum." -ForegroundColor Green
}

# Check for malformed lines (should be 3 fields: module, version, hash)
$malformed = $lines | Where-Object { ($_ -split '\s+').Count -ne 3 }
if ($malformed) {
    Write-Host "Malformed lines found in go.sum:" -ForegroundColor Red
    foreach ($line in $malformed) {
        Write-Host ("  " + $line)
    }
} else {
    Write-Host "No malformed lines found in go.sum." -ForegroundColor Green
}

# Check for missing .mod/.zip pairs
$moduleVersions = @{ }
foreach ($line in $lines) {
    $parts = $line -split '\s+'
    if ($parts.Count -eq 3) {
        $mod = $parts[0]
        $ver = $parts[1]
        $key = if ($ver -like "*/go.mod") { "$mod $($ver -replace '/go.mod', '')" } else { "$mod $ver" }
        if (-not $moduleVersions.ContainsKey($key)) {
            $moduleVersions[$key] = @{ zip = $false; mod = $false }
        }
        if ($ver -like "*/go.mod") {
            $moduleVersions[$key].mod = $true
        } else {
            $moduleVersions[$key].zip = $true
        }
    }
}
$missingPairs = $moduleVersions.GetEnumerator() | Where-Object { -not ($_.Value.zip -and $_.Value.mod) }
if ($missingPairs) {
    Write-Host "Module versions missing .mod or .zip pairs:" -ForegroundColor Yellow
    foreach ($pair in $missingPairs) {
        $missing = @()
        if (-not $pair.Value.zip) { $missing += ".zip" }
        if (-not $pair.Value.mod) { $missing += ".mod" }
        Write-Host ("  $($pair.Key): missing $($missing -join ", ")")
    }
} else {
    Write-Host "All module versions have both .mod and .zip pairs." -ForegroundColor Green
}

# List modules in go.sum not present in go.mod
$goModModules = $goModContent | Where-Object { $_ -match '^[ \t]*require ' -or $_ -match '^[ \t]*replace ' } | ForEach-Object {
    ($_ -replace '^[ \t]*(require|replace)[ \t]+', '') -replace ' v[0-9].*', '' -replace '=>.*', '' -replace '"', ''
} | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' } | Sort-Object -Unique
$goSumModules = $lines | ForEach-Object {
    ($_ -split '\s+')[0]
} | Sort-Object -Unique
$notInGoMod = $goSumModules | Where-Object { $goModModules -notcontains $_ }
if ($notInGoMod) {
    Write-Host "Modules in go.sum but not in go.mod:" -ForegroundColor Yellow
    foreach ($mod in $notInGoMod) {
        Write-Host ("  " + $mod)
    }
} else {
    Write-Host "All modules in go.sum are referenced in go.mod." -ForegroundColor Green
}

# List modules with multiple versions
$modVerGroups = $lines | ForEach-Object {
    $parts = $_ -split '\s+'
    if ($parts.Count -eq 3) {
        $mod = $parts[0]
        $ver = $parts[1] -replace '/go.mod', ''
        "$mod $ver"
    }
} | Group-Object | Where-Object { $_.Count -gt 1 }
$multiVersions = $modVerGroups | Group-Object { ($_ -split ' ')[0] } | Where-Object { $_.Count -gt 1 }
if ($multiVersions) {
    Write-Host "Modules with multiple versions in go.sum:" -ForegroundColor Yellow
    foreach ($group in $multiVersions) {
        Write-Host ("  " + $group.Name)
    }
} else {
    Write-Host "No modules with multiple versions in go.sum." -ForegroundColor Green
}

# 1. Unused indirect dependencies (in go.mod but not used in code)
$indirects = $goModContent | Select-String '// indirect' | ForEach-Object {
    ($_ -split '\s+')[1]
}
if ($indirects) {
    $unusedIndirects = @()
    foreach ($ind in $indirects) {
        $found = Get-ChildItem -Recurse -Include *.go | Select-String -Pattern $ind -SimpleMatch -Quiet
        if (-not $found) {
            $unusedIndirects += $ind
        }
    }
    if ($unusedIndirects) {
        Write-Host "Unused indirect dependencies (in go.mod but not used in code):" -ForegroundColor Yellow
        foreach ($ind in $unusedIndirects) {
            Write-Host ("  " + $ind)
        }
    } else {
        Write-Host "All indirect dependencies are used in the codebase." -ForegroundColor Green
    }
} else {
    Write-Host "No indirect dependencies found in go.mod." -ForegroundColor Green
}

# 2. Stale replace directives (replace in go.mod not required)
$replaceDirectives = $goModContent | Where-Object { $_ -match '^[ \t]*replace ' }
$replaceTargets = $replaceDirectives | ForEach-Object {
    if ($_ -match 'replace ([^ ]+) => ([^ ]+)') { $matches[1].Trim() } else { $null }
} | Where-Object { $_ }
$staleReplaces = $replaceTargets | Where-Object { $goModModules -notcontains $_ }
if ($staleReplaces) {
    Write-Host "Stale replace directives (not required by any module):" -ForegroundColor Yellow
    foreach ($rep in $staleReplaces) {
        Write-Host ("  " + $rep)
    }
} else {
    Write-Host "No stale replace directives found." -ForegroundColor Green
}

# 3. Version mismatches (required version in go.mod vs. highest in go.sum)
$requireLines = $goModContent | Where-Object { $_ -match '^[ \t]*require ' }
$requireModules = $requireLines | ForEach-Object {
    $parts = $_ -split '\s+'
    if ($parts.Count -ge 3) { [PSCustomObject]@{ Name = $parts[1]; Version = $parts[2] } }
} | Where-Object { $_ }
$sumVersions = $lines | ForEach-Object {
    $parts = $_ -split '\s+'
    if ($parts.Count -eq 3) {
        [PSCustomObject]@{ Name = $parts[0]; Version = ($parts[1] -replace '/go.mod', '') }
    }
} | Group-Object Name
$versionMismatches = @()
foreach ($req in $requireModules) {
    $sumGroup = $sumVersions | Where-Object { $_.Name -eq $req.Name }
    if ($sumGroup) {
        $maxVer = ($sumGroup.Group | Sort-Object Version -Descending | Select-Object -First 1).Version
        if ($req.Version -ne $maxVer) {
            $versionMismatches += "  $($req.Name): go.mod=$($req.Version), go.sum=$maxVer"
        }
    }
}
if ($versionMismatches) {
    Write-Host "Version mismatches between go.mod and go.sum:" -ForegroundColor Yellow
    $versionMismatches | ForEach-Object { Write-Host $_ }
} else {
    Write-Host "No version mismatches between go.mod and go.sum." -ForegroundColor Green
}

# 4. Local replace directories (replace directives pointing to local paths)
$localReplaces = $replaceDirectives | Where-Object { $_ -match '=>\s*\.' -or $_ -match '=>\s*\\' -or $_ -match '=>\s*/' }
if ($localReplaces) {
    Write-Host "Local replace directives (may break builds on other machines):" -ForegroundColor Yellow
    foreach ($rep in $localReplaces) {
        Write-Host ("  " + $rep)
    }
} else {
    Write-Host "No local replace directives found." -ForegroundColor Green
}

# 5. Unpinned versions (pseudo-versions or branches instead of tags)
$unpinned = $requireModules | Where-Object {
    $_.Version -match '-' -or $_.Version -match 'master' -or $_.Version -match 'main'
}
if ($unpinned) {
    Write-Host "Unpinned versions (pseudo-versions or branches):" -ForegroundColor Yellow
    foreach ($mod in $unpinned) {
        Write-Host ("  $($mod.Name) $($mod.Version)")
    }
} else {
    Write-Host "All modules are pinned to tagged versions." -ForegroundColor Green
} 