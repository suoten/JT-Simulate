# JT-Simulate Docker 镜像构建
#
# 国内网络拉不动 Docker Hub 基础镜像时，加镜像源前缀构建：
#   docker build --build-arg IMAGE_MIRROR=docker.m.daocloud.io/ -t jt-simulate .
# （compose 用户：IMAGE_MIRROR=docker.m.daocloud.io/ docker compose up -d --build）
# 海外网络想用官方 Go 代理：
#   docker build --build-arg GOPROXY=https://proxy.golang.org,direct -t jt-simulate .
#
# IMAGE_MIRROR 是带尾斜杠的 registry 前缀，留空 = 直连 Docker Hub
ARG IMAGE_MIRROR=""

FROM ${IMAGE_MIRROR}golang:1.25-alpine AS builder

# Go 模块代理（默认国内友好；海外构建用 --build-arg GOPROXY=direct 覆盖）
ARG GOPROXY="https://goproxy.cn,direct"
ENV GOPROXY=${GOPROXY}

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o jt-simulate .

# ARG 作用域不跨阶段，第二阶段需重新声明
ARG IMAGE_MIRROR=""
FROM ${IMAGE_MIRROR}alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/jt-simulate .
COPY configs/ ./configs/
COPY data/ ./data/
EXPOSE 8095
CMD ["./jt-simulate", "serve"]
