@echo off
chcp 65001 >nul
echo ========================================
echo   Generator Platform - Backend Service
echo ========================================
echo.

cd /d "%~dp0backend-service"

echo [1/2] Checking PostgreSQL container...
docker ps --filter "name=local-postgres" --format "{{.Names}}" | findstr "local-postgres" >nul
if errorlevel 1 (
    echo     Starting PostgreSQL...
    docker start local-postgres
    timeout /t 3 /nobreak >nul
) else (
    echo     ✓ PostgreSQL is running
)

echo.
echo [2/2] Starting Backend Service...
echo     Database: PostgreSQL (localhost:5432)
echo     Port: 8080
echo.
set DB_PASSWORD=123456
backend-service.exe

pause
