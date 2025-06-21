$content = Get-Content go.sum
$newContent = $content[0..78]
$newContent += "cosmossdk.io/snapshots v1.0.2 h1:lSg5BTvJBHUDwswNNyeh4K/CbqiHER73VU4nDNb8uk0="
$newContent += "cosmossdk.io/snapshots v1.0.2/go.mod h1:EFtENTqVTuWwitGW1VwaBct+yDagk7oG/axBMPH+FXs="
$newContent += $content[79..($content.Length-1)]
$newContent | Set-Content go.sum 