@echo off
chcp 65001 >nul 2>&1
title Generator Platform - Quick Start

echo.
echo ====================================
echo   Generator Platform Launcher
echo ====================================
echo.

:: Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go not found in PATH
    echo Please install Go from: https://golang.org/dl/
    goto :error_exit
)

:: Check if Node.js is installed
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Node.js not found in PATH
    echo Please install Node.js from: https://nodejs.org/
    goto :error_exit
)

echo [1/5] Checking environment...
for /f "tokens=*" %%i in ('go version') do echo   [OK] %%i
for /f "tokens=*" %%i in ('node --version') do echo   [OK] %%i
echo.

echo [2/5] Building backend...
cd /d "%~dp0generator-service"
go build -o generator-service.exe . 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Build failed
    goto :error_exit
)
echo   [OK] Build successful
echo.

echo [3/5] Checking frontend...
cd /d "%~dp0web-admin"
if not exist "node_modules" (
    echo   Installing dependencies...
    call npm install >nul 2>&1
)
echo   [OK] Frontend ready
echo.

echo [4/5] Cleaning ports...
call :kill_port 8084
call :kill_port 3000
echo   [OK] Ports cleaned
echo.

echo [5/5] Starting services...

:: Create logs directory
if not exist "%~dp0logs" mkdir "%~dp0logs" >nul 2>&1

:: Start backend
cd /d "%~dp0generator-service"
start "" /B generator-service.exe > "%~dp0logs\backend.log" 2>&1
timeout /t 2 /nobreak >nul

:: Check if backend is running
netstat -ano | findstr ":8084.*LISTENING" >nul 2>&1
if %errorlevel%==0 (
    echo   [OK] Backend started on port 8084
) else (
    echo   [!] Backend may have failed to start
)

:: Start frontend
cd /d "%~dp0web-admin"
start "" /B cmd /c "npm run dev" > "%~dp0logs\frontend.log" 2>&1
timeout /t 3 /nobreak >nul

netstat -ano | findstr ":3000.*LISTENING" >nul 2>&1
if %errorlevel%==0 (
    echo   [OK] Frontend started on port 3000
) else (
    echo   [!] Frontend may need more time
)
echo.

echo ====================================
echo   READY! All Services Started!
echo ====================================
echo.
echo   Open: http://localhost:3000
echo   Account: admin / admin123
echo.
echo   Commands:
echo     Stop all:  double-click stop.bat
echo     View logs: type logs\backend.log
echo.
pause
goto :eof

:error_exit
echo.
echo ====================================
echo   Startup Failed!
echo ====================================
echo.
pause
exit /b 1

:kill_port
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%1.*LISTENING" 2^>nul') do (
    taskkill /PID %%a /F >nul 2>&1
)
goto :eof
