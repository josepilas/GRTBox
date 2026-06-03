$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

function Get-AppVersion {
    $appFile = Join-Path $root "app.go"
    $content = Get-Content -Raw -LiteralPath $appFile
    $match = [regex]::Match($content, 'CurrentGRTBoxVersion\s*=\s*"([^"]+)"')
    if ($match.Success) {
        return $match.Groups[1].Value
    }
    return "0.1.0"
}

$version = Get-AppVersion
$releaseRoot = Join-Path $root "dist"
$releaseDir = Join-Path $releaseRoot ("GRTBox-{0}-windows-x64" -f $version)
$primaryExe = Join-Path $root "build\bin\GRTBox.exe"
$releaseExe = Join-Path $root "build\bin\GRTBox-release.exe"
$exe = $primaryExe

if (Test-Path -LiteralPath $releaseExe) {
    $exe = $releaseExe
}

if (-not (Test-Path -LiteralPath $exe)) {
    throw "No GRTBox executable was found in build\bin. Run Build-GRTBox-Windows.cmd first."
}

if (Test-Path -LiteralPath $releaseDir) {
    Remove-Item -LiteralPath $releaseDir -Recurse -Force
}

New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $releaseDir "docs") | Out-Null

Copy-Item -LiteralPath $exe -Destination (Join-Path $releaseDir "GRTBox.exe") -Force
Copy-Item -LiteralPath (Join-Path $root "README.md") -Destination (Join-Path $releaseDir "README.md") -Force

$devguide = Join-Path $root "docs\devguide.md"
if (Test-Path -LiteralPath $devguide) {
    Copy-Item -LiteralPath $devguide -Destination (Join-Path $releaseDir "docs\devguide.md") -Force
}

$checklist = Join-Path $root "docs\RELEASE_CHECKLIST.md"
if (Test-Path -LiteralPath $checklist) {
    Copy-Item -LiteralPath $checklist -Destination (Join-Path $releaseDir "docs\RELEASE_CHECKLIST.md") -Force
}

$portability = Join-Path $root "docs\PORTABILITY_ANALYSIS.md"
if (Test-Path -LiteralPath $portability) {
    Copy-Item -LiteralPath $portability -Destination (Join-Path $releaseDir "docs\PORTABILITY_ANALYSIS.md") -Force
}

$manifest = [PSCustomObject]@{
    name = "GRTBox"
    version = $version
    platform = "windows-x64"
    created_at = (Get-Date).ToString("o")
    executable = "GRTBox.exe"
    source_executable = (Split-Path -Leaf $exe)
    bundled_tools = @()
    first_run_tool_store_url = "https://grtbox.unaux.com/tools/org/tools.json"
    first_run_tool_ids = @("dtb_general", "easy_firmware", "internet_bridge")
}
$manifest | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $releaseDir "release.json") -Encoding UTF8

$zipPath = Join-Path $releaseRoot ("GRTBox-{0}-windows-x64.zip" -f $version)
if (Test-Path -LiteralPath $zipPath) {
    Remove-Item -LiteralPath $zipPath -Force
}
Compress-Archive -LiteralPath $releaseDir -DestinationPath $zipPath -Force

Write-Host "Release folder created:"
Write-Host $releaseDir
Write-Host "Release zip created:"
Write-Host $zipPath
