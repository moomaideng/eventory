param(
  [string]$EnvFile = ".env",
  [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
  [string[]]$GoArgs
)

. "$PSScriptRoot/Import-EventoryEnv.ps1" -EnvFile $EnvFile

$postgresPort = if ($env:POSTGRES_PORT) { $env:POSTGRES_PORT } else { "5432" }
$backendPort = if ($env:BACKEND_PORT) { $env:BACKEND_PORT } else { "8080" }

$env:DB_DSN = "host=localhost user=$($env:POSTGRES_USER) password=$($env:POSTGRES_PASSWORD) dbname=$($env:POSTGRES_DB) port=$postgresPort sslmode=disable"
$env:PORT = $backendPort
$env:ENVIRONMENT = "development"

Push-Location "$PSScriptRoot/../backend"
try {
  & go @GoArgs
  exit $LASTEXITCODE
} finally {
  Pop-Location
}
