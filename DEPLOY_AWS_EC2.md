# AWS EC2 部署指南

本文档专门说明如何将 Sub2API 部署到 AWS EC2 服务器。

## 前置条件

### 1. EC2 实例要求

- ✅ 实例已启动并运行
- ✅ 安全组允许 SSH 访问（端口 22）
- ✅ 安全组允许应用访问（端口 8080 或你配置的端口）
- ✅ 已下载 `.pem` 密钥文件到本地

### 2. 本地准备

确保 `.pem` 密钥文件权限正确：

```bash
# 查找你的密钥文件（通常在 ~/.ssh/ 或 ~/Downloads/）
ls -lh ~/.ssh/*.pem
ls -lh ~/Downloads/*.pem

# 设置正确的权限（必须是 400 或 600）
chmod 400 ~/.ssh/your-key.pem
```

## 快速开始

### 步骤 1: 配置部署脚本

```bash
cd /Volumes/mac/work/ai/wei-sub2api/sub2api_new/sub2api

# 复制配置模板
cp deploy.config.example deploy.config

# 编辑配置文件
nano deploy.config
```

**AWS EC2 配置示例**:

```bash
# 服务器配置
SERVER_USER="ubuntu"                    # Ubuntu AMI 使用 ubuntu
                                        # Amazon Linux 使用 ec2-user
                                        # 如果自定义用户，使用对应用户名

SERVER_HOST="ec2-xx-xx-xx-xx.compute-1.amazonaws.com"  # 公有 DNS 或 IP
                                                        # 例如: 3.80.123.45

SERVER_PORT="22"                        # 默认 SSH 端口

# AWS EC2 密钥文件配置
SSH_KEY_FILE="/Users/你的用户名/.ssh/your-key.pem"  # 你下载的 .pem 密钥文件路径
                                                     # 例如: ~/.ssh/my-ec2-key.pem

# 远程路径配置
REMOTE_DIR="/home/sub2api"              # 部署目录
SERVICE_NAME="sub2api"                  # systemd 服务名

# 本地路径配置
LOCAL_BINARY="backend/bin/server"       # 编译后的二进制文件

# 备份配置
KEEP_BACKUPS=5                          # 保留 5 个备份
```

### 步骤 2: 测试连接

```bash
# 测试 SSH 连接
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-host

# 如果连接成功，输入 exit 退出
exit
```

### 步骤 3: 执行部署

```bash
./deploy.sh
```

## 常见 AMI 的默认用户名

| AMI 类型 | 默认用户名 |
|---------|----------|
| Ubuntu | `ubuntu` |
| Amazon Linux 2/2023 | `ec2-user` |
| CentOS | `centos` |
| Debian | `admin` |
| RHEL | `ec2-user` |

## AWS EC2 特定配置

### 1. 安全组配置

确保安全组规则允许以下端口：

```
入站规则:
- SSH (22)        来源: 你的 IP 或 0.0.0.0/0
- Custom TCP (8080) 来源: 0.0.0.0/0  # 应用端口
```

在 AWS 控制台中：
1. 进入 **EC2 Dashboard**
2. 选择 **Security Groups**
3. 找到你的实例安全组
4. 点击 **Edit inbound rules**
5. 添加规则

### 2. 获取 EC2 公有 IP/DNS

```bash
# 方法 1: 在 AWS 控制台查看
EC2 Dashboard > Instances > 选择实例 > 查看 "Public IPv4 address" 或 "Public IPv4 DNS"

# 方法 2: 在 EC2 实例内查询
curl http://169.254.169.254/latest/meta-data/public-ipv4
curl http://169.254.169.254/latest/meta-data/public-hostname
```

### 3. 密钥文件位置

下载的 `.pem` 文件通常在：
- macOS/Linux: `~/Downloads/your-key.pem`
- Windows: `C:\Users\YourName\Downloads\your-key.pem`

**推荐**：将密钥移动到 `~/.ssh/` 目录：

```bash
mv ~/Downloads/your-key.pem ~/.ssh/
chmod 400 ~/.ssh/your-key.pem
```

## 故障排查

### 问题 1: 密钥权限错误

**错误信息**:
```
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@         WARNING: UNPROTECTED PRIVATE KEY FILE!          @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
Permissions 0644 for '/path/to/key.pem' are too open.
```

**解决方法**:
```bash
chmod 400 ~/.ssh/your-key.pem
```

### 问题 2: 无法连接到服务器

**可能原因**:

1. **安全组未开放 SSH 端口**
   - 检查安全组入站规则
   - 确保允许端口 22

2. **密钥文件路径错误**
   ```bash
   # 验证密钥文件存在
   ls -lh ~/.ssh/your-key.pem
   ```

