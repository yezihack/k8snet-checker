# Helm Repository 状态说明

## 当前状态

Helm Chart 仓库尚未完全设置。你有以下几种方式使用 Helm Chart：

## ✅ 方法一：从 GitHub Release 安装（推荐，立即可用）

这是最简单的方法，不需要设置 Helm 仓库。

### 步骤

1. **创建 Release**

   在 GitHub 仓库页面创建一个 Release（例如 v0.1.0），GitHub Actions 会自动打包并上传 Helm Chart。

2. **安装 Chart**

   ```bash
   helm install k8snet-checker \
     https://github.com/yezihack/k8snet-checker/releases/download/v0.1.0/k8snet-checker-0.1.0.tgz \
     -n kube-system --create-namespace
   ```

### 优点

- ✅ 无需额外设置
- ✅ 立即可用
- ✅ 版本明确
- ✅ 适合快速测试

### 缺点

- ❌ 需要完整 URL
- ❌ 不支持 `helm search`
- ❌ 每次都要指定版本

## 🔧 方法二：设置 GitHub Pages Helm 仓库（推荐用于生产）

设置后可以使用 `helm repo add` 命令。

### 快速设置

```bash
# 1. 运行自动化脚本
chmod +x scripts/quick-helm-setup.sh
./scripts/quick-helm-setup.sh

# 2. 在 GitHub 启用 Pages
# 进入 Settings > Pages > Source: gh-pages

# 3. 等待几分钟后使用
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

详细步骤请参考：[HELM_REPOSITORY_SETUP.md](./HELM_REPOSITORY_SETUP.md)

### 优点

- ✅ 支持 `helm repo add`
- ✅ 支持 `helm search`
- ✅ 自动更新索引
- ✅ 专业的使用体验

### 缺点

- ❌ 需要初始设置
- ❌ 需要启用 GitHub Pages

## 📦 方法三：从本地安装（开发测试）

适合本地开发和测试。

```bash
# 克隆仓库
git clone https://github.com/yezihack/k8snet-checker.git
cd k8snet-checker

# 直接安装
helm install k8snet-checker ./chart/k8snet-checker -n kube-system

# 或者先打包
helm package chart/k8snet-checker
helm install k8snet-checker k8snet-checker-0.1.0.tgz -n kube-system
```

### 优点

- ✅ 适合开发测试
- ✅ 可以修改配置
- ✅ 无需网络

### 缺点

- ❌ 需要克隆仓库
- ❌ 不适合生产环境

## 🎯 推荐方案

### 对于用户

**立即使用**：
```bash
# 从 Release 安装（需要先创建 Release）
helm install k8snet-checker \
  https://github.com/yezihack/k8snet-checker/releases/download/v0.1.0/k8snet-checker-0.1.0.tgz \
  -n kube-system --create-namespace
```

**长期使用**：
```bash
# 等待 GitHub Pages 设置完成后
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

### 对于维护者

1. **立即可用**：创建 GitHub Release，让用户从 Release 安装
2. **长期规划**：设置 GitHub Pages，提供完整的 Helm 仓库体验

## 📝 设置 GitHub Pages 的步骤

### 1. 运行设置脚本

```bash
chmod +x scripts/quick-helm-setup.sh
./scripts/quick-helm-setup.sh
```

### 2. 启用 GitHub Pages

1. 进入 GitHub 仓库页面
2. 点击 **Settings**
3. 在左侧菜单找到 **Pages**
4. **Source** 选择 **gh-pages** 分支
5. 点击 **Save**

### 3. 验证

等待 2-5 分钟后：

```bash
# 测试 index.yaml 是否可访问
curl https://yezihack.github.io/k8snet-checker/index.yaml

# 添加仓库
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update

# 搜索 Chart
helm search repo k8snet-checker
```

## ❓ 常见问题

### Q: 为什么会出现 404 错误？

A: 可能的原因：
1. GitHub Pages 未启用
2. gh-pages 分支不存在
3. index.yaml 文件不存在
4. GitHub Pages 还在部署中（需要等待几分钟）

### Q: 我应该选择哪种方法？

A: 
- **快速测试**：从 Release 安装
- **生产使用**：设置 GitHub Pages
- **开发调试**：从本地安装

### Q: 如何更新 Chart？

A: 
1. 修改 Chart 版本号
2. 运行 `./scripts/quick-helm-setup.sh`
3. 或者创建新的 GitHub Release

### Q: 设置失败怎么办？

A: 查看详细的故障排查指南：[HELM_REPOSITORY_SETUP.md](./HELM_REPOSITORY_SETUP.md#故障排查)

## 📚 相关文档

- [Helm Repository 设置指南](./HELM_REPOSITORY_SETUP.md)
- [Chart README](../chart/k8snet-checker/README.md)
- [部署指南](../DEPLOY.md)

## 🔗 有用的链接

- [Helm 官方文档](https://helm.sh/docs/)
- [GitHub Pages 文档](https://docs.github.com/en/pages)
- [Chart Releaser Action](https://github.com/helm/chart-releaser-action)

