$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoDir = Split-Path -Parent $PSScriptRoot
$originalLocation = Get-Location
$processes = @()

try {
    Set-Location $repoDir

    $requiredFiles = @("backend/.env", "frontend/.env.local")
    foreach ($envFile in $requiredFiles) {
        if (-not (Test-Path $envFile -PathType Leaf)) {
            $exampleFile = if ($envFile -eq "backend/.env") {
                "backend/.env.example"
            } else {
                "frontend/.env.example"
            }
            throw "Missing $envFile. Copy $exampleFile to $envFile and configure it."
        }
    }

    foreach ($commandName in @("docker", "go", "npm.cmd")) {
        if (-not (Get-Command $commandName -ErrorAction SilentlyContinue)) {
            throw "Required command not found: $commandName"
        }
    }

    $airCommand = Get-Command air -ErrorAction SilentlyContinue
    if ($airCommand) {
        $airPath = $airCommand.Source
    } else {
        $goBin = (& go env GOBIN).Trim()
        if (-not $goBin) {
            $goBin = Join-Path ((& go env GOPATH).Trim()) "bin"
        }
        $airPath = Join-Path $goBin "air.exe"
    }
    if (-not (Test-Path $airPath -PathType Leaf)) {
        throw "Air is not installed. Run 'make install' first."
    }

    $composeArgs = @(
        "compose",
        "--env-file", "backend/.env",
        "--env-file", "frontend/.env.local"
    )

    Write-Host "Stopping containerized frontend and backend..."
    & docker @composeArgs stop frontend backend
    if ($LASTEXITCODE -ne 0) {
        throw "Docker Compose failed to stop the containerized application services."
    }

    Write-Host "Starting PostgreSQL..."
    & docker @composeArgs up -d --wait postgres
    if ($LASTEXITCODE -ne 0) {
        throw "Docker Compose failed to start PostgreSQL."
    }

    Write-Host "Running database migrations..."
    Push-Location backend
    try {
        & go run ./cmd/migrate
        if ($LASTEXITCODE -ne 0) {
            throw "Database migration failed."
        }
    } finally {
        Pop-Location
    }

    Write-Host "Starting backend with Air and frontend with Next.js..."
    $npm = Get-Command npm.cmd

    $backendProcess = Start-Process `
        -FilePath $airPath `
        -ArgumentList @("-c", ".air.toml") `
        -WorkingDirectory (Join-Path $repoDir "backend") `
        -NoNewWindow `
        -PassThru
    $processes += $backendProcess

    $frontendProcess = Start-Process `
        -FilePath $npm.Source `
        -ArgumentList @("run", "dev") `
        -WorkingDirectory (Join-Path $repoDir "frontend") `
        -NoNewWindow `
        -PassThru
    $processes += $frontendProcess

    while (-not $backendProcess.HasExited -and -not $frontendProcess.HasExited) {
        Start-Sleep -Milliseconds 500
        $backendProcess.Refresh()
        $frontendProcess.Refresh()
    }

    if ($backendProcess.HasExited) {
        exit $backendProcess.ExitCode
    }
    exit $frontendProcess.ExitCode
} finally {
    foreach ($process in $processes) {
        if (-not $process.HasExited) {
            Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            $process.WaitForExit()
        }
    }
    Set-Location $originalLocation
}
