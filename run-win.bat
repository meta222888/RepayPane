@echo off
setlocal EnableDelayedExpansion
cd /d "%~dp0"

set CGO_ENABLED=1

where go >nul 2>&1
if errorlevel 1 (
  echo [run-win.bat] Go not found in PATH.
  pause
  exit /b 1
)

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
  echo [run-win.bat] GCC not found. walk requires CGO/MinGW on Windows.
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  pause
  exit /b 1
)

:gcc_ok
if not exist "dev" mkdir dev

if exist logo.png copy /Y logo.png internal\assets\logo.png >nul

set "BUILD_ID=dev"
set "GIT_BRANCH="
for /f "delims=" %%B in ('git branch --show-current 2^>nul') do set "GIT_BRANCH=%%B"
for /f "delims=" %%H in ('git rev-parse --short HEAD 2^>nul') do set "BUILD_ID=%%H"

echo.
echo ============================================================
echo   RelayPane Win32 walk UI  ^(NOT Fyne — use run.bat for Fyne^)
echo   Branch: %GIT_BRANCH%   Build: %BUILD_ID%
echo   Output: dev\RelayPane-walk-dev.exe
echo ============================================================
echo.
echo [run-win.bat] Tip: if UI still looks old, run: git pull
echo.

taskkill /IM RelayPane-walk-dev.exe /F >nul 2>&1

if exist "dev\RelayPane-walk-dev.exe" del /F /Q "dev\RelayPane-walk-dev.exe"

echo [run-win.bat] Building dev\RelayPane-walk-dev.exe ...
echo.

go build -v -ldflags="-X main.buildID=%BUILD_ID%" -o dev\RelayPane-walk-dev.exe ./cmd/relaypane-walk
if errorlevel 1 (
  echo.
  echo [run-win.bat] Build failed.
  pause
  exit /b 1
)

if not exist "dev\RelayPane-walk-dev.exe" (
  echo [run-win.bat] Build reported success but exe was not created.
  pause
  exit /b 1
)

echo.
echo [run-win.bat] Launching dev\RelayPane-walk-dev.exe ^(build %BUILD_ID%^) ...
echo.
dev\RelayPane-walk-dev.exe
if errorlevel 1 (
  echo [run-win.bat] RelayPane-walk exited with error.
  pause
)
exit /b %ERRORLEVEL%
