@echo off
chcp 65001 >nul
echo ==========================================
echo   GitHub Fast DNS - One-Click Run
echo ==========================================

set "PROJECT_ROOT=%~dp0.."
cd /d "%PROJECT_ROOT%"

:: Check if running as administrator
net session >nul 2>&1
if errorlevel 1 (
    echo This script requires administrator privileges.
    echo Requesting elevation...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

if not exist "bin\github-fast-dns.exe" (
    echo Binary not found, building first...
    call scripts\build.bat
    if errorlevel 1 exit /b 1
)

echo.
echo Starting GitHub Fast DNS service...
echo Press Ctrl+C to stop
echo.

bin\github-fast-dns.exe

pause
