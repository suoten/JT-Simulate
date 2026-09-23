# JT-Simulate 发布打包脚本
# 用法: powershell -ExecutionPolicy Bypass -File make-release.ps1
# 用法: powershell -ExecutionPolicy Bypass -File make-release.ps1 -Version 1.0.0
# 产出: release/jt-simulate-<version>-<platform>.zip
#
# 流程:
#   1. Windows: wails build（桌面应用，双击弹出原生窗口）
#      Linux/macOS: go build（Web 服务模式，命令行启动+浏览器访问）
#   2. 打包发布 zip（二进制 + 配置文件 + README.md）
#   3. 生成 SHA256 校验文件

param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ProjectRoot

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  JT-Simulate v$Version Release Builder" -ForegroundColor Cyan
Write-Host "  部标协议通用仿真平台" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================
# 1. 编译多平台二进制
# ============================================================
Write-Host "[1/3] Building multi-platform binaries..." -ForegroundColor Green

$distDir = "dist"
New-Item -ItemType Directory -Path $distDir -Force | Out-Null

$ldflags = "-s -w -X main.version=$Version"
$binaries = @{}

# --- Windows amd64 (Wails 桌面应用) ---
Write-Host "  Building windows/amd64 (Wails desktop app)..." -ForegroundColor DarkGray
$env:GOPROXY = "https://goproxy.cn,direct"
$prevEAP = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
& wails build -platform windows/amd64 -ldflags $ldflags -nopackage 2>&1 | Out-Null
$buildExit = $LASTEXITCODE; $ErrorActionPreference = $prevEAP
$wailsExe = "build/bin/jt-simulate.exe"
if ($buildExit -ne 0 -or -not (Test-Path $wailsExe)) {
    Write-Host "  wails build failed, falling back to go build..." -ForegroundColor Yellow
    $env:GOOS = "windows"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
    $out = "$distDir/jt-simulate-windows-amd64.exe"
    & go build -ldflags="$ldflags" -o $out . 2>&1 | Out-Null
    $env:GOOS = ""; $env:GOARCH = ""; $env:CGO_ENABLED = ""
    if (-not (Test-Path $out)) { Write-Host "ERROR: Build failed for windows-amd64" -ForegroundColor Red; exit 1 }
} else {
    $out = "$distDir/jt-simulate-windows-amd64.exe"
    Copy-Item $wailsExe $out -Force
}
$binaries["windows-amd64"] = $out
Write-Host "    -> jt-simulate-windows-amd64.exe ($([math]::Round((Get-Item $out).Length / 1MB, 1)) MB)" -ForegroundColor Yellow

# --- Linux amd64 (Web 服务模式) ---
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"; $env:GOPROXY = "https://goproxy.cn,direct"
$out = "$distDir/jt-simulate-linux-amd64"
Write-Host "  Building linux/amd64..." -ForegroundColor DarkGray
$prevEAP = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
& go build -ldflags="$ldflags" -o $out . 2>&1 | Out-Null
$buildExit = $LASTEXITCODE; $ErrorActionPreference = $prevEAP
if ($buildExit -ne 0 -or -not (Test-Path $out)) { Write-Host "ERROR: Build failed for linux-amd64 (exit $buildExit)" -ForegroundColor Red; exit 1 }
$binaries["linux-amd64"] = $out
Write-Host "    -> jt-simulate-linux-amd64 ($([math]::Round((Get-Item $out).Length / 1MB, 1)) MB)" -ForegroundColor Yellow

# --- Linux arm64 (Web 服务模式) ---
$env:GOOS = "linux"; $env:GOARCH = "arm64"; $env:CGO_ENABLED = "0"; $env:GOPROXY = "https://goproxy.cn,direct"
$out = "$distDir/jt-simulate-linux-arm64"
Write-Host "  Building linux/arm64..." -ForegroundColor DarkGray
$prevEAP = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
& go build -ldflags="$ldflags" -o $out . 2>&1 | Out-Null
$buildExit = $LASTEXITCODE; $ErrorActionPreference = $prevEAP
if ($buildExit -ne 0 -or -not (Test-Path $out)) { Write-Host "ERROR: Build failed for linux-arm64 (exit $buildExit)" -ForegroundColor Red; exit 1 }
$binaries["linux-arm64"] = $out
Write-Host "    -> jt-simulate-linux-arm64 ($([math]::Round((Get-Item $out).Length / 1MB, 1)) MB)" -ForegroundColor Yellow

# --- macOS amd64 (Web 服务模式) ---
$env:GOOS = "darwin"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"; $env:GOPROXY = "https://goproxy.cn,direct"
$out = "$distDir/jt-simulate-darwin-amd64"
Write-Host "  Building darwin/amd64..." -ForegroundColor DarkGray
$prevEAP = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
& go build -ldflags="$ldflags" -o $out . 2>&1 | Out-Null
$buildExit = $LASTEXITCODE; $ErrorActionPreference = $prevEAP
if ($buildExit -ne 0 -or -not (Test-Path $out)) { Write-Host "ERROR: Build failed for darwin-amd64 (exit $buildExit)" -ForegroundColor Red; exit 1 }
$binaries["darwin-amd64"] = $out
Write-Host "    -> jt-simulate-darwin-amd64 ($([math]::Round((Get-Item $out).Length / 1MB, 1)) MB)" -ForegroundColor Yellow

