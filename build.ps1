[CmdletBinding()]
param(
    [switch]$all,
    [switch]$cli,
    [switch]$web,
    [switch]$c
)

$ErrorActionPreference = 'Stop'

# Windows PowerShell 5.1 doesn't define $IsWindows
if ($null -eq $IsWindows) { $IsWindows = $true }

$buildDir = Join-Path $PSScriptRoot "build"

function New-OutDir([string]$name) {
    $dir = Join-Path $buildDir $name
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    return $dir
}

function Assert-Success([string]$what) {
    if ($LASTEXITCODE -ne 0) { throw "$what failed (exit code $LASTEXITCODE)" }
}

function Build-Cli {
    $outDir  = New-OutDir "cli"
    $exeName = if ($IsWindows) { "queso.exe" } else { "queso" }

    Push-Location (Join-Path $PSScriptRoot "cmd/cli")
    try {
        go build -o (Join-Path $outDir $exeName) .
        Assert-Success "Cli build"
    }
    finally { Pop-Location }

    Write-Host "Queso - Cli build completed."
}

function Build-Web {
    $outDir = New-OutDir "web"
    $prevGOOS, $prevGOARCH = $env:GOOS, $env:GOARCH

    Push-Location (Join-Path $PSScriptRoot "cmd/web")
    try {
        $env:GOOS   = "js"
        $env:GOARCH = "wasm"

        go build -o (Join-Path $outDir "main.wasm")
        Assert-Success "Web build"

        Copy-Item index.html, queso.js -Destination $outDir -Force
        Copy-Item "$(go env GOROOT)/lib/wasm/wasm_exec.js" -Destination $outDir -Force
    }
    finally {
        $env:GOOS, $env:GOARCH = $prevGOOS, $prevGOARCH
        Pop-Location
    }

    Write-Host "Queso - Web build completed."
}

function Build-C {
    $outDir = New-OutDir "c"
    $prevCGO = $env:CGO_ENABLED

    Push-Location (Join-Path $PSScriptRoot "cmd/c")
    try {
        $env:CGO_ENABLED = 1

        go build -buildmode=c-archive -o queso.a c.go
        Assert-Success "C build"

        Move-Item queso.a -Destination $outDir -Force
        Move-Item queso.h -Destination $outDir -Force
        Copy-Item libqueso.h -Destination $outDir -Force
    }
    finally {
        $env:CGO_ENABLED = $prevCGO
        Pop-Location
    }

    Write-Host "Queso - C build completed."
}

if (-not ($cli -or $web -or $c -or $all)) {
    Write-Host "Usage: ./build.ps1 [-cli] [-web] [-c] [-all]"
    exit 1
}

if ($cli -or $all) { Build-Cli }
if ($web -or $all) { Build-Web }
if ($c   -or $all) { Build-C }

Write-Host "Queso completed."
