.PHONY: build build-server build-desktop clean test run docker-build docker-up docker-pull

BINARY=jt-simulate
BUILD_DIR=bin

build: build-server

# 构建 Web/服务模式二进制
build-server:
	go build -o $(BUILD_DIR)/$(BINARY) .

# 构建 Windows 桌面应用（需要 Wails CLI）
build-desktop:
	wails build -platform windows/amd64

# 构建并运行
run: build-server
	$(BUILD_DIR)/$(BINARY) serve

# 测试
test:
	go test ./... -v

# 清理
clean:
	rm -rf $(BUILD_DIR) build dist

# 依赖
deps:
	go mod tidy

# Docker 镜像构建（国内网络加 IMAGE_MIRROR，如：make docker-build IMAGE_MIRROR=docker.m.daocloud.io/）
docker-build:
	docker build --build-arg IMAGE_MIRROR=$(or $(IMAGE_MIRROR),) -t jt-simulate .

# Docker 一键部署（构建并启动，访问 http://localhost:8095）
docker-up:
	IMAGE_MIRROR=$(or $(IMAGE_MIRROR),) docker compose up -d --build

# 镜像多源自动拉取（加速器失效救急）：make docker-pull IMAGE=minio/minio:tag
docker-pull:
	./scripts/pull-image.sh $(IMAGE)

# 跨平台构建
build-all: build-windows build-linux build-mac

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe .

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-linux-amd64 .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 .
