# Sub2API 部署指南

本文档说明如何使用自动化脚本部署 Sub2API 到远程服务器。

## 前置条件

1. **SSH 密钥配置**（推荐）
   ```bash
   # 生成 SSH 密钥（如果还没有）
   ssh-keygen -t rsa -b 4096

   # 复制公钥到服务器
   ssh-copy-id user@your-server-ip
   ```

2. **服务器环境**
   - systemd 已安装
   - sub2api.service 已配置
   - 部署目录已创建（如 /home/sub2api）

## 快速开始

### 方法 1: 直接使用命令行参数

```bash
./deploy.sh root@192.168.1.100
```

### 方法 2: 使用配置文件（推荐）

1. 复制配置文件模板：
   ```bash
   cp deploy.config.example deploy.config
   ```

2. 编辑 `deploy.config`，修改为你的实际配置：
   ```bash
   SERVER_USER="root"
   SERVER_HOST="192.168.1.100"
   SERVER_PORT="22"
   REMOTE_DIR="/home/sub2api"
   SERVICE_NAME="sub2api"
   ```

3. 运行部署脚本：
   ```bash
   ./deploy.sh
   ```

## 部署流程

脚本会自动执行以下步骤：

1. ✅ **编译** - 本地编译 Go 二进制文件
2. ✅ **连接测试** - 测试 SSH 连接是否正常
3. ✅ **停止服务** - 停止远程服务器上的服务
4. ✅ **备份** - 备份当前版本（带时间戳）
5. ✅ **上传** - 上传新编译的二进制文件
6. ✅ **启动** - 启动服务并设置权限
7. ✅ **检查** - 验证服务是否正常运行

## 常见问题

### 1. SSH 连接失败

**错误**: `无法连接到服务器`

**解决方法**:
```bash
# 测试 SSH 连接
ssh user@server-ip

# 如果使用非标准端口
ssh -p 2222 user@server-ip
```

### 2. 权限不足

**错误**: `Permission denied`

**解决方法**:
- 确保 SSH 用户有 sudo 权限
- 或者直接使用 root 用户部署

### 3. 服务启动失败

**查看日志**:
```bash
# 查看 systemd 日志
ssh user@server-ip "sudo journalctl -u sub2api -n 50"

# 查看应用日志
ssh user@server-ip "tail -n 100 /home/sub2api/logs/sub2api.log"
```

**回滚到备份版本**:
```bash
# 查看备份文件
ssh user@server-ip "ls -lh /home/sub2api/sub2api.backup.*"

# 手动回滚
ssh user@server-ip "sudo systemctl stop sub2api && \
  cp /home/sub2api/sub2api.backup.20260124_143000 /home/sub2api/sub2api && \
  sudo systemctl start sub2api"
```

## 部署后验证

### 查看服务状态
```bash
ssh user@server-ip "sudo systemctl status sub2api"
```

### 查看实时日志
```bash
ssh user@server-ip "tail -f /home/sub2api/logs/sub2api.log"
```

### 查看日志文件列表
```bash
ssh user@server-ip "ls -lh /home/sub2api/logs/"
```

应该看到按天分割的日志文件：
```
sub2api.log -> sub2api-2026-01-24.log  # 软链接指向当天日志
sub2api-2026-01-24.log                  # 今天的日志
sub2api-2026-01-23.log                  # 昨天的日志
sub2api-2026-01-22.log                  # 前天的日志
```

## 高级配置

### 自定义编译参数

修改 `backend/Makefile`:
```makefile
build:
	go build -ldflags "-s -w" -o bin/server ./cmd/server
```

### 多环境部署

创建不同的配置文件：
```bash
# 生产环境
cp deploy.config.example deploy.config.prod

# 测试环境
cp deploy.config.example deploy.config.test
```

使用时：
```bash
# 部署到生产环境
cp deploy.config.prod deploy.config
./deploy.sh

# 部署到测试环境
cp deploy.config.test deploy.config
./deploy.sh
```

## 安全建议

1. ✅ 使用 SSH 密钥认证，不要使用密码
2. ✅ 不要将 `deploy.config` 提交到 Git（已在 .gitignore）
3. ✅ 限制 SSH 用户的 sudo 权限
4. ✅ 定期清理旧的备份文件

## 故障排查

如果部署失败，按以下步骤排查：

1. **检查本地编译**
   ```bash
   cd backend
   go build -o bin/server ./cmd/server
   ```

2. **检查 SSH 连接**
   ```bash
   ssh user@server-ip "echo 'SSH OK'"
   ```

3. **检查远程目录权限**
   ```bash
   ssh user@server-ip "ls -ld /home/sub2api"
   ```

4. **手动执行部署步骤**
   ```bash
   # 按照脚本步骤一步步执行
   ssh user@server-ip "sudo systemctl stop sub2api"
   scp backend/bin/server user@server-ip:/home/sub2api/sub2api
   ssh user@server-ip "sudo systemctl start sub2api"
   ```
