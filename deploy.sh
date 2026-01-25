#!/bin/bash

# Sub2API 自动化部署脚本
# 使用方法:
#   ./deploy.sh [server_user@server_ip]
#   或配置 deploy.config 文件后直接运行: ./deploy.sh

set -e  # 遇到错误立即退出

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置参数
SERVER_USER="ubuntu"
SERVER_HOST="your-server-ip"
SERVER_PORT="22"
SSH_KEY_FILE=""
REMOTE_DIR="/home/sub2api"
SERVICE_NAME="sub2api"
LOCAL_BINARY="backend/bin/server"
KEEP_BACKUPS=5

# 尝试加载配置文件
CONFIG_FILE="deploy.config"
if [ -f "$CONFIG_FILE" ]; then
    echo -e "${BLUE}加载配置文件: ${CONFIG_FILE}${NC}"
    # shellcheck source=deploy.config
    source "$CONFIG_FILE"
fi

# 如果提供了命令行参数，优先使用命令行参数
if [ -n "$1" ]; then
    SERVER_USER_HOST="$1"
else
    SERVER_USER_HOST="${SERVER_USER}@${SERVER_HOST}"
fi

# 构建 SSH 命令参数
SSH_OPTS=""
if [ -n "$SSH_KEY_FILE" ]; then
    if [ ! -f "$SSH_KEY_FILE" ]; then
        echo -e "${RED}错误: SSH 密钥文件不存在: ${SSH_KEY_FILE}${NC}"
        exit 1
    fi
    # 检查密钥文件权限
    KEY_PERMS=$(stat -f "%OLp" "$SSH_KEY_FILE" 2>/dev/null || stat -c "%a" "$SSH_KEY_FILE" 2>/dev/null)
    if [ "$KEY_PERMS" != "400" ] && [ "$KEY_PERMS" != "600" ]; then
        echo -e "${YELLOW}警告: SSH 密钥文件权限不正确 (当前: ${KEY_PERMS})${NC}"
        echo -e "${YELLOW}修复权限: chmod 400 ${SSH_KEY_FILE}${NC}"
        chmod 400 "$SSH_KEY_FILE"
    fi
    SSH_OPTS="-i ${SSH_KEY_FILE}"
    echo -e "${BLUE}使用 SSH 密钥: ${SSH_KEY_FILE}${NC}"
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Sub2API 自动化部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查参数
if [ "$SERVER_USER_HOST" = "root@your-server-ip" ]; then
    echo -e "${RED}错误: 请提供服务器地址${NC}"
    echo "使用方法: ./deploy.sh user@server-ip"
    echo "示例: ./deploy.sh root@192.168.1.100"
    exit 1
fi

echo -e "${YELLOW}目标服务器: ${SERVER_USER_HOST}${NC}"
echo -e "${YELLOW}远程目录: ${REMOTE_DIR}${NC}"
echo ""

# 步骤 1: 本地编译
echo -e "${GREEN}[1/7] 开始编译...${NC}"
cd backend
if [ -f "Makefile" ]; then
    make build
else
    go build -o bin/server ./cmd/server
fi
cd ..

if [ ! -f "$LOCAL_BINARY" ]; then
    echo -e "${RED}编译失败: 找不到 ${LOCAL_BINARY}${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 编译成功${NC}"
echo ""

# 步骤 2: 测试 SSH 连接
echo -e "${GREEN}[2/7] 测试服务器连接...${NC}"
# shellcheck disable=SC2086
if ! ssh $SSH_OPTS -o ConnectTimeout=5 "$SERVER_USER_HOST" "echo '连接成功'" > /dev/null 2>&1; then
    echo -e "${RED}无法连接到服务器: ${SERVER_USER_HOST}${NC}"
    echo "请检查:"
    echo "  1. 服务器地址是否正确"
    echo "  2. SSH 密钥文件路径是否正确 (${SSH_KEY_FILE:-使用默认配置})"
    echo "  3. 密钥文件权限是否正确 (需要 400 或 600)"
    echo "  4. 服务器安全组是否允许 SSH (端口 22)"
    echo "  5. 服务器是否在线"
    exit 1
fi
echo -e "${GREEN}✓ 服务器连接正常${NC}"
echo ""

# 步骤 3: 停止服务
echo -e "${GREEN}[3/7] 停止远程服务...${NC}"
# shellcheck disable=SC2086
ssh $SSH_OPTS "$SERVER_USER_HOST" "sudo systemctl stop ${SERVICE_NAME} || true"
echo -e "${GREEN}✓ 服务已停止${NC}"
echo ""

# 步骤 4: 备份旧版本
echo -e "${GREEN}[4/7] 备份旧版本...${NC}"
BACKUP_NAME="${SERVICE_NAME}.backup.$(date +%Y%m%d_%H%M%S)"
# shellcheck disable=SC2086
ssh $SSH_OPTS "$SERVER_USER_HOST" "
    if [ -f ${REMOTE_DIR}/${SERVICE_NAME} ]; then
        cp ${REMOTE_DIR}/${SERVICE_NAME} ${REMOTE_DIR}/${BACKUP_NAME}
        echo '备份文件: ${BACKUP_NAME}'
    else
        echo '没有旧版本需要备份'
    fi
"
echo -e "${GREEN}✓ 备份完成${NC}"
echo ""

# 步骤 5: 上传新版本
echo -e "${GREEN}[5/7] 上传新版本...${NC}"
# shellcheck disable=SC2086
scp $SSH_OPTS "$LOCAL_BINARY" "${SERVER_USER_HOST}:${REMOTE_DIR}/${SERVICE_NAME}"
echo -e "${GREEN}✓ 上传完成${NC}"
echo ""

# 步骤 6: 设置权限并启动服务
echo -e "${GREEN}[6/7] 启动服务...${NC}"
# shellcheck disable=SC2086
ssh $SSH_OPTS "$SERVER_USER_HOST" "
    chmod +x ${REMOTE_DIR}/${SERVICE_NAME}
    sudo systemctl start ${SERVICE_NAME}
    sleep 2
"
echo -e "${GREEN}✓ 服务已启动${NC}"
echo ""

# 步骤 7: 检查服务状态
echo -e "${GREEN}[7/7] 检查服务状态...${NC}"
# shellcheck disable=SC2086
SERVICE_STATUS=$(ssh $SSH_OPTS "$SERVER_USER_HOST" "sudo systemctl is-active ${SERVICE_NAME}" || echo "failed")

if [ "$SERVICE_STATUS" = "active" ]; then
    echo -e "${GREEN}✓ 服务运行正常${NC}"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}部署成功! 🎉${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "查看日志:"
    echo "  ssh ${SERVER_USER_HOST} \"tail -f ${REMOTE_DIR}/logs/sub2api.log\""
    echo ""
    echo "查看服务状态:"
    echo "  ssh ${SERVER_USER_HOST} \"sudo systemctl status ${SERVICE_NAME}\""
    echo ""
else
    echo -e "${RED}✗ 服务启动失败${NC}"
    echo ""
    echo "查看错误日志:"
    echo "  ssh ${SERVER_USER_HOST} \"sudo journalctl -u ${SERVICE_NAME} -n 50\""
    echo ""
    echo "回滚到备份版本:"
    echo "  ssh ${SERVER_USER_HOST} \"sudo systemctl stop ${SERVICE_NAME} && cp ${REMOTE_DIR}/${BACKUP_NAME} ${REMOTE_DIR}/${SERVICE_NAME} && sudo systemctl start ${SERVICE_NAME}\""
    exit 1
fi
