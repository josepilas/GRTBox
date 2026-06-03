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
    throw "Go is required to run GRTBox in desktop dev mode. Install Go 1.22 or newer and reopen this terminal."
}

if (-not $npm) {
    throw "Node.js/npm is required to run GRTBox in desktop dev mode. Install Node.js 18 or newer and reopen this terminal."
}

if (-not $wails) {
    throw "Wails CLI is required. Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2"
}

Write-Host "Starting GRTBox as a Wails desktop application..."
Write-Host "This is not an Electron app and not a web-only launch."
& $wails dev
