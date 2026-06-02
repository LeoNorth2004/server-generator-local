@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo.
echo ╔════════════════════════════════════════════╗
echo   ║  Generator Platform - 一键启动脚本        ║
echo ╚════════════════════════════════════════════╝
echo.

:: ==================== [1/5] PostgreSQL ====================
echo [1/5] 检查 PostgreSQL...
docker ps --filter "name=local-postgres" --format "{{.Names}}" | findstr "local-postgres" >nul 2>&1
if errorlevel 1 (
    echo     [创建] PostgreSQL 容器（首次运行）...
    docker run -d --name local-postgres -e POSTGRES_PASSWORD=123456 -e POSTGRES_DB=generator_platform -p 5432:5432 postgres:15-alpine >nul 2>&1
    if errorlevel 1 (
        echo     [错误] Docker 未运行或安装失败！请先启动 Docker Desktop
        pause
        exit /b 1
    )
    echo     [等待] 数据库初始化（约5秒）...
    timeout /t 5 /nobreak >nul
) else (
    echo     [OK] PostgreSQL 已运行
)
echo.

:: ==================== [2/5] Backend Service ====================
echo [2/5] 启动 Backend Service (:8080)...
cd /d "%~dp0backend-service"
if not exist backend-service.exe (
    echo     [编译] Backend Service（首次需要1-2分钟）...
    call go build -o backend-service.exe .
    if errorlevel 1 (
        echo     [错误] 编译失败！
        pause
        exit /b 1
    )
)
:: 后台启动 Backend（使用 cmd /c 让它在后台运行）
start "Backend-8080" /min cmd /c "set DB_PASSWORD=123456 && backend-service.exe"
echo     [等待] Backend 启动中...
timeout /t 4 /nobreak >nul
:: 健康检查
curl -s http://localhost:8080/api/v1/operations/health >nul 2>&1
if errorlevel 1 (
    echo     [警告] Backend 可能还在启动中，稍后手动检查 :8080
) else (
    echo     [OK] Backend 已就绪
)
echo.

:: ==================== [3/5] Generator Service ====================
echo [3/5] 启动 Generator Service (:8084)...
cd /d "%~dp0generator-service"
if not exist generator-service.exe (
    echo     [编译] Generator Service（首次需要1-2分钟）...
    call go build -o generator-service.exe .
    if errorlevel 1 (
        echo     [错误] 编译失败！
        pause
        exit /b 1
    )
)
start "Generator-8084" /min cmd /c "set DB_PASSWORD=123456 && generator-service.exe"
echo     [等待] Generator 启动中...
timeout /t 4 /nobreak >nul
curl -s http://localhost:8084/health >nul 2>&1
if errorlevel 1 (
    echo     [警告] Generator 可能还在启动中，稍后手动检查 :8084
) else (
    echo     [OK] Generator 已就绪
)
echo.

:: ==================== [4/5] Frontend ====================
echo [4/5] 启动 Web Admin Frontend (:3000)...
cd /d "%~dp0web-admin"
if not exist node_modules (
    echo     [安装] npm 依赖（首次需要1-3分钟）...
    call npm install
    if errorlevel 1 (
        echo     [错误] npm install 失败！请检查 Node.js 是否已安装
        pause
        exit /b 1
    )
)
start "WebAdmin-3000" /min cmd /c "npm run dev"
echo     [等待] 前端开发服务器启动中...
timeout /t 6 /nobreak >nul
echo     [OK] 前端已启动（端口可能是 3000 或 3001）
echo.

:: ==================== [5/5] 完成 ====================
echo.
echo ╔════════════════════════════════════════════╗
echo   ║       ✅ 所有服务已启动完成！           ║
echo ╚════════════════════════════════════════════╝
echo.
echo   服务地址：
echo   ┌───────────────────────────────────────┐
echo   │ 🌐 Web Admin:   http://localhost:3000 │
echo   │ 🔧 Backend:     http://localhost:8080 │
echo   │ ⚡ Generator:   http://localhost:8084 │
echo   │ 🐘 PostgreSQL:  localhost:5432         │
echo   └───────────────────────────────────────┘
echo.
echo   👤 默认账户：admin / admin123
echo.
echo   💡 提示：
echo      - 如端口 3000 被占用，前端会自动使用 3001
echo      - 运行 stop.bat 可一键停止所有服务
echo      - 各服务窗口可单独关闭进行调试
echo.

:: 尝试打开浏览器
set "FRONTEND_URL=http://localhost:3000"
timeout /t 1 /nobreak >nul
start "" "%FRONTEND_URL%" 2>nul

echo.
echo   按任意键退出此窗口（服务继续在后台运行）...
pause >nul
endlocal
