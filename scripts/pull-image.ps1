# pull-image.ps1 — Docker 镜像多源自动拉取（国内加速器失效时的救急工具，Windows 版）
#
# 直接拉原名，失败则依次尝试镜像源前缀拉取并 tag 回原名。
# 用法：
#   .\scripts\pull-image.ps1 minio/minio:RELEASE.2024-07-16T15-48-33Z
#   .\scripts\pull-image.ps1 nginx:1.27-alpine
#   .\scripts\pull-image.ps1 mysql:8.0 -Mirror docker.xuanyuan.me   # 追加自定义源
#
# 自定义默认源列表（逗号分隔，优先级高于内置）：
#   $env:JT_PULL_MIRRORS = "docker.m.daocloud.io,docker.1ms.run"

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Image,

    [Parameter()]
    [string[]]$Mirror = @()
)

$ErrorActionPreference = "Stop"
$DefaultMirrors = @("docker.m.daocloud.io", "docker.1ms.run")

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "错误：未找到 docker 命令" -ForegroundColor Red
    exit 1
}

$mirrors = @()
if ($env:JT_PULL_MIRRORS) {
    $mirrors += ($env:JT_PULL_MIRRORS -split ',')
} else {
    $mirrors += $DefaultMirrors
}
$mirrors += $Mirror

# Docker Hub 官方镜像（无仓库名，如 nginx:1.27）在镜像站需要 library/ 前缀
$target = if ($Image -match '/') { $Image } else { "library/$Image" }

Write-Host "==> [1/2] 直接拉取 $Image（若本机加速器可用，这一步就完成了）"
docker pull $Image 2>$null | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 已拉取 $Image" -ForegroundColor Green
    exit 0
}
Write-Host "  直接拉取失败，开始尝试镜像源..."

Write-Host "==> [2/2] 依次尝试镜像源：$($mirrors -join ', ')"
foreach ($m in $mirrors) {
    $m = $m.TrimEnd('/')
    $full = "$m/$target"
    Write-Host "  docker pull $full ..."
    docker pull $full 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) {
        docker tag $full $Image
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 已通过 $m 拉取，并 tag 回 $Image" -ForegroundColor Green
            exit 0
        }
    }
}

$msg = @"

✗ 所有镜像源均失败。请按 README「常见问题」的 Docker 镜像排障步骤处理：
   1. docker info 查看 Registry Mirrors 是否配置了失效的加速器，换成可用的；
   2. 报 failed size validation → 本地缓存了坏 blob：
      Docker Desktop → Troubleshoot → Clean / Purge data 后重拉；
   3. 检查系统/公司代理是否篡改了 registry 响应（临时关闭代理直连试试）。
"@
Write-Host $msg -ForegroundColor Red
exit 1