# --- macOS arm64 (Web 服务模式) ---
$env:GOOS = "darwin"; $env:GOARCH = "arm64"; $env:CGO_ENABLED = "0"; $env:GOPROXY = "https://goproxy.cn,direct"
$out = "$distDir/jt-simulate-darwin-arm64"
Write-Host "  Building darwin/arm64..." -ForegroundColor DarkGray
$prevEAP = $ErrorActionPreference; $ErrorActionPreference = 'Continue'
& go build -ldflags="$ldflags" -o $out . 2>&1 | Out-Null
$buildExit = $LASTEXITCODE; $ErrorActionPreference = $prevEAP
if ($buildExit -ne 0 -or -not (Test-Path $out)) { Write-Host "ERROR: Build failed for darwin-arm64 (exit $buildExit)" -ForegroundColor Red; exit 1 }
$binaries["darwin-arm64"] = $out
Write-Host "    -> jt-simulate-darwin-arm64 ($([math]::Round((Get-Item $out).Length / 1MB, 1)) MB)" -ForegroundColor Yellow

# 清理交叉编译环境变量
$env:GOOS = ""; $env:GOARCH = ""; $env:CGO_ENABLED = ""
Write-Host "  All binaries built OK" -ForegroundColor DarkGray

# ============================================================
# 2. 打包发布 zip
# ============================================================
Write-Host "[2/3] Preparing release packages..." -ForegroundColor Green

$releaseDir = "release"
New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null

function Package-Zip($tag, $binKey, $isWinHost) {
    $ext = if ($isWinHost) { ".exe" } else { "" }
    $pkgDir = "release/jt-simulate-$tag-pkg"
    New-Item -ItemType Directory -Path $pkgDir -Force | Out-Null

    # 复制二进制
    Copy-Item $binaries[$binKey] "$pkgDir/jt-simulate$ext" -Force

    # 复制配置文件
    if (Test-Path "configs/config.yaml") {
        New-Item -ItemType Directory -Path "$pkgDir/configs" -Force | Out-Null
        Copy-Item "configs/config.yaml" "$pkgDir/configs/config.yaml" -Force
    }

    # 复制 README
    if (Test-Path "README.md") { Copy-Item "README.md" "$pkgDir/README.md" -Force }

    # 打 zip 包
    $zipball = "release/jt-simulate-$Version-$tag.zip"
    Write-Host "  Creating $zipball..." -ForegroundColor DarkGray
    Compress-Archive -Path "$pkgDir/*" -DestinationPath $zipball -Force

    $size = [math]::Round((Get-Item $zipball).Length / 1MB, 1)
    Write-Host "    -> $zipball ($size MB)" -ForegroundColor Yellow
}

Package-Zip -tag "windows-amd64" -binKey "windows-amd64" -isWinHost $true
Package-Zip -tag "linux-amd64" -binKey "linux-amd64" -isWinHost $false
Package-Zip -tag "linux-arm64" -binKey "linux-arm64" -isWinHost $false
Package-Zip -tag "darwin-amd64" -binKey "darwin-amd64" -isWinHost $false
Package-Zip -tag "darwin-arm64" -binKey "darwin-arm64" -isWinHost $false

# ============================================================
# 3. 生成 SHA256 校验文件
# ============================================================
Write-Host "[3/3] Generating checksums..." -ForegroundColor Green
Push-Location $releaseDir
$hashes = Get-ChildItem -Filter "*.zip" | ForEach-Object {
    $hash = (Get-FileHash $_.Name -Algorithm SHA256).Hash
    "$hash  $($_.Name)"
}
$hashes | Out-File -Encoding ASCII -FilePath "checksums.txt"
Pop-Location

# ============================================================
# 输出结果
# ============================================================
Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "  Release packages created:" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Get-ChildItem $releaseDir -Filter "*.zip" | ForEach-Object {
    $size = if ($_.Length -gt 1MB) { "$([math]::Round($_.Length/1MB,1)) MB" } else { "$([math]::Round($_.Length/1KB,0)) KB" }
    Write-Host "  release/$($_.Name)  ($size)" -ForegroundColor White
}
Write-Host ""
Write-Host "Each package contains:" -ForegroundColor Cyan
Write-Host "  jt-simulate(.exe)       <- binary" -ForegroundColor DarkGray
Write-Host "  configs/config.yaml     <- default config (for serve mode)" -ForegroundColor DarkGray
Write-Host "  README.md               <- full documentation" -ForegroundColor DarkGray
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Yellow
Write-Host "  Windows: double-click jt-simulate.exe (desktop app)" -ForegroundColor Yellow
Write-Host "  Linux:   ./jt-simulate serve -c configs/config.yaml" -ForegroundColor Yellow
Write-Host "  macOS:   ./jt-simulate serve -c configs/config.yaml" -ForegroundColor Yellow
Write-Host ""
