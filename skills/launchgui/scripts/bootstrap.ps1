param(
    [string]$BaseDir = "D:\Workspace\guilauncher",
    [string]$TaskName = "guilauncher"
)

$ErrorActionPreference = "Stop"

$skillDir = Split-Path -Parent $PSScriptRoot
$tplDir = Join-Path $skillDir "templates"
$launchTpl = Join-Path $tplDir "launchgui.ps1"
$configTpl = Join-Path $tplDir "config.json"

if (-not (Test-Path $launchTpl)) {
    throw "template missing: $launchTpl"
}
if (-not (Test-Path $configTpl)) {
    throw "template missing: $configTpl"
}

New-Item -ItemType Directory -Force -Path $BaseDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $BaseDir "logs") | Out-Null

$launchPath = Join-Path $BaseDir "launchgui.ps1"
$configPath = Join-Path $BaseDir "config.json"

if (-not (Test-Path $launchPath)) {
    Copy-Item $launchTpl $launchPath
    Write-Output "created: $launchPath"
} else {
    Write-Output "exists:  $launchPath"
}

if (-not (Test-Path $configPath)) {
    Copy-Item $configTpl $configPath
    Write-Output "created: $configPath"
} else {
    Write-Output "exists:  $configPath"
}

# ONDEMAND is not supported on all environments; ONCE works for manual /run triggering.
$taskCmd = "powershell.exe -ExecutionPolicy Bypass -File $launchPath"
$createCmd = "schtasks /create /tn `"$TaskName`" /tr `"$taskCmd`" /sc ONCE /st 00:00 /rl LIMITED /f"
try {
    cmd /c $createCmd 2>$null | Out-Null
    $q = cmd /c "schtasks /query /tn `"$TaskName`"" 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Output "ensured task: $TaskName"
    } else {
        Write-Warning "task create/query failed in current environment, task may be unavailable: $TaskName"
    }
} catch {
    Write-Warning "failed to create/query schtasks in current environment: $($_.Exception.Message)"
    Write-Warning "fallback: run launch script directly -> powershell -ExecutionPolicy Bypass -File $launchPath"
}

Write-Output ""
Write-Output "Next:"
Write-Output "@'"
Write-Output "{"
Write-Output "  `"app`": `"tankbattle`","
Write-Output "  `"restart`": true"
Write-Output "}"
Write-Output "'@ | Set-Content -Path $BaseDir\\run.json -Encoding UTF8"
Write-Output "schtasks /run /tn `"$TaskName`""