3. **用户名错误**
   - Ubuntu 使用 `ubuntu`
   - Amazon Linux 使用 `ec2-user`

4. **实例已停止**
   - 检查 EC2 实例状态

### 问题 3: sudo 权限不足

**错误**: `sudo: a terminal is required to read the password`

**解决方法**:

```bash
# 在 EC2 实例上配置 sudo 免密码
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-host

# 添加当前用户到 sudoers
echo "ubuntu ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/ubuntu
```

### 问题 4: 端口被占用

检查端口使用情况：

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@your-ec2-host "sudo lsof -i :8080"
```

## 完整部署示例

假设你的配置：
- EC2 公有 IP: `3.80.123.45`
- 密钥文件: `~/.ssh/my-ec2-key.pem`
- AMI: Ubuntu 22.04

**配置文件** (`deploy.config`):
```bash
SERVER_USER="ubuntu"
SERVER_HOST="3.80.123.45"
SERVER_PORT="22"
SSH_KEY_FILE="/Users/yourname/.ssh/my-ec2-key.pem"
REMOTE_DIR="/home/sub2api"
SERVICE_NAME="sub2api"
LOCAL_BINARY="backend/bin/server"
KEEP_BACKUPS=5
```

**执行部署**:
```bash
./deploy.sh
```

**预期输出**:
```
========================================
Sub2API 自动化部署脚本
========================================

加载配置文件: deploy.config
使用 SSH 密钥: /Users/yourname/.ssh/my-ec2-key.pem
目标服务器: ubuntu@3.80.123.45
远程目录: /home/sub2api

[1/7] 开始编译...
✓ 编译成功

[2/7] 测试服务器连接...
✓ 服务器连接正常

[3/7] 停止远程服务...
✓ 服务已停止

[4/7] 备份旧版本...
备份文件: sub2api.backup.20260124_143025
✓ 备份完成

[5/7] 上传新版本...
sub2api                    100%   45MB  15.2MB/s   00:03
✓ 上传完成

[6/7] 启动服务...
✓ 服务已启动

[7/7] 检查服务状态...
✓ 服务运行正常

========================================
部署成功! 🎉
========================================

查看日志:
  ssh -i /Users/yourname/.ssh/my-ec2-key.pem ubuntu@3.80.123.45 "tail -f /home/sub2api/logs/sub2api.log"

查看服务状态:
  ssh -i /Users/yourname/.ssh/my-ec2-key.pem ubuntu@3.80.123.45 "sudo systemctl status sub2api"
```

## 验证部署

### 1. 查看服务状态

```bash
ssh -i ~/.ssh/your-key.pem ubuntu@3.80.123.45 "sudo systemctl status sub2api"
```

### 2. 查看日志

```bash
# 实时查看日志
ssh -i ~/.ssh/your-key.pem ubuntu@3.80.123.45 "tail -f /home/sub2api/logs/sub2api.log"

# 查看日志文件列表（检查按天轮换）
ssh -i ~/.ssh/your-key.pem ubuntu@3.80.123.45 "ls -lh /home/sub2api/logs/"
```

应该看到：
```
lrwxrwxrwx 1 sub2api sub2api   24 Jan 24 14:30 sub2api.log -> sub2api-2026-01-24.log
-rw-r--r-- 1 sub2api sub2api 1.2M Jan 24 14:35 sub2api-2026-01-24.log
-rw-r--r-- 1 sub2api sub2api 5.4M Jan 23 23:59 sub2api-2026-01-23.log
-rw-r--r-- 1 sub2api sub2api 4.8M Jan 22 23:59 sub2api-2026-01-22.log
```

### 3. 测试 API

```bash
# 从本地测试
curl http://3.80.123.45:8080/health

# 或者使用公有 DNS
curl http://ec2-xx-xx-xx-xx.compute-1.amazonaws.com:8080/health
```

## 性能优化建议

### 1. 使用弹性 IP

为避免重启后 IP 变化，分配弹性 IP：
1. EC2 Dashboard > Elastic IPs
2. Allocate Elastic IP address
3. Associate with instance

### 2. 自动扩展

考虑使用 Auto Scaling Group 进行自动扩展。

### 3. 负载均衡

多实例部署时使用 Application Load Balancer。

## 安全建议

1. ✅ 限制 SSH 访问来源 IP（安全组规则）
2. ✅ 定期轮换密钥
3. ✅ 使用 IAM 角色而非密钥（如果可能）
4. ✅ 启用 CloudWatch 日志监控
5. ✅ 定期更新系统补丁

## 成本优化

1. 使用 Spot Instances 降低成本（非生产环境）
2. 监控 CPU 使用率，选择合适的实例类型
3. 使用 Reserved Instances（长期运行）
