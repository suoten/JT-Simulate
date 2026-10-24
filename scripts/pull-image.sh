#!/usr/bin/env bash
# pull-image.sh — Docker 镜像多源自动拉取（国内加速器失效时的救急工具）
#
# 直接拉原名，失败则依次尝试镜像源前缀拉取并 tag 回原名。
# 用法：
#   ./scripts/pull-image.sh minio/minio:RELEASE.2024-07-16T15-48-33Z
#   ./scripts/pull-image.sh nginx:1.27-alpine
#   ./scripts/pull-image.sh mysql:8.0 --mirror docker.xuanyuan.me   # 追加自定义源
#
# 自定义默认源列表（逗号分隔，优先级高于内置）：
#   export JT_PULL_MIRRORS="docker.m.daocloud.io,docker.1ms.run"
#
# 退出码：0 成功；1 全部源失败。

set -u

DEFAULT_MIRRORS="docker.m.daocloud.io,docker.1ms.run"

usage() {
  sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
  exit 1
}

[ $# -ge 1 ] || usage
IMAGE=""
EXTRA_MIRRORS=""
while [ $# -gt 0 ]; do
  case "$1" in
    --mirror) shift; EXTRA_MIRRORS="$EXTRA_MIRRORS $1" ;;
    -h|--help) usage ;;
    *) IMAGE="$1" ;;
  esac
  shift
done
[ -n "$IMAGE" ] || usage

command -v docker >/dev/null 2>&1 || { echo "错误：未找到 docker 命令"; exit 1; }

MIRRORS="${JT_PULL_MIRRORS:-$DEFAULT_MIRRORS} $EXTRA_MIRRORS"

# Docker Hub 官方镜像（无仓库名，如 nginx:1.27）在镜像站需要 library/ 前缀
normalise() {
  local img="$1"
  case "$img" in
    */*) echo "$img" ;;
    *)   echo "library/$img" ;;
  esac
}

echo "==> [1/2] 直接拉取 $IMAGE（若本机加速器可用，这一步就完成了）"
if docker pull "$IMAGE" >/dev/null 2>&1; then
  echo "✓ 已拉取 $IMAGE"
  exit 0
fi
echo "  直接拉取失败，开始尝试镜像源..."

TARGET=$(normalise "$IMAGE")
echo "==> [2/2] 依次尝试镜像源：$MIRRORS"
for m in $MIRRORS; do
  m="${m%/}"          # 容忍带尾斜杠的输入
  FULL="$m/$TARGET"
  echo "  docker pull $FULL ..."
  if docker pull "$FULL" >/dev/null 2>&1; then
    if docker tag "$FULL" "$IMAGE"; then
      echo "✓ 已通过 $m 拉取，并 tag 回 $IMAGE"
      exit 0
    fi
  fi
done

cat <<'EOF'

✗ 所有镜像源均失败。请按 README「常见问题」的 Docker 镜像排障步骤处理：
   1. docker info 查看 Registry Mirrors 是否配置了失效的加速器，换成可用的；
   2. 报 failed size validation → 本地缓存了坏 blob：
      Docker Desktop → Troubleshoot → Clean / Purge data 后重拉；
   3. 检查系统/公司代理是否篡改了 registry 响应（临时关闭代理直连试试）。
EOF
exit 1
