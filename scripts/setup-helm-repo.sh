#!/bin/bash
set -euo pipefail

# 脚本说明：初始化 Helm Chart 仓库到 GitHub Pages
# 作者：sgfoot
# 日期：2025-12-12

# 颜色定义
readonly COLOR_RESET='\033[0m'
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[0;33m'
readonly COLOR_BLUE='\033[0;34m'

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

# 检查必需的命令
function require_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "Required command not found: $1"
        exit 1
    fi
}

log_info "检查必需的命令..."
require_command helm
require_command git

# 获取当前分支
CURRENT_BRANCH=$(git branch --show-current)
log_info "当前分支: $CURRENT_BRANCH"

# 确保在主分支
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" ]]; then
    log_warn "建议在 main 或 master 分支执行此脚本"
    read -r -p "是否继续? [y/N] " response
    case "${response}" in
        [yY][eE][sS]|[yY]) 
            log_info "继续执行..."
            ;;
        *)
            log_info "已取消"
            exit 0
            ;;
    esac
fi

# 创建临时目录
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "${TEMP_DIR}"' EXIT

log_info "创建临时目录: $TEMP_DIR"

# 打包 Helm Chart
log_info "打包 Helm Chart..."
helm package chart/k8snet-checker -d "${TEMP_DIR}"

# 获取打包的文件名
CHART_PACKAGE=$(ls "${TEMP_DIR}"/*.tgz)
CHART_NAME=$(basename "${CHART_PACKAGE}")
log_info "Chart 包: $CHART_NAME"

# 检查 gh-pages 分支是否存在
if git show-ref --verify --quiet refs/heads/gh-pages; then
    log_info "gh-pages 分支已存在"
    BRANCH_EXISTS=true
else
    log_info "gh-pages 分支不存在，将创建新分支"
    BRANCH_EXISTS=false
fi

# 切换到 gh-pages 分支
if [[ "$BRANCH_EXISTS" == "true" ]]; then
    log_info "切换到 gh-pages 分支..."
    git checkout gh-pages
else
    log_info "创建 gh-pages 分支..."
    git checkout --orphan gh-pages
    git rm -rf .
    
    # 创建 README
    cat > README.md << 'EOF'
# K8s Network Checker Helm Repository

这是 K8s Network Checker 的 Helm Chart 仓库。

## 使用方法

### 添加仓库

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
```

### 安装

```bash
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

### 查看可用版本

```bash
helm search repo k8snet-checker -l
```

## 文档

- [项目主页](https://github.com/yezihack/k8snet-checker)
- [Chart 文档](https://github.com/yezihack/k8snet-checker/tree/main/chart/k8snet-checker)

EOF
    
    git add README.md
    git commit -m "Initial commit for gh-pages"
fi

# 复制 Chart 包到当前目录
log_info "复制 Chart 包..."
cp "${CHART_PACKAGE}" .

# 生成或更新 index.yaml
log_info "生成 index.yaml..."
if [[ -f "index.yaml" ]]; then
    helm repo index . --url https://yezihack.github.io/k8snet-checker --merge index.yaml
else
    helm repo index . --url https://yezihack.github.io/k8snet-checker
fi

# 提交更改
log_info "提交更改..."
git add "${CHART_NAME}" index.yaml
git commit -m "Release ${CHART_NAME}"

# 推送到远程
log_info "推送到远程仓库..."
if git push origin gh-pages; then
    log_success "成功推送到 gh-pages 分支"
else
    log_error "推送失败，请检查权限"
    exit 1
fi

# 切换回原分支
log_info "切换回 $CURRENT_BRANCH 分支..."
git checkout "$CURRENT_BRANCH"

log_success "Helm 仓库设置完成！"
echo ""
log_info "接下来的步骤："
echo "1. 在 GitHub 仓库设置中启用 GitHub Pages"
echo "   - 进入 Settings > Pages"
echo "   - Source 选择 'gh-pages' 分支"
echo "   - 保存设置"
echo ""
echo "2. 等待几分钟后，使用以下命令添加仓库："
echo "   helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker"
echo "   helm repo update"
echo ""
echo "3. 安装 Chart："
echo "   helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system"

