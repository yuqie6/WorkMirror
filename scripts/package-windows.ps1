Param(
  [string]$Version = "",
  [string]$OutDir = "",
  [string]$IsccPath = ""
)

$ErrorActionPreference = "Stop"

function Assert-CommandExists([string]$name) {
  $cmd = Get-Command $name -ErrorAction SilentlyContinue
  if (!$cmd) { throw "未找到命令：$name。请先安装并确保在 PATH 中可用。" }
  return $cmd.Source
}

function Try-WarnVersion([string]$name, [ScriptBlock]$getVersion, [ScriptBlock]$isOk) {
  try {
    $v = & $getVersion
    if (![string]::IsNullOrWhiteSpace($v) -and !( & $isOk $v)) {
      Write-Warning "$name 版本可能不符合要求：$v"
    }
  } catch {
    Write-Warning "无法检查 $name 版本：$($_.Exception.Message)"
  }
}

function Read-VersionFromConfigExample([string]$repoRoot) {
  $cfgPath = Join-Path $repoRoot "config/config.yaml.example"
  if (!(Test-Path $cfgPath)) { return "" }
  $content = Get-Content -Path $cfgPath -Raw
  $m = [Regex]::Match($content, '(?m)^\s*version:\s*"?([0-9]+\.[0-9]+\.[0-9]+)"?\s*$')
  if ($m.Success) { return $m.Groups[1].Value }
  return ""
}

function Find-Iscc {
  if (![string]::IsNullOrWhiteSpace($IsccPath)) {
    if (Test-Path $IsccPath) { return $IsccPath }
    Write-Warning "指定了 -IsccPath '$IsccPath' 但文件不存在。"
  }

  $iscc = (Get-Command "ISCC.exe" -ErrorAction SilentlyContinue)
  if ($iscc) { return $iscc.Source }

  $candidates = @(
    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
    "${env:ProgramFiles}\Inno Setup 6\ISCC.exe",
    "${env:LocalAppData}\Programs\Inno Setup 6\ISCC.exe"
  )
  foreach ($c in $candidates) {
    if (Test-Path $c) { return $c }
  }
  return ""
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $repoRoot

try {
  Assert-CommandExists "go" | Out-Null
  Assert-CommandExists "npm" | Out-Null
  Assert-CommandExists "wails" | Out-Null

  Try-WarnVersion "Go" { (& go version) } { param($s) $s -match 'go1\.25\.' }
  Try-WarnVersion "Node" { (& node -v) } { param($s) ([int]($s.TrimStart('v').Split('.')[0])) -ge 16 }

  if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = Read-VersionFromConfigExample -repoRoot $repoRoot
  }
  if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = "0.1.0"
  }

  $stagingDir = Join-Path $repoRoot "build/out/windows/staging"
  $installerOutDir = if ([string]::IsNullOrWhiteSpace($OutDir)) { Join-Path $repoRoot "build/installer/windows/out" } else { $OutDir }

  New-Item -ItemType Directory -Force -Path $stagingDir | Out-Null
  New-Item -ItemType Directory -Force -Path $installerOutDir | Out-Null
  New-Item -ItemType Directory -Force -Path (Join-Path $stagingDir "config") | Out-Null

  Write-Host "==> build agent (mirror.exe)"
  & go build -trimpath -ldflags "-H=windowsgui -s -w" -o (Join-Path $stagingDir "mirror.exe") "./cmd/mirror-agent/"

  Write-Host "==> build ui (mirror-ui.exe)"
  $nodeModulesPath = Join-Path $repoRoot "cmd/mirror-ui/frontend/node_modules"
  if (!(Test-Path $nodeModulesPath)) {
    Write-Host "==> install frontend deps (npm install)"
    Push-Location (Join-Path $repoRoot "cmd/mirror-ui/frontend")
    try {
      & npm install
    } finally {
      Pop-Location
    }
  }
  Push-Location (Join-Path $repoRoot "cmd/mirror-ui")
  try {
    & wails build -clean
  } finally {
    Pop-Location
  }
  Copy-Item -Force -Path (Join-Path $repoRoot "cmd/mirror-ui/build/bin/mirror-ui.exe") -Destination (Join-Path $stagingDir "mirror-ui.exe")

  Copy-Item -Force -Path (Join-Path $repoRoot "config/config.yaml.example") -Destination (Join-Path $stagingDir "config/config.yaml.example")

  $foundIsccPath = Find-Iscc
  if ([string]::IsNullOrWhiteSpace($foundIsccPath)) {
    $candidatesStr = "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe, ${env:ProgramFiles}\Inno Setup 6\ISCC.exe"
    throw "未找到 Inno Setup 编译器 ISCC.exe。已搜索 PATH 及默认路径: $candidatesStr。请安装 Inno Setup 6，或使用 -IsccPath 参数指定路径。"
  }
  Write-Host "Using ISCC at: $foundIsccPath"

  $env:MIRROR_APP_VERSION = $Version
  $env:MIRROR_STAGING_DIR = $stagingDir
  $env:MIRROR_OUTPUT_DIR = $installerOutDir

  $issPath = Join-Path $repoRoot "installer/windows/mirror.iss"
  Write-Host "==> build installer: $issPath"
  & "$foundIsccPath" "$issPath"

  Write-Host "==> done"
  Write-Host "Installer output: $installerOutDir"
} finally {
  Pop-Location
}
