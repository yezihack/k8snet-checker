#!/bin/bash

# 脚本说明：加载镜像并推送到自定义仓库（修复版）
# 使用方法：./push-to-registry.sh --registry-host registry.example.com

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

function log_success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $*"
}

# 显示使用说明
function usage() {
    cat << USAGE
使用方法: $0 [选项]

选项:
    --registry-host HOST    目标镜像仓库地址（必需）
    --namespace NAMESPACE   镜像命名空间（可选，默认: sgfoot）
    --help                  显示此帮助信息

示例:
    $0 --registry-host registry.example.com
    $0 --registry-host registry.example.com --namespace myproject

USAGE
    exit 1
}

# 解析命令行参数
REGISTRY_HOST=""
NAMESPACE="sgfoot"

while [[ $# -gt 0 ]]; do
    case $1 in
        --registry-host)
            REGISTRY_HOST="$2"
            shift 2
            ;;
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --help)
            usage
            ;;
        *)
            log_error "未知选项: $1"
            usage
            ;;
    esac
done

# 验证必需参数
if [[ -z "$REGISTRY_HOST" ]]; then
    log_error "缺少必需参数: --registry-host"
    usage
fi

log_info "目标仓库: $REGISTRY_HOST"
log_info "命名空间: $NAMESPACE"

# 检查 Docker 是否运行
if ! docker info &> /dev/null; then
    log_error "Docker 未运行或无权限访问"
    exit 1
fi

# 镜像列表
declare -A IMAGES
IMAGES["k8snet-checker-client:v0.1.0"]="sgfoot_k8snet-checker-client_v0.1.0"
IMAGES["k8snet-checker-server:v0.1.0"]="sgfoot_k8snet-checker-server_v0.1.0"

ARCHITECTURES=("amd64" "arm64")

# 加载并推送单个架构的镜像
function load_and_push_arch() {
    local image_tag=$1
    local filename=$2
    local arch=$3
    
    local tar_file="$arch/${filename}.tar"
    
    if [[ ! -f "$tar_file" ]]; then
        log_error "镜像文件不存在: $tar_file"
        return 1
    fi
    
    log_info "加载镜像: $tar_file [$arch]"
    if ! docker load -i "$tar_file"; then
        log_error "镜像加载失败: $tar_file"
        return 1
    fi
    
    # 获取加载的镜像名称（可能是 sgfoot/xxx 或 docker.io/sgfoot/xxx）
    local source_image="sgfoot/$image_tag"
    
    # 检查镜像是否存在
    if ! docker image inspect "$source_image" &> /dev/null; then
        log_error "源镜像不存在: $source_image"
        return 1
    fi
    
    # 标记为目标镜像
    local target_image="$REGISTRY_HOST/$NAMESPACE/$image_tag-$arch"
    
    log_info "标记镜像: $source_image -> $target_image"
    if ! docker tag "$source_image" "$target_image"; then
        log_error "镜像标记失败"
        return 1
    fi
    
    # 推送镜像
    log_info "推送镜像: $target_image"
    if docker push "$target_image"; then
        log_success "镜像推送成功: $target_image"
        
        # 清理本地镜像
        docker rmi "$target_image" &> /dev/null || true
        docker rmi "$source_image" &> /dev/null || true
        
        return 0
    else
        log_error "镜像推送失败: $target_image"
        return 1
    fi
}

# 处理所有镜像
function process_all_images() {
    log_info "=========================================="
    log_info "开始处理镜像"
    log_info "=========================================="
    
    for image_tag in "${!IMAGES[@]}"; do
        local filename="${IMAGES[$image_tag]}"
        
        log_info "处理镜像: $image_tag"
        
        for arch in "${ARCHITECTURES[@]}"; do
            load_and_push_arch "$image_tag" "$filename" "$arch"
            echo ""
        done
    done
}

# 创建多架构 manifest
function create_manifests() {
    log_info "=========================================="
    log_info "创建多架构 manifest"
    log_info "=========================================="
    
    for image_tag in "${!IMAGES[@]}"; do
        local manifest_name="$REGISTRY_HOST/$NAMESPACE/$image_tag"
        
        log_info "创建 manifest: $manifest_name"
        
        # 删除已存在的 manifest（如果有）
        docker manifest rm "$manifest_name" 2>/dev/null || true
        
        # 创建新的 manifest
        local manifest_images=()
        for arch in "${ARCHITECTURES[@]}"; do
            manifest_images+=("$manifest_name-$arch")
        done
        
        if docker manifest create "$manifest_name" "${manifest_images[@]}"; then
            # 为每个架构添加注解
            for arch in "${ARCHITECTURES[@]}"; do
                docker manifest annotate "$manifest_name" "$manifest_name-$arch" \
                    --os linux --arch "$arch"
            done
            
            log_success "Manifest 创建成功"
            
            # 推送 manifest
            log_info "推送 manifest: $manifest_name"
            if docker manifest push "$manifest_name"; then
                log_success "Manifest 推送成功"
            else
                log_error "Manifest 推送失败"
            fi
        else
            log_error "Manifest 创建失败"
        fi
        
        echo ""
    done
}

# 主函数
function main() {
    process_all_images
    create_manifests
    
    log_info "=========================================="
    log_success "镜像推送完成"
    log_info "=========================================="
    
    echo ""
    log_info "推送的镜像列表："
    for image_tag in "${!IMAGES[@]}"; do
        echo "  - $REGISTRY_HOST/$NAMESPACE/$image_tag (multi-arch)"
        for arch in "${ARCHITECTURES[@]}"; do
            echo "    - $REGISTRY_HOST/$NAMESPACE/$image_tag-$arch"
        done
    done
}

main
