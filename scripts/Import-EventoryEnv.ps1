param(
  [Parameter(Mandatory = $true)]
  [string]$EnvFile
)

if (-not (Test-Path -LiteralPath $EnvFile)) {
  throw "Missing $EnvFile. Copy .env.example to .env first."
}

Get-Content -LiteralPath $EnvFile | ForEach-Object {
  $line = $_.Trim()

  if (-not $line -or $line.StartsWith("#")) {
    return
  }

  $parts = $line -split "=", 2
  if ($parts.Count -ne 2 -or -not $parts[0].Trim()) {
    throw "Invalid environment variable in $EnvFile: $line"
  }

  [Environment]::SetEnvironmentVariable(
    $parts[0].Trim(),
    $parts[1].Trim(),
    "Process"
  )
}
