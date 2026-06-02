@echo off
chcp 65001 >nul 2>&1
title Generator Platform v2.0 - Quick Start

echo.
echo ====================================
echo   Generator Platform v2.0 Launcher
echo ====================================
echo.

:: Database configuration
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_PASSWORD=postgres
set DB_NAME=generator_platform

:: Check if PostgreSQL is available (Docker or local)
echo [0/6] Checking environment...
where psql >nul 2>&1
if %errorlevel%==0 (
    echo   [OK] PostgreSQL client found
    
    :: Test connection
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -c "SELECT 1;" >nul 2>&1
    if %errorlevel%==0 (
        echo   [OK] PostgreSQL connected on %DB_HOST%:%DB_PORT%
        
        :: Check if database exists
        psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -t -A -c "SELECT 1 FROM pg_database WHERE datname='%DB_NAME%'" >nul 2>&1
        if errorlevel 1 (
            echo   [*] Creating database: %DB_NAME%...
            psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -c "CREATE DATABASE "%DB_NAME%" WITH ENCODING='UTF8';" >nul 2>&1
            echo   [OK] Database created
        ) else (
            echo   [OK] Database '%DB_NAME%' already exists
        )
        
        :: Enable PostgreSQL for the application
        set DATABASE_URL=postgres://%DB_USER%:%DB_PASSWORD%@%DB_HOST%:%DB_PORT%/%DB_NAME%?sslmode=disable
        set DB_HOST=%DB_HOST%
        set DB_PORT=%DB_PORT%
        set DB_USER=%DB_USER%
        set DB_PASSWORD=%DB_PASSWORD%
        set DB_NAME=%DB_NAME%
        echo   [*] Will use PostgreSQL (persistent storage)
    ) else (
        echo   [!] PostgreSQL not responding, will use memory storage
    )
) else (
    echo   [!] PostgreSQL not found, will use memory storage
)

:: Check Go
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo   [FAIL] Go not found
    goto :error_exit
)
for /f "tokens=*" %%i in ('go version') do echo   [OK] %%i

:: Check Node.js
where node >nul 2>&1
if %errorlevel% neq 0 (
    echo   [FAIL] Node.js not found
    goto :error_exit
)
for /f "tokens=*" %%i in ('node --version') do echo   [OK] %%i
echo.

echo [1/6] Building backend...
cd /d "%~dp0generator-service"
go build -o generator-service.exe . 2>&1
if %errorlevel% neq 0 (
    echo   [FAIL] Build failed
    goto :error_exit
)
echo   [OK] Build successful
echo.

echo [2/6] Checking frontend...
cd /d "%~dp0web-admin"
if not exist "node_modules" (
    echo   Installing dependencies...
    call npm install >nul 2>&1
)
echo   [OK] Frontend ready
echo.

echo [3/6] Cleaning ports...
call :kill_port 8084
call :kill_port 3000
echo   [OK] Ports cleaned
echo.

echo [4/6] Starting services...
if not exist "%~dp0logs" mkdir "%~dp0logs" >nul 2>&1

cd /d "%~dp0generator-service"
start "" /B generator-service.exe > "%~dp0logs\backend.log" 2>&1
timeout /t 3 /nobreak >nul

netstat -ano | findstr ":8084.*LISTENING" >nul 2>&1
if %errorlevel%==0 (
    echo   [OK] Backend started on port 8084
) else (
    echo   [!] Backend may have failed to start
)

cd /d "%~dp0web-admin"
start "" /B cmd /c "npm run dev" > "%~dp0logs\frontend.log" 2>&1
timeout /t 4 /nobreak >nul

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
echo   Database:
if defined DATABASE_URL (
    echo     Type: PostgreSQL (persistent)
    echo     Host: %DB_HOST%:%DB_PORT%
    echo     Name: %DB_NAME%
) else (
    echo     Type: Memory (ephemeral)
)
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
