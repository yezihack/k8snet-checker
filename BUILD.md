# 构建指南

本文档说明如何构建 K8s Network Checker 的 Docker 镜像和二进制文件。

## 目录

- [前置要求](#前置要求)
- [构建 Docker 镜像](#构建-docker-镜像)
- [构建本地二进制文件](#构建本地二进制文件)
- [多架构构建](#多架构构建)
- [构建优化](#构建优化)
- [故障排查](#故障排查)

## 前置要求

### 必需工具

- **Docker**: 20.10+ 或更高版本
- **Go**: 1.24.0 或更高版本（用于本地构建）
- **Git**: 用于克隆代码仓库

### 验证工具版本

```bash
# 检查 Docker 版本
docker --version

# 检查 Go 版本
go version

# 检查 Git 版本
git --version
```

## 构建 Docker 镜像

### 1. 克隆代码仓库

```bash
git clone <repository-url>
cd k8snet-checker
```

### 2. 构建服务器镜像

```bash
# 构建服务器镜像
docker build -t k8snet-checker-server:latest -f Dockerfile.server .

# 构建带版本标签的镜像
docker build -t k8snet-checker-server:v1.0.0 -f Dockerfile.server .
```

**构建参数说明**:
- `-t`: 指定镜像名称和标签
- `-f`: 指定 Dockerfile 文件
- `.`: 构建上下文路径

### 3. 构建客户端镜像

```bash
# 构建客户端镜像
docker build -t k8snet-checker-client:latest -f Dockerfile.client .

# 构建带版本标签的镜像
docker build -t k8snet-checker-client:v1.0.0 -f Dockerfile.client .

# 构建 Linux AMD64
docker build -f Dockerfile.server \
  --build-arg TARGETOS=linux \
  --build-arg TARGETARCH=amd64 \
  -t Dockerfile.client .

# 构建 Linux ARM64
docker build -f Dockerfile.server \
  --build-arg TARGETOS=linux \
  --build-arg TARGETARCH=arm64 \
  -t  Dockerfile.client .

docker buildx build -f Dockerfile.server \
  --platform linux/amd64,linux/arm64 \
  -t Dockerfile.client .
```

### 4. 验证镜像

```bash
# 查看构建的镜像
docker images | grep k8snet-checker

# 查看镜像详细信息
docker inspect k8snet-checker-server:latest
docker inspect k8snet-checker-client:latest

# 查看镜像大小
docker images k8snet-checker-server:latest --format "{{.Size}}"
docker images k8snet-checker-client:latest --format "{{.Size}}"
```

### 5. 测试镜像

```bash
# 测试服务器镜像
docker run --rm k8snet-checker-server:latest --help

# 测试客户端镜像
docker run --rm k8snet-checker-client:latest --help
```

## 构建本地二进制文件

### 1. 安装依赖

```bash
# 下载 Go 模块依赖
go mod download

# 整理依赖
go mod tidy
```

### 2. 构建服务器二进制文件

```bash
# 构建服务器（当前平台）
go build -o bin/server ./cmd/server

# 构建服务器（Linux AMD64）
GOOS=linux GOARCH=amd64 go build -o bin/server-linux-amd64 ./cmd/server

# 构建服务器（优化版本，减小文件大小）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o bin/server-linux-amd64 \
    ./cmd/server
```

### 3. 构建客户端二进制文件

```bash
# 构建客户端（当前平台）
go build -o bin/client ./cmd/client

# 构建客户端（Linux AMD64）
GOOS=linux GOARCH=amd64 go build -o bin/client-linux-amd64 ./cmd/client

# 构建客户端（优化版本，减小文件大小）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o bin/client-linux-amd64 \
    ./cmd/client
```

### 4. 验证二进制文件

```bash
# 查看二进制文件信息
ls -lh bin/

# 运行服务器（需要配置环境变量）
./bin/server

# 运行客户端（需要配置环境变量）
./bin/client
```

## 多架构构建

### 使用 Docker Buildx

Docker Buildx 支持构建多架构镜像。

#### 1. 启用 Buildx

```bash
# 创建新的 builder 实例
docker buildx create --name multiarch-builder --use

# 启动 builder
docker buildx inspect --bootstrap
```

#### 2. 构建多架构镜像

```bash
# 构建服务器镜像（AMD64 和 ARM64）
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t k8snet-checker-server:latest \
    -f Dockerfile.server \
    --push \
    .

# 构建客户端镜像（AMD64 和 ARM64）
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t k8snet-checker-client:latest \
    -f Dockerfile.client \
    --push \
    .
```

**注意**: `--push` 参数会将镜像推送到 Docker Registry。如果只想本地构建，使用 `--load` 参数（但只支持单一架构）。

#### 3. 本地多架构构建（不推送）

```bash
# 构建并加载到本地（仅支持当前架构）
docker buildx build \
    --platform linux/amd64 \
    -t k8snet-checker-server:latest \
    -f Dockerfile.server \
    --load \
    .
```

## 推送镜像到 Registry

### 推送到 Docker Hub

```bash
# 登录 Docker Hub
docker login

# 标记镜像
docker tag k8snet-checker-server:latest <username>/k8snet-checker-server:latest
docker tag k8snet-checker-client:latest <username>/k8snet-checker-client:latest

# 推送镜像
docker push <username>/k8snet-checker-server:latest
docker push <username>/k8snet-checker-client:latest
```

### 推送到私有 Registry

```bash
# 登录私有 Registry
docker login <registry-url>

# 标记镜像
docker tag k8snet-checker-server:latest <registry-url>/k8snet-checker-server:latest
docker tag k8snet-checker-client:latest <registry-url>/k8snet-checker-client:latest

# 推送镜像
docker push <registry-url>/k8snet-checker-server:latest
docker push <registry-url>/k8snet-checker-client:latest
```

### 推送到阿里云容器镜像服务

```bash
# 登录阿里云 Registry
docker login --username=<your-username> registry.cn-hangzhou.aliyuncs.com

# 标记镜像
docker tag k8snet-checker-server:latest registry.cn-hangzhou.aliyuncs.com/<namespace>/k8snet-checker-server:latest
docker tag k8snet-checker-client:latest registry.cn-hangzhou.aliyuncs.com/<namespace>/k8snet-checker-client:latest

# 推送镜像
docker push registry.cn-hangzhou.aliyuncs.com/<namespace>/k8snet-checker-server:latest
docker push registry.cn-hangzhou.aliyuncs.com/<namespace>/k8snet-checker-client:latest
```

## 构建优化

### 1. 使用构建缓存

Docker 会自动缓存构建层，加快后续构建速度。

```bash
# 使用缓存构建
docker build -t k8snet-checker-server:latest -f Dockerfile.server .

# 不使用缓存构建（强制重新构建）
docker build --no-cache -t k8snet-checker-server:latest -f Dockerfile.server .
```

### 2. 减小镜像大小

Dockerfile 已经使用了多阶段构建和优化技术：

- **多阶段构建**: 分离构建环境和运行环境
- **静态编译**: `CGO_ENABLED=0` 避免动态链接
- **去除调试信息**: `-ldflags="-w -s"` 减小二进制文件大小
- **最小化基础镜像**: 使用 `alpine:latest` 作为运行时镜像
- **非 root 用户**: 提高安全性

### 3. 并行构建

```bash
# 同时构建服务器和客户端镜像
docker build -t k8snet-checker-server:latest -f Dockerfile.server . &
docker build -t k8snet-checker-client:latest -f Dockerfile.client . &
wait
```

### 4. 使用 BuildKit

BuildKit 是 Docker 的新一代构建引擎，提供更好的性能和缓存。

```bash
# 启用 BuildKit
export DOCKER_BUILDKIT=1

# 构建镜像
docker build -t k8snet-checker-server:latest -f Dockerfile.server .
```

## 构建脚本

创建一个构建脚本 `build.sh` 简化构建流程：

```bash
#!/bin/bash
set -euo pipefail

# 颜色定义
readonly COLOR_RESET='\033[0m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[0;33m'
readonly COLOR_RED='\033[0;31m'

# 日志函数
log_info() {
    echo -e "${COLOR_GREEN}[INFO]${COLOR_RESET} $*"
}

log_warn() {
    echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET} $*"
}

log_error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $*"
}

# 默认参数
VERSION="${VERSION:-latest}"
REGISTRY="${REGISTRY:-}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
PUSH="${PUSH:-false}"
LOAD="${LOAD:-false}"

# 显示帮助信息
show_help() {
    cat << EOF
用法: $0 [选项]

选项:
    -v, --version VERSION       镜像版本标签 (默认: latest)
    -r, --registry REGISTRY     镜像仓库地址 (默认: 不指定)
    -p, --platforms PLATFORMS   构建平台 (默认: linux/amd64,linux/arm64)
    -P, --push                  构建后推送到仓库 (需要指定 -r)
    -L, --load                  加载到本地 Docker (仅支持单一架构)
    -h, --help                  显示此帮助信息

示例:
    # 构建双架构镜像
    $0 -v v1.0.0

    # 构建并推送到 Docker Hub
    $0 -v v1.0.0 -r myregistry -P

    # 仅构建 amd64
    $0 -p linux/amd64 -L

    # 构建双架构并推送到阿里云
    VERSION=v1.0.0 REGISTRY=registry.cn-hangzhou.aliyuncs.com/myspace ./build.sh -P
EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -p|--platforms)
                PLATFORMS="$2"
                shift 2
                ;;
            -P|--push)
                PUSH="true"
                shift
                ;;
            -L|--load)
                LOAD="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查 docker buildx 是否可用
check_buildx() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi

    if ! docker buildx version &> /dev/null; then
        log_warn "docker buildx 不可用，尝试启用..."
        if ! docker run --rm --privileged tonistiigi/binfmt --install all &> /dev/null; then
            log_error "无法启用 docker buildx，请确保已安装 Docker 19.03+"
            exit 1
        fi
    fi
}

# 构建服务器镜像
build_server() {
    local image_name="k8snet-checker-server:${VERSION}"
    local full_image_name="${image_name}"
    
    if [ -n "${REGISTRY}" ]; then
        full_image_name="${REGISTRY}/k8snet-checker-server:${VERSION}"
    fi

    log_info "构建服务器镜像: ${full_image_name}"
    log_info "平台: ${PLATFORMS}"

    local build_cmd="docker buildx build"
    build_cmd="${build_cmd} --platform ${PLATFORMS}"
    build_cmd="${build_cmd} -t ${full_image_name}"
    build_cmd="${build_cmd} -f Dockerfile.server"

    if [ "${PUSH}" = "true" ]; then
        build_cmd="${build_cmd} --push"
        log_info "构建后将推送到仓库"
    elif [ "${LOAD}" = "true" ]; then
        build_cmd="${build_cmd} --load"
        log_warn "使用 --load 仅支持单一架构"
    fi

    build_cmd="${build_cmd} ."

    if eval "${build_cmd}"; then
        log_info "服务器镜像构建完成"
    else
        log_error "服务器镜像构建失败"
        exit 1
    fi
}

# 构建客户端镜像
build_client() {
    local image_name="k8snet-checker-client:${VERSION}"
    local full_image_name="${image_name}"
    
    if [ -n "${REGISTRY}" ]; then
        full_image_name="${REGISTRY}/k8snet-checker-client:${VERSION}"
    fi

    log_info "构建客户端镜像: ${full_image_name}"
    log_info "平台: ${PLATFORMS}"

    local build_cmd="docker buildx build"
    build_cmd="${build_cmd} --platform ${PLATFORMS}"
    build_cmd="${build_cmd} -t ${full_image_name}"
    build_cmd="${build_cmd} -f Dockerfile.client"

    if [ "${PUSH}" = "true" ]; then
        build_cmd="${build_cmd} --push"
        log_info "构建后将推送到仓库"
    elif [ "${LOAD}" = "true" ]; then
        build_cmd="${build_cmd} --load"
        log_warn "使用 --load 仅支持单一架构"
    fi

    build_cmd="${build_cmd} ."

    if eval "${build_cmd}"; then
        log_info "客户端镜像构建完成"
    else
        log_error "客户端镜像构建失败"
        exit 1
    fi
}

# 验证参数
validate_params() {
    if [ "${PUSH}" = "true" ] && [ -z "${REGISTRY}" ]; then
        log_error "使用 --push 必须指定 --registry"
        exit 1
    fi

    if [ "${LOAD}" = "true" ] && [ "${PUSH}" = "true" ]; then
        log_error "不能同时使用 --load 和 --push"
        exit 1
    fi

    if [ "${LOAD}" = "true" ] && [ "${PLATFORMS}" != "linux/amd64" ] && [ "${PLATFORMS}" != "linux/arm64" ]; then
        log_error "--load 仅支持单一架构，当前: ${PLATFORMS}"
        exit 1
    fi
}

# 主函数
main() {
    log_info "=========================================="
    log_info "K8s Network Checker 构建脚本"
    log_info "=========================================="
    log_info "版本: ${VERSION}"
    log_info "仓库: ${REGISTRY:-(本地)}"
    log_info "平台: ${PLATFORMS}"
    log_info "推送: ${PUSH}"
    log_info "加载: ${LOAD}"
    log_info "=========================================="

    check_buildx
    validate_params

    build_server
    build_client

    log_info "=========================================="
    log_info "构建完成！"
    log_info "=========================================="

    if [ "${PUSH}" = "true" ]; then
        log_info "镜像已推送到 ${REGISTRY}"
    elif [ "${LOAD}" = "true" ]; then
        log_info "镜像已加载到本地 Docker"
        docker images | grep k8snet-checker || true
    else
        log_info "构建完成，镜像已构建但未推送"
        log_info "查看镜像: docker buildx du"
    fi
}

# 解析参数并执行主函数
parse_args "$@"
main
```

使用构建脚本：

```bash
# 赋予执行权限
chmod +x build.sh

# 构建双架构镜像（本地缓存）
./build.sh -v v1.0.0

# 构建并推送到 Docker Hub
./build.sh -v v1.0.0 -r myusername -P

# 构建并推送到阿里云
./build.sh -v v1.0.0 -r registry.cn-hangzhou.aliyuncs.com/myspace -P

# 构建并推送到私有仓库
./build.sh -v v1.0.0 -r my-registry.example.com -P

# 仅构建 AMD64 架构并加载到本地
./build.sh -v v1.0.0 -p linux/amd64 -L

# 仅构建 ARM64 架构并加载到本地
./build.sh -v v1.0.0 -p linux/arm64 -L

# 使用环境变量
VERSION=v1.0.0 REGISTRY=sgfoot PUSH=true ./build.sh

# 显示帮助
./build.sh -h
```

**脚本特性**:
- ✅ 支持双架构构建（amd64 和 arm64）
- ✅ 支持自定义架构组合
- ✅ 支持本地加载（单架构）
- ✅ 支持推送到任意镜像仓库
- ✅ 彩色输出和详细日志
- ✅ 参数验证和错误处理

## 故障排查

### 问题 1: 构建失败 - 依赖下载超时

**错误信息**:
```
go: downloading github.com/gin-gonic/gin v1.9.1
timeout
```

**解决方案**:
```bash
# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct

# 或者在 Dockerfile 中添加
ENV GOPROXY=https://goproxy.cn,direct
```

### 问题 2: 镜像体积过大

**问题**: 构建的镜像超过 100MB

**解决方案**:
- 确保使用了多阶段构建
- 检查是否使用了 `-ldflags="-w -s"` 参数
- 使用 `alpine` 而不是 `ubuntu` 作为基础镜像
- 清理不必要的文件

```bash
# 查看镜像层信息
docker history k8snet-checker-server:latest

# 分析镜像大小
docker images k8snet-checker-server:latest
```

### 问题 3: 权限错误

**错误信息**:
```
permission denied while trying to connect to the Docker daemon socket
```

**解决方案**:
```bash
# 将当前用户添加到 docker 组
sudo usermod -aG docker $USER

# 重新登录或执行
newgrp docker

# 或者使用 sudo
sudo docker build -t k8snet-checker-server:latest -f Dockerfile.server .
```

### 问题 4: 构建缓存问题

**问题**: 代码更新后构建仍使用旧版本

**解决方案**:

```bash
# 清除构建缓存
docker builder prune

# 强制重新构建
docker build --no-cache -t k8snet-checker-server:latest -f Dockerfile.server .
```

### 问题 5: 多架构构建失败

**错误信息**:
```
multiple platforms feature is currently not supported for docker driver
```

**解决方案**:
```bash
# 创建并使用新的 builder
docker buildx create --name multiarch-builder --use
docker buildx inspect --bootstrap

# 然后重新构建
docker buildx build --platform linux/amd64,linux/arm64 ...
```

## 最佳实践

1. **版本管理**: 始终为镜像打上版本标签，不要只使用 `latest`
2. **构建缓存**: 合理利用 Docker 构建缓存，加快构建速度
3. **安全扫描**: 定期扫描镜像漏洞
4. **镜像大小**: 保持镜像尽可能小，提高部署速度
5. **多阶段构建**: 分离构建环境和运行环境
6. **自动化**: 使用 CI/CD 自动化构建和推送流程

## GitHub Actions 自动化构建

项目已配置 GitHub Actions 工作流，可自动构建多架构镜像并推送到 Docker Hub。

### 配置步骤

1. **创建 Docker Hub Access Token**
   - 登录 [Docker Hub](https://hub.docker.com/)
   - 进入 Account Settings -> Security
   - 创建新的 Access Token

2. **配置 GitHub Secrets**
   - 进入 GitHub 仓库 Settings -> Secrets and variables -> Actions
   - 添加以下 Secrets:
     - `DOCKER_HUB_USERNAME`: Docker Hub 用户名
     - `DOCKER_HUB_TOKEN`: Docker Hub Access Token

3. **触发构建**
   - 推送到 `main` 或 `develop` 分支
   - 创建版本标签（如 `v1.0.0`）
   - 手动触发（Actions 页面）

详细配置说明请参考：[.github/DOCKER_HUB_SETUP.md](.github/DOCKER_HUB_SETUP.md)

### 工作流特性

- ✅ 自动构建 amd64 和 arm64 架构
- ✅ 自动标签管理（版本号、分支名、latest）
- ✅ 构建缓存优化
- ✅ PR 安全检查（仅构建不推送）
- ✅ 手动触发支持

### 使用构建脚本

项目提供了便捷的构建脚本：

**Linux/macOS:**

```bash
# 赋予执行权限
chmod +x scripts/build-multiarch.sh

# 查看帮助
./scripts/build-multiarch.sh -h

# 构建并推送
./scripts/build-multiarch.sh -u your-username -t v1.0.0 -p
```

**Windows PowerShell:**

```powershell
# 查看帮助
Get-Help .\scripts\build-multiarch.ps1

# 构建并推送
.\scripts\build-multiarch.ps1 -Username your-username -Tag v1.0.0 -Push
```

## 相关文档

- [部署指南](DEPLOY.md)
- [测试指南](TESTING.md)
- [README](README.md)
- [Docker Hub 配置指南](.github/DOCKER_HUB_SETUP.md)
