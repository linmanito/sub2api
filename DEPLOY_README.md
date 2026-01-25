# 部署文档索引

本目录包含 Sub2API 的完整部署方案和工具。

## 📚 文档列表

| 文档 | 适用场景 | 说明 |
|-----|---------|-----|
| [DEPLOY.md](./DEPLOY.md) | 通用部署 | 通用部署指南，适用于所有服务器 |
| [DEPLOY_AWS_EC2.md](./DEPLOY_AWS_EC2.md) | AWS EC2 | AWS EC2 专用部署指南（**推荐 EC2 用户优先阅读**） |

## 🚀 快速开始

### AWS EC2 用户（推荐流程）

```bash
# 1. 复制 AWS EC2 配置模板
cp deploy.config.aws.example deploy.config

# 2. 编辑配置文件
nano deploy.config

# 修改以下内容:
# - SERVER_HOST: 你的 EC2 公有 IP 或 DNS
# - SSH_KEY_FILE: 你的 .pem 密钥文件路径

# 3. 设置密钥权限
chmod 400 ~/.ssh/your-key.pem

# 4. 执行部署
./deploy.sh
```

### 其他服务器用户

```bash
# 1. 复制通用配置模板
cp deploy.config.example deploy.config

# 2. 编辑配置文件
nano deploy.config

# 3. 执行部署
./deploy.sh
```

## 📁 文件说明

```
部署相关文件:
├── deploy.sh                      # 自动化部署脚本（主文件）
├── deploy.config.example          # 通用配置模板
├── deploy.config.aws.example      # AWS EC2 配置模板
├── DEPLOY.md                      # 通用部署指南
├── DEPLOY_AWS_EC2.md             # AWS EC2 部署指南
└── DEPLOY_README.md              # 本文件（索引）
```

## 🎯 日志轮换功能

部署后，日志会自动按天轮换：

```bash
/home/sub2api/logs/
├── sub2api.log          → sub2api-2026-01-24.log  # 软链接
├── sub2api-2026-01-24.log                          # 今天
├── sub2api-2026-01-23.log                          # 昨天
└── sub2api-2026-01-22.log                          # 前天
```

**特性**:
- ✅ 每天 00:00 自动创建新日志文件
- ✅ 文件名格式: `sub2api-YYYY-MM-DD.log`
- ✅ 保留 90 天历史
- ✅ 时间格式: `2026-01-24 14:30:25`

## 🔧 常见任务

### 查看日志

```bash
# 实时查看日志
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-ip "tail -f /home/sub2api/logs/sub2api.log"

# 查看昨天的日志
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-ip "tail /home/sub2api/logs/sub2api-2026-01-23.log"
```

### 查看服务状态

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-ip "sudo systemctl status sub2api"
```

### 重启服务

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-ip "sudo systemctl restart sub2api"
```

## ⚠️ 注意事项

1. **不要提交配置文件到 Git**
   - `deploy.config` 包含敏感信息
   - 已在 `.gitignore` 中排除

2. **密钥文件权限**
   - `.pem` 文件权限必须是 `400` 或 `600`
   - 使用 `chmod 400 ~/.ssh/your-key.pem` 设置

3. **首次部署**
   - 确保服务器已安装并配置好 systemd 服务
   - 确保部署目录已创建

## 💡 故障排查

遇到问题？参考文档：
- AWS EC2 问题 → [DEPLOY_AWS_EC2.md](./DEPLOY_AWS_EC2.md#故障排查)
- 通用问题 → [DEPLOY.md](./DEPLOY.md#故障排查)

## 🆘 获取帮助

如果部署遇到问题：

1. 查看对应的详细文档
2. 检查脚本输出的错误信息
3. 查看服务器日志
4. 提交 Issue 到项目仓库
