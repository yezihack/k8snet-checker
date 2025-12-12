# Helm 仓库设置说明

## 🚨 当前状态

Helm Chart 仓库的 404 错误是因为 GitHub Pages 还没有设置。

## ✅ 立即可用的方法

### 方法 1: 从 GitHub Release 安装（推荐）

1. **创建一个 Release**
   - 在 GitHub 仓库页面点击 "Releases"
   - 点击 "Create a new release"
   - Tag 填写：`v0.1.0`
   - Title 填写：`v0.1.0`
   - 点击 "Publish release"
   - GitHub Actions 会自动打包 Helm Chart 并上传

2. **安装 Chart**
   ```bash
   helm install k8snet-checker \
     https://github.com/yezihack/k8snet-checker/releases/download/v0.1.0/k8snet-checker-0.1.0.tgz \
     -n kube-system --create-namespace
   ```

### 方法 2: 从本地安装

```bash
# 克隆仓库
git clone https://github.com/yezihack/k8snet-checker.git
cd k8snet-checker

# 安装
helm install k8snet-checker ./chart/k8snet-checker -n kube-system
```

## 🔧 设置 GitHub Pages Helm 仓库

如果你想让用户可以使用 `helm repo add` 命令，需要设置 GitHub Pages。

### 快速设置（3 步）

#### 1. 运行设置脚本

```bash
chmod +x scripts/quick-helm-setup.sh
./scripts/quick-helm-setup.sh
```

#### 2. 启用 GitHub Pages

1. 进入 GitHub 仓库页面
2. 点击 **Settings** > **Pages**
3. **Source** 选择 **gh-pages** 分支
4. 点击 **Save**

#### 3. 验证（等待 2-5 分钟）

```bash
# 测试 index.yaml
curl https://yezihack.github.io/k8snet-checker/index.yaml

# 添加仓库
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update

# 搜索 Chart
helm search repo k8snet-checker

# 安装
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

## 📚 详细文档

- [完整设置指南](docs/HELM_REPOSITORY_SETUP.md)
- [状态说明](docs/HELM_REPO_STATUS.md)
- [Chart 文档](chart/k8snet-checker/README.md)

## 💡 推荐方案

**对于用户**：
- 立即使用：从 Release 安装
- 长期使用：等待 GitHub Pages 设置完成

**对于维护者**：
1. 先创建 Release，让用户可以立即使用
2. 再设置 GitHub Pages，提供更好的体验

