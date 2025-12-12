#!/bin/bash
set -euo pipefail

# 快速设置 Helm 仓库脚本

echo "🚀 开始设置 Helm Chart 仓库..."

# 1. 打包 Chart
echo "📦 打包 Helm Chart..."
mkdir -p .helm-packages
helm package chart/k8snet-checker -d .helm-packages

# 2. 创建 gh-pages 分支（如果不存在）
if ! git show-ref --verify --quiet refs/heads/gh-pages; then
    echo "📝 创建 gh-pages 分支..."
    git checkout --orphan gh-pages
    git rm -rf .
    
    # 创建 README
    cat > README.md << 'EOF'
# K8s Network Checker Helm Repository

## 添加仓库

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
```

## 安装

```bash
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

EOF
    
    git add README.md
    git commit -m "Initialize gh-pages"
    git push -u origin gh-pages
    git checkout main
fi

# 3. 切换到 gh-pages 分支
echo "🔄 切换到 gh-pages 分支..."
git checkout gh-pages

# 4. 复制 Chart 包
echo "📋 复制 Chart 包..."
cp .helm-packages/*.tgz .

# 5. 生成 index.yaml
echo "📄 生成 index.yaml..."
helm repo index . --url https://yezihack.github.io/k8snet-checker --merge index.yaml 2>/dev/null || \
helm repo index . --url https://yezihack.github.io/k8snet-checker

# 6. 提交并推送
echo "⬆️  提交并推送..."
git add *.tgz index.yaml
git commit -m "Update Helm repository $(date +'%Y-%m-%d %H:%M:%S')"
git push origin gh-pages

# 7. 切换回主分支
echo "🔙 切换回主分支..."
git checkout main

# 8. 清理临时文件
rm -rf .helm-packages

echo ""
echo "✅ Helm 仓库设置完成！"
echo ""
echo "📌 下一步操作："
echo "1. 在 GitHub 仓库设置中启用 GitHub Pages (Settings > Pages > Source: gh-pages)"
echo "2. 等待几分钟后使用："
echo "   helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker"
echo "   helm repo update"
echo "   helm search repo k8snet-checker"

