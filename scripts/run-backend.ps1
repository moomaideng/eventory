param(
  [string]$EnvFile = ".env",
  [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
  [string[]]$CommandArgs
)

. "$PSScriptRoot/Import-EventoryEnv.ps1" -EnvFile $EnvFile

$postgresPort = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { "5432" }
$backendPort = if ($env:BACKEND_PORT) { $env:BACKEND_PORT } else { "8080" }

$env:DB_DSN = "host=localhost user=$($env:POSTGRES_USER) password=$($env:POSTGRES_PASSWORD) dbname=$($env:POSTGRES_DB) port=$postgresPort sslmode=disable"
$env:PORT = $backendPort
$env:ENVIRONMENT = "development"

Push-Location "$PSScriptRoot/../backend"
try {
  if ($CommandArgs[0] -eq "air") {
    $remaining = if ($CommandArgs.Length -gt 1) { $CommandArgs[1..($CommandArgs.Length - 1)] } else { @() }
    & air @remaining
  } elseif ($CommandArgs[0] -eq "go") {
    $remaining = if ($CommandArgs.Length -gt 1) { $CommandArgs[1..($CommandArgs.Length - 1)] } else { @() }
    & go @remaining
  } else {
    & go @CommandArgs
  }
  if ($LASTEXITCODE -ne $null) {
    exit $LASTEXITCODE
  }
} finally {
  Pop-Location
}
