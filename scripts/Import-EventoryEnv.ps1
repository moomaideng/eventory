param(
  [Parameter(Mandatory = $true)]
  [string]$EnvFile
)

if (-not (Test-Path -LiteralPath $EnvFile)) {
  throw "Missing ${EnvFile}. Copy .env.example to .env first."
}

Get-Content -LiteralPath $EnvFile | ForEach-Object {
  $line = $_.Trim()

  if (-not $line -or $line.StartsWith("#")) {
    return
  }

  $parts = $line -split "=", 2
  if ($parts.Count -ne 2 -or -not $parts[0].Trim()) {
    throw "Invalid environment variable in ${EnvFile}: $line"
  }

  $key = $parts[0].Trim()
  $value = $parts[1].Trim()

  # Strip surrounding quotes if present (e.g. KEY="value" or KEY='value')
  if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
    if ($value.Length -ge 2) {
      $value = $value.Substring(1, $value.Length - 2)
    }
  }

  Set-Item "env:$key" $value
  [Environment]::SetEnvironmentVariable($key, $value, "Process")
}
