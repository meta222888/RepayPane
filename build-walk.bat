@echo off
setlocal EnableDelayedExpansion
cd /d "%~dp0"

set CGO_ENABLED=1

where gcc >nul 2>&1
if not errorlevel 1 goto :gcc_ok

set "MINGW="
for /d %%D in ("%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_*") do (
  if exist "%%D\mingw64\bin\gcc.exe" set "MINGW=%%D\mingw64\bin"
)
if not defined MINGW if exist "C:\mingw64\bin\gcc.exe" set "MINGW=C:\mingw64\bin"
if not defined MINGW if exist "C:\TDM-GCC-64\bin\gcc.exe" set "MINGW=C:\TDM-GCC-64\bin"

if defined MINGW (
  set "PATH=%MINGW%;%PATH%"
) else (
  echo [build-walk.bat] GCC not found. walk requires CGO/MinGW on Windows.
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  pause
  exit /b 1
)

:gcc_ok
if not exist logo.png (
  echo [build-walk.bat] logo.png not found in project root
  pause
  exit /b 1
)

copy /Y logo.png internal\assets\logo.png >nul

echo.
echo [build-walk.bat] Embedding icon and manifest...
go install github.com/tc-hib/go-winres@latest
if errorlevel 1 (
  echo [build-walk.bat] go-winres install failed.
  pause
  exit /b 1
)

go-winres simply --icon logo.png --manifest gui ^
  --file-description "RelayPane SFTP Client (Win32)" ^
  --product-name "RelayPane" ^
  --original-filename "RelayPane-walk.exe" ^
  --out cmd\relaypane-walk\rsrc --arch amd64
if errorlevel 1 (
  echo [build-walk.bat] go-winres failed.
  pause
  exit /b 1
)

echo [build-walk.bat] Building RelayPane-walk.exe ...
go build -ldflags="-H=windowsgui -s -w" -o RelayPane-walk.exe ./cmd/relaypane-walk
if errorlevel 1 (
  echo [build-walk.bat] Build FAILED.
  pause
  exit /b 1
)

echo [build-walk.bat] Built: %CD%\RelayPane-walk.exe
exit /b 0
