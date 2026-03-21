$ErrorActionPreference = "Stop"

param(
    [string]$BaseDir = "D:\Workspace\guilauncher"
)

$configFile = Join-Path $BaseDir "config.json"
$paramFile  = Join-Path $BaseDir "run.json"
$logDir     = Join-Path $BaseDir "logs"
$logFile    = Join-Path $logDir "launcher.log"

function Write-Log {
    param([string]$Message)
    $time = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Add-Content -Path $logFile -Value "$time $Message"
}

function Stop-AppProcesses {
    param(
        [string[]]$ProcessNames,
        [int]$TimeoutSec = 5
    )

    $allStopped = $true
    foreach ($name in $ProcessNames) {
        if ([string]::IsNullOrWhiteSpace($name)) {
            continue
        }
        $procs = @(Get-Process -Name $name -ErrorAction SilentlyContinue)
        if ($procs.Count -eq 0) {
            Write-Log "restart: no running process matched name=$name"
            continue
        }
        Write-Log "restart: stopping name=$name count=$($procs.Count)"
        foreach ($p in $procs) {
            try { $null = $p.CloseMainWindow() } catch {}
        }
        Start-Sleep -Milliseconds 500
        $remaining = @(Get-Process -Name $name -ErrorAction SilentlyContinue)
        foreach ($p in $remaining) {
            try {
                Stop-Process -Id $p.Id -Force -ErrorAction Stop
            } catch {
                Write-Log "restart: failed pid=$($p.Id) name=$name err=$($_.Exception.Message)"
                $allStopped = $false
            }
        }
        $deadline = (Get-Date).AddSeconds($TimeoutSec)
        do {
            $left = @(Get-Process -Name $name -ErrorAction SilentlyContinue)
            if ($left.Count -eq 0) { break }
            Start-Sleep -Milliseconds 200
        } while ((Get-Date) -lt $deadline)
        if (@(Get-Process -Name $name -ErrorAction SilentlyContinue).Count -gt 0) {
            Write-Log "restart: process still running name=$name"
            $allStopped = $false
        } else {
            Write-Log "restart: stopped name=$name"
        }
    }
    return $allStopped
}

try {
    if (-not (Test-Path $logDir)) {
        New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    }
    if (-not (Test-Path $configFile)) {
        throw "config.json not found: $configFile"
    }
    if (-not (Test-Path $paramFile)) {
        Write-Log "run.json not found, nothing to do"
        exit 0
    }

    $config = Get-Content -Path $configFile -Raw | ConvertFrom-Json
    $param  = Get-Content -Path $paramFile -Raw | ConvertFrom-Json
    if (-not $param.app) {
        throw "missing field: app"
    }

    $appId = [string]$param.app
    if (-not $config.apps.PSObject.Properties.Name.Contains($appId)) {
        throw "app id not allowed: $appId"
    }

    $appConfig = $config.apps.$appId
    $exe = [string]$appConfig.exe
    $workdir = [string]$appConfig.workdir
    $restart = $false
    if ($null -ne $param.restart) {
        $restart = [bool]$param.restart
    }
    if ([string]::IsNullOrWhiteSpace($exe)) {
        throw "exe is empty for app id: $appId"
    }
    if (-not (Test-Path $exe)) {
        throw "executable not found: $exe"
    }
    if ([string]::IsNullOrWhiteSpace($workdir)) {
        $workdir = Split-Path -Path $exe -Parent
    }
    if (-not (Test-Path $workdir)) {
        throw "working directory not found: $workdir"
    }

    $processNames = @()
    if ($null -ne $appConfig.processNames) {
        foreach ($name in $appConfig.processNames) {
            $processNames += [string]$name
        }
    }
    if ($processNames.Count -eq 0) {
        $processNames = @([System.IO.Path]::GetFileNameWithoutExtension($exe))
    }

    if ($restart) {
        Write-Log "restart requested app=$appId names=$($processNames -join ',')"
        $stopped = Stop-AppProcesses -ProcessNames $processNames -TimeoutSec 5
        if (-not $stopped) {
            Write-Log "restart warning: some processes may still be running"
        }
    }

    Write-Log "launching app=$appId exe=$exe workdir=$workdir restart=$restart"
    Start-Process -FilePath $exe -WorkingDirectory $workdir -WindowStyle Normal
    Remove-Item -Path $paramFile -Force -ErrorAction SilentlyContinue
    Write-Log "launch success app=$appId"
}
catch {
    Write-Log "ERROR: $($_.Exception.Message)"
    exit 1
}
