#!/bin/bash
set -euo pipefail

# 脚本说明：下载多架构镜像并生成推送脚本
# 作者：sgfoot
# 日期：2024-12-02

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

# 镜像列表
IMAGE_LIST=(
    "docker.io/sgfoot/k8snet-checker-client:v0.1.0"
    "docker.io/sgfoot/k8snet-checker-server:v0.1.0"
)

# 支持的架构
ARCHITECTURES=("amd64" "arm64")

# 镜像保存目录
readonly IMAGES_DIR="images"

# 检查必需命令
function check_requirements() {
    local missing_commands=()
    
    if ! command -v docker &> /dev/null; then
        missing_commands+=("docker")
    fi
    
    if [ ${#missing_commands[@]} -gt 0 ]; then
        log_error "缺少必需命令: ${missing_commands[*]}"
        log_info "请安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # 检查 Docker 是否运行
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行或无权限访问"
        exit 1
    fi
}

# 创建目录结构
function create_directories() {
    log_info "创建目录结构..."
    
    mkdir -p "$IMAGES_DIR"
    
    for arch in "${ARCHITECTURES[@]}"; do
        mkdir -p "$IMAGES_DIR/$arch"
    done
    
    log_success "目录结构创建完成"
}

# 提取镜像信息
function extract_image_info() {
    local image=$1
    local image_name
    local image_tag
    
    # 移除 docker.io/ 前缀
    image=${image#docker.io/}
    
    # 提取镜像名和标签
    image_name=$(echo "$image" | cut -d':' -f1 | tr '/' '_')
    image_tag=$(echo "$image" | cut -d':' -f2)
    
    echo "${image_name}_${image_tag}"
}

# 下载指定架构的镜像
function pull_image_arch() {
    local image=$1
    local arch=$2
    local filename=$3
    
    local tar_file="$IMAGES_DIR/$arch/$filename.tar"
    
    # 检查文件是否已存在
    if [[ -f "$tar_file" ]]; then
        log_info "镜像文件已存在，跳过下载: $tar_file [$arch]"
        return 0
    fi
    
    log_info "下载镜像: $image [$arch]"
    
    # 使用 Docker 拉取指定架构的镜像
    local platform="linux/$arch"
    
    if docker pull --platform "$platform" "$image"; then
        log_info "保存镜像到文件: $filename.tar"
        
        # 保存镜像到 tar 文件
        if docker save "$image" -o "$tar_file"; then
            log_success "镜像保存成功: $filename.tar [$arch]"
            
            # 删除本地镜像以节省空间
            docker rmi "$image" &> /dev/null || true
            
            return 0
        else
            log_error "镜像保存失败: $image [$arch]"
            return 1
        fi
    else
        log_error "镜像下载失败: $image [$arch]"
        return 1
    fi
}

# 下载所有镜像
function pull_all_images() {
    log_info "开始下载镜像..."
    
    local total_images=$((${#IMAGE_LIST[@]} * ${#ARCHITECTURES[@]}))
    local current=0
    local failed=0
    local skipped=0
    
    for image in "${IMAGE_LIST[@]}"; do
        local filename
        filename=$(extract_image_info "$image")
        
        for arch in "${ARCHITECTURES[@]}"; do
            ((current++))
            log_info "进度: [$current/$total_images]"
            
            local tar_file="$IMAGES_DIR/$arch/$filename.tar"
            if [[ -f "$tar_file" ]]; then
                ((skipped++))
            fi
            
            if ! pull_image_arch "$image" "$arch" "$filename"; then
                ((failed++))
            fi
            
            echo ""
        done
    done
    
    if [ $failed -gt 0 ]; then
        log_warn "下载完成，但有 $failed 个镜像失败"
    else
        log_success "所有镜像下载完成"
    fi
    
    if [ $skipped -gt 0 ]; then
        log_info "跳过已存在的镜像: $skipped 个"
    fi
}

# 生成加载和推送脚本
function generate_push_script() {
    log_info "生成推送脚本..."
    
    local script_file="$IMAGES_DIR/push-to-registry.sh"
    
    cat > "$script_file" << 'EOF'
#!/bin/bash
set -euo pipefail

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
EOF

    # 添加镜像列表
    for image in "${IMAGE_LIST[@]}"; do
        local image_without_registry=${image#docker.io/}
        local image_name=${image_without_registry%:*}
        local image_tag=${image_without_registry#*:}
        local short_name=${image_name#*/}
        local filename
        filename=$(extract_image_info "$image")
        
        echo "IMAGES[\"$short_name:$image_tag\"]=\"$filename\"" >> "$script_file"
    done

    # 添加架构列表和主要逻辑
    cat >> "$script_file" << 'EOF'

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
EOF

    chmod +x "$script_file"
    log_success "推送脚本生成完成: $script_file"
}

# 生成 README 文件
function generate_readme() {
    log_info "生成 README 文件..."
    
    local readme_file="$IMAGES_DIR/README.md"
    
    cat > "$readme_file" << 'EOF'
# 镜像文件说明

本目录包含下载的多架构 Docker 镜像文件。

## 目录结构

```
images/
├── amd64/                    # AMD64 架构镜像
│   ├── image1.tar
│   └── image2.tar
├── arm64/                    # ARM64 架构镜像
│   ├── image1.tar
│   └── image2.tar
├── push-to-registry.sh       # 推送脚本
└── README.md                 # 本文件
```

## 使用方法

### 1. 推送到自定义仓库

```bash
# 基本用法
./push-to-registry.sh --registry-host registry.example.com

# 指定命名空间
./push-to-registry.sh --registry-host registry.example.com --namespace myproject
```

### 2. 手动加载镜像

```bash
# 加载 AMD64 镜像
docker load -i amd64/image_name.tar

# 加载 ARM64 镜像
docker load -i arm64/image_name.tar
```

### 3. 查看镜像信息

```bash
# 查看镜像列表
docker images | grep k8snet-checker

# 查看镜像详细信息
docker inspect <image-id>
```

## 镜像列表

EOF

    # 添加镜像列表
    for image in "${IMAGE_LIST[@]}"; do
        echo "- \`$image\`" >> "$readme_file"
    done
    
    cat >> "$readme_file" << 'EOF'

## 架构支持

- linux/amd64 (x86_64)
- linux/arm64 (aarch64)

## 注意事项

1. 确保目标仓库已登录：`docker login registry.example.com`
2. 确保有足够的磁盘空间存储镜像文件
3. 推送前请确认目标仓库地址正确
4. 多架构镜像需要 Docker 19.03+ 版本支持

## 故障排查

### 加载镜像失败

```bash
# 检查文件完整性
ls -lh amd64/*.tar arm64/*.tar

# 检查 Docker 状态
docker info
```

### 推送镜像失败

```bash
# 检查登录状态
docker login registry.example.com

# 检查网络连接
ping registry.example.com
```

## 相关文档

- [Docker 官方文档](https://docs.docker.com/)
- [项目主页](https://github.com/yezihack/k8snet-checker)
EOF

    log_success "README 文件生成完成: $readme_file"
}

# 显示摘要信息
function show_summary() {
    log_info "=========================================="
    log_success "镜像下载完成"
    log_info "=========================================="
    
    echo ""
    log_info "目录结构："
    tree -L 2 "$IMAGES_DIR" 2>/dev/null || find "$IMAGES_DIR" -type f
    
    echo ""
    log_info "镜像文件大小："
    du -sh "$IMAGES_DIR"/* 2>/dev/null || true
    
    echo ""
    log_info "下一步操作："
    echo "  1. 查看 README: cat $IMAGES_DIR/README.md"
    echo "  2. 推送到仓库: cd $IMAGES_DIR && ./push-to-registry.sh --registry-host <your-registry>"
    echo ""
}

# 主函数
function main() {
    log_info "=========================================="
    log_info "多架构镜像下载工具"
    log_info "=========================================="
    
    check_requirements
    create_directories
    pull_all_images
    generate_push_script
    generate_readme
    show_summary
}

main "$@"
