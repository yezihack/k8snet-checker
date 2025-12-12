#!/bin/bash
set -euo pipefail

# 脚本说明：手动同步 README 到 Docker Hub
# 作者：sgfoot
# 日期：2025-12-12

# 颜色定义
readonly COLOR_RESET='\033[0m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_RED='\033[0;31m'
readonly COLOR_YELLOW='\033[0;33m'

function log_info() {
    echo -e "${COLOR_GREEN}[INFO]${COLOR_RESET} $*"
}

function log_error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $*" >&2
}

function log_warn() {
    echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET} $*" >&2
}

function log_success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $*"
}

# 检查必需的环境变量
if [[ -z "${DOCKER_HUB_USERNAME:-}" ]]; then
    log_error "DOCKER_HUB_USERNAME 环境变量未设置"
    echo "请设置: export DOCKER_HUB_USERNAME=your_username"
    exit 1
fi

if [[ -z "${DOCKER_HUB_TOKEN:-}" ]]; then
    log_error "DOCKER_HUB_TOKEN 环境变量未设置"
    echo "请设置: export DOCKER_HUB_TOKEN=your_token"
    echo "Token 可以在 Docker Hub > Account Settings > Security 中创建"
    exit 1
fi

# 检查 README 文件是否存在
if [[ ! -f "docker/README-server.md" ]]; then
    log_error "docker/README-server.md 文件不存在"
    exit 1
fi

if [[ ! -f "docker/README-client.md" ]]; then
    log_error "docker/README-client.md 文件不存在"
    exit 1
fi

log_info "开始同步 Docker Hub README..."

# 同步 Server README
log_info "同步 Server README..."
SERVER_REPO="${DOCKER_HUB_USERNAME}/k8snet-checker-server"
SERVER_README=$(cat docker/README-server.md)

# 获取 Docker Hub Token
log_info "获取 Docker Hub Token..."
TOKEN=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${DOCKER_HUB_USERNAME}\", \"password\": \"${DOCKER_HUB_TOKEN}\"}" \
    https://hub.docker.com/v2/users/login/ | jq -r .token)

if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    log_error "获取 Docker Hub Token 失败"
    exit 1
fi

# 更新 Server README
log_info "更新 Server README..."
RESPONSE=$(curl -s -X PATCH \
    -H "Authorization: JWT ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"full_description\": $(jq -Rs . < docker/README-server.md)}" \
    "https://hub.docker.com/v2/repositories/${SERVER_REPO}/")

if echo "$RESPONSE" | jq -e '.name' > /dev/null 2>&1; then
    log_success "Server README 同步成功"
else
    log_error "Server README 同步失败"
    echo "$RESPONSE"
fi

# 同步 Client README
log_info "同步 Client README..."
CLIENT_REPO="${DOCKER_HUB_USERNAME}/k8snet-checker-client"

log_info "更新 Client README..."
RESPONSE=$(curl -s -X PATCH \
    -H "Authorization: JWT ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"full_description\": $(jq -Rs . < docker/README-client.md)}" \
    "https://hub.docker.com/v2/repositories/${CLIENT_REPO}/")

if echo "$RESPONSE" | jq -e '.name' > /dev/null 2>&1; then
    log_success "Client README 同步成功"
else
    log_error "Client README 同步失败"
    echo "$RESPONSE"
fi

log_success "所有 README 同步完成！"
echo ""
log_info "验证："
echo "  Server: https://hub.docker.com/r/${SERVER_REPO}"
echo "  Client: https://hub.docker.com/r/${CLIENT_REPO}"

