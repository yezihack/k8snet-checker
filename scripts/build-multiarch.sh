#!/bin/bash
set -euo pipefail

# 脚本说明：构建多架构 Docker 镜像
# 作者：K8s Network Checker Team
# 日期：2025-11-28

# 颜色定义
readonly COLOR_RESET='\033[0m'
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[0;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_CYAN='\033[0;36m'

# 日志函数
function log_info() {
    echo -e "${COLOR_GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] [INFO]${COLOR_RESET} $*"
}

function log_error() {
    echo -e "${COLOR_RED}[$(date +'%Y-%m-%d %H:%M:%S')] [ERROR]${COLOR_RESET} $*" >&2
}

function log_warn() {
    echo -e "${COLOR_YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] [WARN]${COLOR_RESET} $*" >&2
}

function log_success() {
    echo -e "${COLOR_GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] [SUCCESS]${COLOR_RESET} $*"
}

# 使用说明
function usage() {
    cat << EOF
使用方法: $0 [OPTIONS]

选项:
    -u, --username      Docker Hub 用户名（必需）
    -t, --tag           镜像标签（默认：latest）
    -p, --push          推送到 Docker Hub
    -s, --server-only   仅构建 Server 镜像
    -c, --client-only   仅构建 Client 镜像
    -a, --arch          指定架构（默认：linux/amd64,linux/arm64）
    -h, --help          显示帮助信息

示例:
    # 构建并推送到 Docker Hub
    $0 -u myusername -t v1.0.0 -p

    # 仅构建 Server 镜像（不推送）
    $0 -u myusername -s

    # 构建单架构镜像
    $0 -u myusername -a linux/amd64

    # 构建并推送特定标签
    $0 -u myusername -t develop -p
EOF
    exit 1
}

# 检查命令是否存在
function require_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "未找到必需的命令: $1"
        exit 1
    fi
}

# 检查 Docker Buildx
function check_buildx() {
    if ! docker buildx version &> /dev/null; then
        log_error "Docker Buildx 未安装或未启用"
        log_info "请运行: docker buildx create --use"
        exit 1
    fi
}

# 创建 Buildx builder
function setup_builder() {
    local builder_name="multiarch-builder"
    
    if ! docker buildx inspect "${builder_name}" &> /dev/null; then
        log_info "创建 Buildx builder: ${builder_name}"
        docker buildx create --name "${builder_name}" --use
        docker buildx inspect --bootstrap
    else
        log_info "使用现有 Buildx builder: ${builder_name}"
        docker buildx use "${builder_name}"
    fi
}

# 构建镜像
function build_image() {
    local dockerfile=$1
    local image_name=$2
    local tag=$3
    local platforms=$4
    local push=$5
    
    local full_image="${image_name}:${tag}"
    
    log_info "开始构建镜像: ${full_image}"
    log_info "架构: ${platforms}"
    log_info "Dockerfile: ${dockerfile}"
    
    local build_args=(
        "buildx" "build"
        "--platform" "${platforms}"
        "-t" "${full_image}"
        "-f" "${dockerfile}"
    )
    
    if [[ "${push}" == "true" ]]; then
        build_args+=("--push")
        log_info "构建完成后将推送到 Docker Hub"
    else
        # 多架构构建不支持 --load，只能推送或不输出
        log_warn "多架构构建不会加载到本地 Docker"
        log_warn "如需本地测试，请使用单架构构建: -a linux/amd64"
    fi
    
    build_args+=(".")
    
    if docker "${build_args[@]}"; then
        log_success "镜像构建成功: ${full_image}"
        if [[ "${push}" == "true" ]]; then
            log_success "镜像已推送到 Docker Hub"
        fi
    else
        log_error "镜像构建失败: ${full_image}"
        exit 1
    fi
}

# 主函数
function main() {
    local username=""
    local tag="latest"
    local push="false"
    local build_server="true"
    local build_client="true"
    local platforms="linux/amd64,linux/arm64"
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--username)
                username="$2"
                shift 2
                ;;
            -t|--tag)
                tag="$2"
                shift 2
                ;;
            -p|--push)
                push="true"
                shift
                ;;
            -s|--server-only)
                build_client="false"
                shift
                ;;
            -c|--client-only)
                build_server="false"
                shift
                ;;
            -a|--arch)
                platforms="$2"
                shift 2
                ;;
            -h|--help)
                usage
                ;;
            *)
                log_error "未知选项: $1"
                usage
                ;;
        esac
    done
    
    # 验证必需参数
    if [[ -z "${username}" ]]; then
        log_error "错误: -u/--username 是必需的"
        usage
    fi
    
    # 检查依赖
    require_command docker
    check_buildx
    
    # 设置 builder
    setup_builder
    
    # 镜像名称
    local server_image="${username}/k8snet-checker-server"
    local client_image="${username}/k8snet-checker-client"
    
    log_info "=========================================="
    log_info "Docker 多架构镜像构建"
    log_info "=========================================="
    log_info "用户名: ${username}"
    log_info "标签: ${tag}"
    log_info "架构: ${platforms}"
    log_info "推送: ${push}"
    log_info "=========================================="
    
    # 构建 Server 镜像
    if [[ "${build_server}" == "true" ]]; then
        build_image "Dockerfile.server" "${server_image}" "${tag}" "${platforms}" "${push}"
    fi
    
    # 构建 Client 镜像
    if [[ "${build_client}" == "true" ]]; then
        build_image "Dockerfile.client" "${client_image}" "${tag}" "${platforms}" "${push}"
    fi
    
    log_success "=========================================="
    log_success "所有镜像构建完成！"
    log_success "=========================================="
    
    if [[ "${push}" == "true" ]]; then
        log_info "查看镜像信息:"
        if [[ "${build_server}" == "true" ]]; then
            log_info "  docker buildx imagetools inspect ${server_image}:${tag}"
        fi
        if [[ "${build_client}" == "true" ]]; then
            log_info "  docker buildx imagetools inspect ${client_image}:${tag}"
        fi
    else
        log_info "镜像未推送到 Docker Hub"
        log_info "如需推送，请添加 -p 或 --push 参数"
    fi
}

# 执行主函数
main "$@"
