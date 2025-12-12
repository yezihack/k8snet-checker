# Helm Repository 设置指南

本文档说明如何设置 GitHub Pages 来托管 Helm Chart 仓库。

## 方法一：使用自动化脚本（推荐）

### 快速设置

```bash
# 赋予执行权限
chmod +x scripts/quick-helm-setup.sh

# 运行脚本
./scripts/quick-helm-setup.sh
```

脚本会自动完成以下操作：

1. 打包 Helm Chart
2. 创建 gh-pages 分支（如果不存在）
3. 生成 index.yaml
4. 推送到 GitHub

### 完整设置（带交互）

```bash
# 赋予执行权限
chmod +x scripts/setup-helm-repo.sh

# 运行脚本
./scripts/setup-helm-repo.sh
```

## 方法二：手动设置

### 1. 打包 Helm Chart

```bash
# 创建临时目录
mkdir -p .helm-packages

# 打包 Chart
helm package chart/k8snet-checker -d .helm-packages
```

### 2. 创建 gh-pages 分支

```bash
# 创建并切换到 gh-pages 分支
git checkout --orphan gh-pages

# 清空工作目录
git rm -rf .

# 创建 README
cat > README.md << 'EOF'
# K8s Network Checker Helm Repository

## 使用方法

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```
EOF

# 提交
git add README.md
git commit -m "Initialize gh-pages"
git push -u origin gh-pages
```

### 3. 发布 Chart

```bash
# 确保在 gh-pages 分支
git checkout gh-pages

# 复制 Chart 包
cp .helm-packages/*.tgz .

# 生成 index.yaml
helm repo index . --url https://yezihack.github.io/k8snet-checker

# 提交并推送
git add *.tgz index.yaml
git commit -m "Release Helm Chart"
git push origin gh-pages

# 切换回主分支
git checkout main
```

## 方法三：使用 GitHub Actions（自动化）

项目已配置 GitHub Actions 自动发布，当 `chart/` 目录有变更时会自动触发。

查看配置文件：`.github/workflows/helm-pages.yml`

### 手动触发

在 GitHub 仓库页面：

1. 进入 Actions 标签
2. 选择 "Release Helm Charts to GitHub Pages"
3. 点击 "Run workflow"

## 启用 GitHub Pages

完成上述步骤后，需要在 GitHub 仓库设置中启用 GitHub Pages：

### 步骤

1. 进入 GitHub 仓库页面
2. 点击 **Settings**（设置）
3. 在左侧菜单找到 **Pages**
4. 在 **Source** 下拉菜单中选择 **gh-pages** 分支
5. 目录选择 **/ (root)**
6. 点击 **Save**（保存）

### 验证

等待几分钟后，访问：
```
https://yezihack.github.io/k8snet-checker
```

应该能看到 README 内容。

访问 index.yaml：
```
https://yezihack.github.io/k8snet-checker/index.yaml
```

应该能看到 Chart 索引信息。

## 使用 Helm 仓库

### 添加仓库

```bash
helm repo add k8snet-checker https://yezihack.github.io/k8snet-checker
helm repo update
```

### 搜索 Chart

```bash
# 搜索最新版本
helm search repo k8snet-checker

# 搜索所有版本
helm search repo k8snet-checker -l
```

### 安装 Chart

```bash
# 安装最新版本
helm install k8snet-checker k8snet-checker/k8snet-checker -n kube-system

# 安装指定版本
helm install k8snet-checker k8snet-checker/k8snet-checker \
  --version 0.1.0 \
  -n kube-system
```

### 查看 Chart 信息

```bash
# 查看 Chart 详情
helm show chart k8snet-checker/k8snet-checker

# 查看 Chart README
helm show readme k8snet-checker/k8snet-checker

# 查看默认配置
helm show values k8snet-checker/k8snet-checker
```

## 更新 Chart

### 修改 Chart 版本

编辑 `chart/k8snet-checker/Chart.yaml`：

```yaml
version: 0.2.0  # 更新版本号
appVersion: "0.2.0"  # 更新应用版本
```

### 发布新版本

#### 使用脚本

```bash
./scripts/quick-helm-setup.sh
```

#### 手动发布

```bash
# 切换到 gh-pages 分支
git checkout gh-pages

# 打包新版本
helm package ../chart/k8snet-checker -d .

# 更新索引（合并现有索引）
helm repo index . --url https://yezihack.github.io/k8snet-checker --merge index.yaml

# 提交并推送
git add *.tgz index.yaml
git commit -m "Release version 0.2.0"
git push origin gh-pages

# 切换回主分支
git checkout main
```

### 用户更新

```bash
# 更新仓库索引
helm repo update

# 查看新版本
helm search repo k8snet-checker -l

# 升级到新版本
helm upgrade k8snet-checker k8snet-checker/k8snet-checker -n kube-system
```

## 故障排查

### 问题 1: 404 Not Found

**症状**：
```
Error: looks like "https://yezihack.github.io/k8snet-checker" is not a valid chart repository
```

**解决方案**：
1. 确认 GitHub Pages 已启用
2. 确认 gh-pages 分支存在且有 index.yaml
3. 等待几分钟让 GitHub Pages 生效
4. 检查 URL 是否正确（不要包含 .git 后缀）

### 问题 2: index.yaml 不存在

**症状**：
```
failed to fetch https://yezihack.github.io/k8snet-checker/index.yaml : 404 Not Found
```

**解决方案**：
```bash
# 切换到 gh-pages 分支
git checkout gh-pages

# 检查是否有 index.yaml
ls -la index.yaml

# 如果不存在，生成它
helm repo index . --url https://yezihack.github.io/k8snet-checker

# 提交并推送
git add index.yaml
git commit -m "Add index.yaml"
git push origin gh-pages
```

### 问题 3: Chart 包不存在

**症状**：
```
Error: chart "k8snet-checker" not found in k8snet-checker index
```

**解决方案**：
```bash
# 确保 Chart 包已打包并推送
git checkout gh-pages
ls -la *.tgz

# 如果没有 .tgz 文件，重新打包
git checkout main
helm package chart/k8snet-checker -d /tmp
git checkout gh-pages
cp /tmp/*.tgz .
helm repo index . --url https://yezihack.github.io/k8snet-checker --merge index.yaml
git add *.tgz index.yaml
git commit -m "Add Chart package"
git push origin gh-pages
```

### 问题 4: GitHub Pages 未启用

**解决方案**：
1. 进入仓库 Settings > Pages
2. 确认 Source 设置为 gh-pages 分支
3. 确认显示绿色的成功消息
4. 等待部署完成（通常 1-5 分钟）

### 问题 5: 权限问题

**症状**：
```
Permission denied (publickey)
```

**解决方案**：
```bash
# 检查 Git 远程 URL
git remote -v

# 如果使用 SSH，确保 SSH 密钥已配置
# 或者改用 HTTPS
git remote set-url origin https://github.com/yezihack/k8snet-checker.git
```

## 最佳实践

### 1. 版本管理

- 遵循语义化版本（Semantic Versioning）
- Chart 版本和应用版本分开管理
- 每次发布都更新 CHANGELOG

### 2. 自动化

- 使用 GitHub Actions 自动发布
- 在 CI/CD 中集成 Helm 测试
- 自动更新文档

### 3. 文档

- 保持 Chart README 更新
- 提供详细的配置说明
- 包含使用示例

### 4. 测试

- 发布前测试 Chart 安装
- 验证升级路径
- 测试回滚功能

## 参考资料

- [Helm Chart Repository Guide](https://helm.sh/docs/topics/chart_repository/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Chart Releaser Action](https://github.com/helm/chart-releaser-action)

