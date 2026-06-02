@echo off
chcp 65001 >nul 2>&1

echo.
echo ╔════════════════════════════════════════════╗
echo   ║  Generator Platform - 一键停止脚本        ║
echo ╚════════════════════════════════════════════╝
echo.

:: ==================== [1/4] Frontend ====================
echo [1/5] 停止 Web Admin Frontend...
tasklist /FI "WINDOWTITLE eq WebAdmin-3000*" 2>nul | find /i "cmd.exe" >nul && (
    taskkill /FI "WINDOWTITLE eq WebAdmin-3000*" /F >nul 2>&1
    echo     [OK] Web Admin 已停止
) || (
    taskkill /FI "IMAGENAME eq node.exe" /F >nul 2>&1
    echo     [OK] Node 进程已停止（或未运行）
)
timeout /t 1 /nobreak >nul
echo.

:: ==================== [2/4] Generator Service ====================
echo [2/5] 停止 Generator Service (:8084)...
tasklist /FI "WINDOWTITLE eq Generator-8084*" 2>nul | find /i "cmd.exe" >nul && (
    taskkill /FI "WINDOWTITLE eq Generator-8084*" /F >nul 2>&1
    echo     [OK] Generator 已停止
) || (
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq generator-service.exe" /NH /FO CSV ^| findstr generator-service') do (
        taskkill /PID %%a /F >nul 2>&1
    )
    echo     [OK] Generator 进程已停止（或未运行）
)
timeout /t 1 /nobreak >nul
echo.

:: ==================== [3/4] Backend Service ====================
echo [3/5] 停止 Backend Service (:8080)...
tasklist /FI "WINDOWTITLE eq Backend-8080*" 2>nul | find /i "cmd.exe" >nul && (
    taskkill /FI "WINDOWTITLE eq Backend-8080*" /F >nul 2>&1
    echo     [OK] Backend 已停止
) || (
    for /f "tokens=2" %%a in ('tasklist /FI "IMAGENAME eq backend-service.exe" /NH /FO CSV ^| findstr backend-service') do (
        taskkill /PID %%a /F >nul 2>&1
    )
    echo     [OK] Backend 进程已停止（或未运行）
)
timeout /t 1 /nobreak >nul
echo.

:: ==================== [4/4] PostgreSQL (可选) ====================
echo [4/5] PostgreSQL 状态:
docker ps --filter "name=local-postgres" --format "{{.Names}}: {{.Status}}" 2>nul
if errorlevel 1 (
    echo     [提示] PostgreSQL 容器不存在
) else (
    echo     [提示] PostgreSQL 继续运行（数据保留）"
    echo           如需停止: docker stop local-postgres"
    echo           如需删除: docker rm -f local-postgres（⚠️ 数据丢失）"
)
echo.

:: ==================== 验证 ====================
echo [5/5] 验证服务状态...
set "ALL_STOPPED=1"

:: 检查端口
netstat -ano | findstr ":8080.*LISTENING" >nul 2>&1 && set "ALL_STOPPED=0"
netstat -ano | findstr ":8084.*LISTENING" >nul 2>&1 && set "ALL_STOPPED=0"

if "%ALL_STOPPED%"=="1" (
    echo     ✅ 所有后端服务已成功停止
) else (
    echo     ⚠️ 部分服务可能仍在运行，请手动检查任务管理器
)

echo.
echo ╔════════════════════════════════════════════╗
echo   ║       ✅ 停止操作完成！               ║
echo ╚════════════════════════════════════════════╝
echo.
echo   💡 提示：
echo      - PostgreSQL 数据库保持运行，数据不会丢失
echo      - 再次运行 start-all.bat 可重新启动所有服务
echo      - 如需完全清理：运行 stop.bat 后执行 docker stop local-postgres
echo.
pause
