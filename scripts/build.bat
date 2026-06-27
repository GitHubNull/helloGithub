@echo off
chcp 65001 >nul
echo ==========================================
echo   GitHub Fast DNS - Build Script
echo ==========================================

set "PROJECT_ROOT=%~dp0.."
cd /d "%PROJECT_ROOT%"

if not exist "bin" mkdir bin

echo.
echo [1/3] Downloading dependencies...
go mod tidy
if errorlevel 1 (
    echo Failed to download dependencies.
    pause
    exit /b 1
)

echo.
echo [2/3] Building for current platform (Windows)...
go build -ldflags "-s -w" -o bin\github-fast-dns.exe src\cmd\main.go
if errorlevel 1 (
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo [3/3] Build completed successfully!
echo   Output: bin\github-fast-dns.exe
echo.

pause
