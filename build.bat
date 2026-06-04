@echo off
setlocal
cd /d "%~dp0"
go build -ldflags="-H=windowsgui -s -w" -o RelayPane.exe ./cmd/relaypane
if errorlevel 1 exit /b 1
echo Built RelayPane.exe (no console window)
