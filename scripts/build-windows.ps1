$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

function Resolve-CommandPath {
    param(
        [Parameter(Mandatory = $true)]
        [string] $Name,
        [string[]] $Fallbacks = @()
    )

    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    foreach ($fallback in $Fallbacks) {
        if ($fallback -and (Test-Path -LiteralPath $fallback)) {
            return $fallback
        }
    }

    return $null
}

$go = Resolve-CommandPath -Name "go.exe"
$npm = Resolve-CommandPath -Name "npm.cmd"
$wails = Resolve-CommandPath -Name "wails.exe" -Fallbacks @(
    (Join-Path $env:USERPROFILE "go\bin\wails.exe")
)

if (-not $go) {
    throw "Go is required to build GRTBox. Install Go 1.22 or newer and reopen this terminal."
}

if (-not $npm) {
    throw "Node.js/npm is required to build GRTBox. Install Node.js 18 or newer and reopen this terminal."
}

if (-not $wails) {
    throw "Wails CLI is required. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2"
}

Write-Host "Building GRTBox.exe for Windows..."
$env:PATH = "$(Split-Path -Parent $go);$env:PATH"
$exe = Join-Path $root "build\bin\GRTBox.exe"
$buildStartedAt = Get-Date
& $wails build -clean -platform windows/amd64
if ($LASTEXITCODE -ne 0) {
    throw "Wails build failed with exit code $LASTEXITCODE."
}

if (Test-Path -LiteralPath $exe) {
    $exeInfo = Get-Item -LiteralPath $exe
    if ($exeInfo.LastWriteTime -lt $buildStartedAt) {
        throw "GRTBox.exe was not updated. Close any running GRTBox process and run this build again."
    }
    Write-Host "Built desktop executable:"
    Write-Host $exe
} else {
    throw "Build completed but build\bin\GRTBox.exe was not created."
}
