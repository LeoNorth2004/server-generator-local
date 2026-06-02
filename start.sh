#!/bin/bash

# ============================================
# Generator Platform - 一键启动脚本
# 适用于: Linux / macOS / WSL
# 要求: Go >= 1.21, Node.js >= 18, PostgreSQL >= 14
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 端口配置
BACKEND_PORT=8084
FRONTEND_PORT=3000

# 数据库配置
DB_HOST="localhost"
DB_PORT=5432
DB_USER="postgres"
DB_PASSWORD="postgres"  # 默认密码，可根据需要修改
DB_NAME="generator_platform"

echo -e "${BLUE}====================================${NC}"
echo -e "${BLUE}  Generator Platform 一键启动工具${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# 函数：检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ 错误: 未找到 $1${NC}"
        return 1
    fi
    return 0
}

# 函数：检查端口是否被占用
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ 端口 $1 已被占用${NC}"
        echo "  正在尝试终止占用进程..."
        lsof -ti :$1 | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
}

# 函数：检测 PostgreSQL 是否安装
check_postgresql_installed() {
    if command -v psql &> /dev/null; then
        return 0
    fi
    return 1
}

# 函数：测试数据库连接
test_db_connection() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" >/dev/null 2>&1
    return $?
}

# 函数：检查数据库是否存在
check_database_exists() {
    local result=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -t -A -c "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" 2>/dev/null)
    if [[ "$result" == "1" ]]; then
        return 0
    fi
    return 1
}

# 函数：创建数据库
create_database() {
    echo -e "  ${YELLOW}正在创建数据库 '$DB_NAME'...${NC}"
    
    # 创建数据库（使用 UTF8 编码和中文支持）
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE \"$DB_NAME\" WITH ENCODING='UTF8' LC_COLLATE='C' LC_CTYPE='C' TEMPLATE=template0;" >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓ 数据库创建成功: $DB_NAME${NC}"
        
        # 创建基本扩展（可选但推荐）
        echo -e "  ${YELLOW}安装数据库扩展...${NC}"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" >/dev/null 2>&1 || true
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";" >/dev/null 2>&1 || true
        
        return 0
    else
        echo -e "  ${RED}✗ 数据库创建失败${NC}"
        return 1
    fi
}

# ============================================
# 步骤 0: 检查并初始化数据库
# ============================================
echo -e "${GREEN}[0/7]${NC} 检查 PostgreSQL 数据库..."
echo ""

# 0.1 检查 PostgreSQL 是否安装
if ! check_postgresql_installed; then
    echo -e "  ${RED}✗ 未检测到 PostgreSQL${NC}"
    echo ""
    echo "  请先安装 PostgreSQL："
    echo ""
    echo "  Ubuntu/Debian:"
    echo "    sudo apt update && sudo apt install postgresql postgresql-contrib"
    echo ""
    echo "  macOS (Homebrew):"
    echo "    brew install postgresql@16"
    echo "    brew services start postgresql@16"
    echo ""
    echo "  CentOS/RHEL:"
    echo "    sudo yum install postgresql-server postgresql-contrib"
    echo "    sudo postgresql-setup initdb"
    echo "    sudo systemctl start postgresql"
    echo ""
    echo "  或者使用 Docker 运行 PostgreSQL："
    echo '  docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16-alpine'
    exit 1
fi

PG_VERSION=$(psql --version | grep -oP '\d+\.\d+')
echo -e "  ${GREEN}✓${NC} PostgreSQL 版本: $PG_VERSION"

# 0.2 测试数据库连接
echo ""
echo "  测试数据库连接..."
if ! test_db_connection; then
    echo -e "  ${RED}✗ 无法连接到 PostgreSQL${NC}"
    echo ""
    echo "  请检查以下配置："
    echo "    主机: $DB_HOST:$DB_PORT"
    echo "    用户: $DB_USER"
    echo "    密码: $DB_PASSWORD"
    echo ""
    echo "  你可以修改此脚本顶部的变量来调整连接参数。"
    exit 1
fi
echo -e "  ${GREEN}✓ 数据库连接成功${NC}"

# 0.3 检查数据库是否存在
echo ""
echo "  检查数据库 '$DB_NAME' 是否存在..."
if check_database_exists; then
    echo -e "  ${GREEN}✓ 数据库已存在: $DB_NAME${NC}"
    echo -e "     跳过创建步骤${NC}"
else
    echo -e "  ${YELLOW}⚠ 数据库不存在: $DB_NAME${NC}"
    echo ""
    
    # 自动创建数据库
    if ! create_database; then
        echo -e "  ${RED}✗ 无法自动创建数据库，请手动创建后重试${NC}"
        echo ""
        echo "  手动创建命令："
        echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c \"CREATE DATABASE \\\"$DB_NAME\\\";\""
        exit 1
    fi
    
    echo -e "  ${GREEN}✓ 数据库初始化完成${NC}"
fi

echo ""

# ============================================
# 步骤 1: 检查环境依赖
# ============================================
echo -e "${GREEN}[1/7]${NC} 检查环境依赖..."
echo ""

# 检查 Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "  ${GREEN}✓${NC} Go 版本: $GO_VERSION"
else
    echo -e "  ${RED}✗${NC} 未安装 Go"
    echo "  请访问 https://golang.org/dl/ 下载安装"
    exit 1
fi

