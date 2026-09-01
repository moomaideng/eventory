param(
  [string]$EnvFile = ".env",
  [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
  [string[]]$NpmArgs
)

. "$PSScriptRoot/Import-EventoryEnv.ps1" -EnvFile $EnvFile

$apiUrl = if ($env:API_URL) { $env:API_URL } else { "http://localhost:8080" }

$env:NEXT_PUBLIC_SUPABASE_URL = $env:SUPABASE_URL
$env:NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY = $env:SUPABASE_PUBLISHABLE_KEY
$env:NEXT_PUBLIC_API_URL = $apiUrl
$env:INTERNAL_API_URL = $apiUrl

Push-Location "$PSScriptRoot/../frontend"
try {
  if ($NpmArgs[0] -eq "npm") {
    $remaining = if ($NpmArgs.Length -gt 1) { $NpmArgs[1..($NpmArgs.Length - 1)] } else { @() }
    & npm @remaining
  } else {
    & npm run @NpmArgs
  }
  if ($LASTEXITCODE -ne $null) {
    exit $LASTEXITCODE
  }
} finally {
  Pop-Location
}
