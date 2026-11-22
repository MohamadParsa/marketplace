[CmdletBinding()]
param([switch]$NoBuild)
$ErrorActionPreference = 'Stop'
$projectRoot = $PSScriptRoot
$runtimeRoot = Join-Path $env:USERPROFILE '.codex/local-runtimes/cherriva'
New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
$go = (Get-Command go.exe -ErrorAction SilentlyContinue).Source
if (-not $go) { $go = Join-Path $env:LOCALAPPDATA 'Programs/Go1.27.4/go/bin/go.exe' }
if (-not (Test-Path $go)) { throw 'Install Go matching apps/api/go.mod before starting.' }
if (-not $env:REVIEW_DATABASE_URL) {
    $pgCtl = Join-Path $runtimeRoot 'postgres-native/package/native/bin/pg_ctl.exe'
    $pgData = Join-Path $runtimeRoot 'postgres-data'
    if (-not (Test-Path $pgCtl)) { throw 'Set REVIEW_DATABASE_URL to an isolated PostgreSQL database. See docs/delivery/local-review-access.md.' }
    & $pgCtl -D $pgData status *> $null
    if ($LASTEXITCODE -ne 0) {
        & $pgCtl -D $pgData -l (Join-Path $runtimeRoot 'postgres.log') -o '-h 127.0.0.1 -p 55437' -w start
        if ($LASTEXITCODE -ne 0) { throw 'Local review PostgreSQL failed to start.' }
    }
    $env:REVIEW_DATABASE_URL = 'postgres://postgres@127.0.0.1:55437/postgres?sslmode=disable'
}
if (-not $NoBuild) {
    foreach ($app in @('backoffice-web', 'storefront-web')) {
        $appPath = Join-Path $projectRoot "apps/$app"
        & npm.cmd --prefix $appPath ci --prefer-offline --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) { throw "Dependency installation failed: $app" }
        & npm.cmd --prefix $appPath run build
        if ($LASTEXITCODE -ne 0) { throw "Build failed: $app" }
    }
}
$apiExe = Join-Path $runtimeRoot 'api.exe'
foreach ($port in @(8098,5197)) {
    $listener = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($listener) {
        $proc = Get-CimInstance Win32_Process -Filter "ProcessId=$($listener.OwningProcess)"
        if ($proc.ExecutablePath -notlike "$runtimeRoot\api*.exe") { throw "Port $port belongs to another application." }
        Stop-Process -Id $listener.OwningProcess
    }
}
& $go -C (Join-Path $projectRoot 'apps/api') build -o $apiExe ./cmd/api
if ($LASTEXITCODE -ne 0) { throw 'API build failed.' }
$env:APP_ENV = 'test'
$env:BIND_HOST = '127.0.0.1'
$env:KAVENEGAR_API_KEY = ''
$env:KAVENEGAR_TEMPLATE = ''
foreach ($surface in @(@{Name='backoffice';App='backoffice-web';Port='8098'},@{Name='storefront';App='storefront-web';Port='5197'})) {
    $env:PORT = $surface.Port
    $env:WEB_SURFACE = $surface.Name
    $env:WEB_ROOT = Join-Path $projectRoot "apps/$($surface.App)/build"
    if (-not (Test-Path (Join-Path $env:WEB_ROOT 'index.html'))) { throw 'Build the frontend before using -NoBuild.' }
    Start-Process $apiExe -WorkingDirectory (Join-Path $projectRoot 'apps/api') -RedirectStandardOutput (Join-Path $runtimeRoot "$($surface.Name).out.log") -RedirectStandardError (Join-Path $runtimeRoot "$($surface.Name).err.log")
}
Write-Host "Source: $projectRoot"
Write-Host 'Backoffice: http://localhost:8098'
Write-Host 'Storefront: http://localhost:5197'
