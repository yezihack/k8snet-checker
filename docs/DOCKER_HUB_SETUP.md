# Docker Hub README 设置指南

本文档说明如何为 Docker Hub 镜像添加 README 介绍。

## 方法一：使用 GitHub Actions（推荐）

### 自动同步

当 `docker/README-*.md` 文件更新时，GitHub Actions 会自动同步到 Docker Hub。

### 配置步骤

1. **创建 Docker Hub Access Token**
   - 登录 Docker Hub
   - 进入 Account Settings > Security
   - 点击 "New Access Token"
   - 名称填写：`github-actions`
   - 权限选择：`Read, Write, Delete`
   - 复制生成的 Token

2. **在 GitHub 设置 Secrets**
   - 进入 GitHub 仓库 Settings > Secrets and variables > Actions
   - 添加以下 Secrets：
     - `DOCKER_HUB_USERNAME`: 你的 Docker Hub 用户名
     - `DOCKER_HUB_TOKEN`: 刚才创建的 Access Token

3. **触发同步**
   - 修改 `docker/README-server.md` 或 `docker/README-client.md`
   - 提交并推送到 main 分支
   - GitHub Actions 会自动同步

### 手动触发

在 GitHub 仓库页面：
1. 进入 Actions 标签
2. 选择 "Sync Docker Hub README"
3. 点击 "Run workflow"

## 方法二：使用脚本手动同步

### 前置要求

- `curl` 命令
- `jq` 命令（JSON 处理工具）

安装 jq：
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### 使用步骤

1. **设置环境变量**

   ```bash
   export DOCKER_HUB_USERNAME=sgfoot
   export DOCKER_HUB_TOKEN=your_access_token
   ```

2. **运行同步脚本**

   ```bash
   chmod +x scripts/sync-docker-readme.sh
   ./scripts/sync-docker-readme.sh
   ```

## 方法三：手动在 Docker Hub 更新

### 步骤

1. 登录 Docker Hub
2. 进入仓库页面（如 `sgfoot/k8snet-checker-server`）
3. 点击 "Description" 标签
4. 点击 "Edit" 按钮
5. 复制 `docker/README-server.md` 的内容
6. 粘贴到编辑器中
7. 点击 "Update" 保存

对 Client 镜像重复相同步骤。

## README 文件说明

### Server README

文件位置：`docker/README-server.md`

包含内容：
- 镜像介绍和徽章
- 快速开始指南
- 环境变量说明
- API 端点列表
- 配置示例
- 故障排查
- 相关链接

### Client README

文件位置：`docker/README-client.md`

包含内容：
- 镜像介绍和徽章
- 快速开始指南
- 环境变量说明
- 功能特性
- 权限要求
- 配置示例
- 故障排查
- 相关链接

## 更新 README

### 修改内容

1. 编辑对应的 README 文件：
   - `docker/README-server.md`
   - `docker/README-client.md`

2. 提交更改：
   ```bash
   git add docker/README-*.md
   git commit -m "docs: update Docker Hub README"
   git push origin main
   ```

3. GitHub Actions 会自动同步到 Docker Hub

### 验证更新

访问 Docker Hub 仓库页面查看：
- Server: https://hub.docker.com/r/sgfoot/k8snet-checker-server
- Client: https://hub.docker.com/r/sgfoot/k8snet-checker-client

## 最佳实践

### README 内容建议

1. **简洁明了**
   - 开头用一句话说明镜像用途
   - 提供快速开始示例
   - 使用清晰的标题结构

2. **完整信息**
   - 环境变量说明
   - 端口说明
   - 资源要求
   - 配置示例

3. **实用性**
   - 提供可直接运行的命令
   - 包含故障排查指南
   - 链接到详细文档

4. **保持更新**
   - 版本号与实际一致
   - 及时更新新功能
   - 修正错误信息

### 徽章使用

README 中使用的徽章：
- 版本徽章：显示最新版本
- 大小徽章：显示镜像大小
- 下载徽章：显示下载次数

这些徽章会自动更新，无需手动维护。

## 故障排查

### 问题 1: GitHub Actions 同步失败

**症状**：
```
Error: unauthorized: authentication required
```

**解决方案**：
1. 检查 `DOCKER_HUB_USERNAME` 和 `DOCKER_HUB_TOKEN` Secrets 是否正确设置
2. 确认 Access Token 有 `Read, Write, Delete` 权限
3. 检查 Token 是否过期

### 问题 2: 手动脚本同步失败

**症状**：
```
获取 Docker Hub Token 失败
```

**解决方案**：
1. 检查环境变量是否正确设置
2. 确认用户名和密码正确
3. 检查网络连接

### 问题 3: README 格式显示异常

**症状**：
Markdown 格式在 Docker Hub 上显示不正确

**解决方案**：
1. Docker Hub 支持标准 Markdown 语法
2. 避免使用复杂的 HTML 标签
3. 测试链接是否有效
4. 图片使用绝对 URL

### 问题 4: 徽章不显示

**症状**：
徽章图片无法加载

**解决方案**：
1. 检查仓库名称是否正确
2. 确认仓库是公开的
3. 等待几分钟让缓存更新

## 相关资源

- [Docker Hub API 文档](https://docs.docker.com/docker-hub/api/latest/)
- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Markdown 语法指南](https://www.markdownguide.org/)
- [Shields.io 徽章生成](https://shields.io/)

## 示例 README 结构

```markdown
# 镜像名称

徽章区域

简短介绍（1-2 句话）

## 快速开始

### 在 Kubernetes 中部署
代码示例

### 单独运行
代码示例

## 环境变量

表格说明

## 功能特性

列表说明

## 配置示例

YAML 示例

## 故障排查

常见问题和解决方案

## 文档

相关链接

## 许可证

许可证信息

## 联系方式

联系方式和链接
```

## 自动化工作流

```
代码更新 → 推送到 GitHub → GitHub Actions 触发
    ↓
构建 Docker 镜像 → 推送到 Docker Hub
    ↓
同步 README → Docker Hub 更新完成
```

## 维护建议

1. **定期检查**
   - 每次发布新版本时更新 README
   - 检查链接是否有效
   - 更新版本号和示例

2. **用户反馈**
   - 关注 Docker Hub 评论
   - 根据反馈改进文档
   - 添加常见问题解答

3. **保持一致**
   - README 与实际功能一致
   - 示例代码可直接运行
   - 版本号准确无误

