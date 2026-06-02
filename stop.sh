#!/bin/bash

# ============================================
# Generator Platform - 停止服务脚本
# ============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "\033[0;33m====================================\033[0m"
echo -e "  Generator Platform - 停止服务"
echo -e "\033[0;33m====================================\033[0m"
echo ""

# 停止后端
if [ -f ".backend.pid" ]; then
    BACKEND_PID=$(cat .backend.pid)
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        kill $BACKEND_PID 2>/dev/null || true
        echo -e "  \033[0;32m✓\033[0m 后端服务已停止 (PID: $BACKEND_PID)"
    else
        echo -e "  \033[0;33m⚠\033[0m 后端进程不存在"
    fi
    rm -f .backend.pid
else
    # 尝试通过端口查找并终止
    if lsof -ti :8084 > /dev/null 2>&1; then
        lsof -ti :8084 | xargs kill -9 2>/dev/null || true
        echo -e "  \033[0;32m✓\033[0m 后端服务已停止 (通过端口)"
    else
        echo -e "  \033[0;33m⚠\033[0m 后端服务未运行"
    fi
fi

# 停止前端
if [ -f ".frontend.pid" ]; then
    FRONTEND_PID=$(cat .frontend.pid)
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
        kill $FRONTEND_PID 2>/dev/null || true
        echo -e "  \033[0;32m✓\033[0m 前端服务已停止 (PID: $FRONTEND_PID)"
    else
        # 可能是子进程（vite），尝试杀死所有 node 进程
        pkill -f "vite" 2>/dev/null || true
        echo -e "  \033[0;32m✓\033[0m 前端服务已停止 (包含子进程)"
    fi
    rm -f .frontend.pid
else
    if lsof -ti :3000 > /dev/null 2>&1; then
        lsof -ti :3000 | xargs kill -9 2>/dev/null || true
        echo -e "  \033[0;32m✓\033[0m 前端服务已停止 (通过端口)"
    else
        echo -e "  \033[0;33m⚠\033[0m 前端服务未运行"
    fi
fi

# 清理 PID 文件
rm -f .pids

echo ""
echo -e "\033[0;33m====================================\033[0m"
echo -e "  \033[0;32m所有服务已成功停止！\033[0m"
echo -e "\033[0;33m====================================\033[0m"
