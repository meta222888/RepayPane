@echo off
setlocal
cd /d "%~dp0"

set "VER=%~1"
if "%VER%"=="" set "VER=1.0.1"

if not exist logo.png (
  echo logo.png not found in project root
  exit /b 1
)

copy /Y logo.png internal\assets\logo.png >nul

go install github.com/tc-hib/go-winres@latest
if errorlevel 1 exit /b 1

go-winres simply --icon logo.png --manifest gui ^
  --file-version %VER% --product-version %VER% ^
  --file-description "RelayPane SFTP Client" ^
  --product-name "RelayPane" ^
  --original-filename "RelayPane.exe" ^
  --out cmd\relaypane\rsrc --arch amd64
if errorlevel 1 exit /b 1

if not exist release mkdir release

go build -ldflags="-H=windowsgui -s -w" -o release\RelayPane.exe ./cmd/relaypane
if errorlevel 1 exit /b 1

echo Built release\RelayPane.exe
