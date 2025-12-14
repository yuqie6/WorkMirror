param(
    [Parameter(Mandatory = $false)]
    [string] $Version = "v0.2.0-alpha.1",

    [Parameter(Mandatory = $false)]
    [string] $Arch = "x64",

    [Parameter(Mandatory = $false)]
    [string] $OutDir = "dist/release",

    [Parameter(Mandatory = $false)]
    [switch] $SkipUI
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Require-Command([string] $Name) {
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $cmd) {
        throw "missing required command: $Name"
    }
}

$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "../..")
Push-Location $RepoRoot
try {
    Require-Command "go"

    if (-not $SkipUI) {
        if (Test-Path -LiteralPath "frontend/package.json") {
            if (-not (Test-Path -LiteralPath "frontend/dist/index.html")) {
                if (Get-Command "pnpm" -ErrorAction SilentlyContinue) {
                    pnpm -C "frontend" build
                }
                elseif (Get-Command "npm" -ErrorAction SilentlyContinue) {
                    npm --prefix "frontend" run build
                }
                else {
                    throw "neither pnpm nor npm found; run UI build manually or pass -SkipUI"
                }
            }
        }
    }

    $StagingRoot = Join-Path $OutDir ("WorkMirror-{0}-windows-{1}" -f $Version, $Arch)
    if (Test-Path -LiteralPath $StagingRoot) {
        Remove-Item -Recurse -Force -LiteralPath $StagingRoot
    }
    New-Item -ItemType Directory -Force -Path $StagingRoot | Out-Null

    $ldflags = @(
        "-H=windowsgui",
        "-s",
        "-w",
        ("-X github.com/yuqie6/WorkMirror/internal/pkg/buildinfo.Version={0}" -f $Version)
    ) -join " "

    go build -trimpath -ldflags $ldflags -o (Join-Path $StagingRoot "workmirror.exe") "./cmd/workmirror-agent/"

    # UI assets: ship as external folder (server prefers ./frontend/dist next to exe).
    if (Test-Path -LiteralPath "frontend/dist/index.html") {
        New-Item -ItemType Directory -Force -Path (Join-Path $StagingRoot "frontend") | Out-Null
        Copy-Item -Recurse -Force -LiteralPath "frontend/dist" -Destination (Join-Path $StagingRoot "frontend")
    }

    # Config template
    New-Item -ItemType Directory -Force -Path (Join-Path $StagingRoot "config") | Out-Null
    if (Test-Path -LiteralPath "config/config.yaml.example") {
        Copy-Item -Force -LiteralPath "config/config.yaml.example" -Destination (Join-Path $StagingRoot "config/config.yaml.example")
    }

    # Quick start note (portable folder model)
    $readme = @"
WorkMirror $Version (portable folder)

1) Unzip to a fixed folder you will keep (avoid running from Downloads).
2) Double-click workmirror.exe.
3) First run auto-creates: ./config/config.yaml ./data/ ./logs/
4) To migrate/backup, move the whole folder.
"@
    Set-Content -Encoding UTF8 -LiteralPath (Join-Path $StagingRoot "README.txt") -Value $readme

    if (Test-Path -LiteralPath "LICENSE") {
        Copy-Item -Force -LiteralPath "LICENSE" -Destination (Join-Path $StagingRoot "LICENSE")
    }

    New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
    $ZipPath = Join-Path $OutDir ("WorkMirror-{0}-windows-{1}.zip" -f $Version, $Arch)
    if (Test-Path -LiteralPath $ZipPath) {
        Remove-Item -Force -LiteralPath $ZipPath
    }
    Compress-Archive -Force -Path $StagingRoot -DestinationPath $ZipPath

    Write-Host "OK: $ZipPath"
}
finally {
    Pop-Location
}