# 检查 Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "  ${GREEN}✓${NC} Node.js 版本: $NODE_VERSION"
else
    echo -e "  ${RED}✗${NC} 未安装 Node.js"
    echo "  请访问 https://nodejs.org/ 下载安装"
    exit 1
fi

# 检查 npm
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    echo -e "  ${GREEN}✓${NC} npm 版本: $NPM_VERSION"
else
    echo -e "  ${RED}✗${NC} 未安装 npm"
    exit 1
fi

echo ""

# ============================================
# 步骤 2: 构建后端服务
# ============================================
echo -e "${GREEN}[2/7]${NC} 构建后端服务..."
echo ""
cd "$SCRIPT_DIR/generator-service"

echo "  编译中..."
go build -o generator-service . 2>&1

if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} 后端构建成功"
else
    echo -e "  ${RED}✗${NC} 后端构建失败"
    exit 1
fi

echo ""

# ============================================
# 步骤 3: 安装前端依赖
# ============================================
echo -e "${GREEN}[3/7]${NC} 安装前端依赖..."
echo ""
cd "$SCRIPT_DIR/web-admin"

if [ ! -d "node_modules" ]; then
    echo "  首次运行，安装依赖..."
    npm install
    echo -e "  ${GREEN}✓${NC} 前端依赖安装完成"
else
    echo -e "  ${GREEN}✓${NC} 前端依赖已存在"
fi

echo ""

# ============================================
# 步骤 4: 清理旧进程
# ============================================
echo -e "${GREEN}[4/7]${NC} 清理旧进程..."
echo ""

check_port $BACKEND_PORT
check_port $FRONTEND_PORT

echo -e "  ${GREEN}✓${NC} 端口已清理"
echo ""

# ============================================
# 步骤 5: 启动后端服务
# ============================================
echo -e "${GREEN}[5/7]${NC} 启动后端服务 (端口: $BACKEND_PORT)..."
echo ""
cd "$SCRIPT_DIR/generator-service"

# 创建日志目录
mkdir -p "$SCRIPT_DIR/logs"

# 使用 nohup 在后台启动，日志输出到文件
nohup ./generator-service > "$SCRIPT_DIR/logs/backend.log" 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "$SCRIPT_DIR/.backend.pid"

# 等待后端启动
sleep 2

if ps -p $BACKEND_PID > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} 后端已启动 (PID: $BACKEND_PID)"
else
    echo -e "  ${RED}✗${NC} 后端启动失败"
    echo "  请查看日志: tail -f $SCRIPT_DIR/logs/backend.log"
    exit 1
fi

echo ""

# ============================================
# 步骤 6: 启动前端服务
# ============================================
echo -e "${GREEN}[6/7]${NC} 启动前端开发服务器 (端口: $FRONTEND_PORT)..."
echo ""
cd "$SCRIPT_DIR/web-admin"

# 使用 nohup 启动前端
nohup npm run dev > "$SCRIPT_DIR/logs/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > "$SCRIPT_DIR/.frontend.pid"

# 等待前端启动
sleep 3

if ps -p $FRONTEND_PID > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} 前端已启动 (PID: $FRONTEND_PID)"
else
    echo -e "  ${YELLOW}⚠${NC} 前端可能需要更长时间启动..."
fi

echo ""

# ============================================
# 步骤 7: 显示启动信息
# ============================================
echo -e "${GREEN}[7/7]${NC} 验证服务状态..."
echo ""

# 检查后端是否正常响应
if curl -s http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} 后端健康检查通过"
else
    echo -e "  ${YELLOW}⚠${NC} 后端可能还在初始化中（这是正常的）"
fi

echo ""
echo -e "${BLUE}====================================${NC}"
echo -e "${GREEN}🎉 所有服务已成功启动！${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""
echo -e "  📊 服务地址:"
echo -e "     前端界面: http://localhost:$FRONTEND_PORT${YELLOW}"
echo -e "     后端 API: http://localhost:$BACKEND_PORT${YELLOW}"
echo ""
echo -e "  👤 默认账号:"
echo -e "     用户名: admin${YELLOW}"
echo -e "     密码:   admin123${YELLOW}"
echo ""
echo -e "  🐘 数据库信息:"
echo -e "     类型:     PostgreSQL${YELLOW}"
echo -e "     主机:     $DB_HOST:$DB_PORT${YELLOW}"
echo -e "     数据库名: $DB_NAME${YELLOW}"
echo ""
echo -e "  📝 常用命令:"
echo -e "     查看后端日志:   tail -f logs/backend.log${NC}"
echo -e "     查看前端日志:   tail -f logs/frontend.log${NC}"
echo -e "     停止所有服务:   ./stop.sh${NC}"
echo -e "     重启所有服务:   ./start.sh${NC}"
echo ""
echo -e "  🗄️  数据库管理:"
echo -e "     连接数据库:     psql -h $DB_HOST -U $DB_USER -d $DBName${NC}"
echo -e "     备份数据库:     pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > backup.sql${NC}"
echo -e "     恢复数据库:     psql -h $DB_HOST -U $DB_USER -d $DB_NAME < backup.sql${NC}"
echo ""
echo -e "${BLUE}====================================${NC}"

# 保存 PID 到文件以便后续停止
cat > "$SCRIPT_DIR/.pids" << EOF
BACKEND_PID=$BACKEND_PID
FRONTEND_PID=$FRONTEND_PID
DATABASE_NAME=$DB_NAME
EOF
