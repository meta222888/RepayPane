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
  echo [run-win.bat] GCC not found. walk requires CGO/MinGW on Windows.
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  pause
  exit /b 1
)

:gcc_ok
if not exist "dev" mkdir dev

if exist logo.png copy /Y logo.png internal\assets\logo.png >nul

echo.
echo [run-win.bat] Building dev\RelayPane-walk-dev.exe ^(Win32 walk UI, console on for errors^)...
if not exist "dev\RelayPane-walk-dev.exe" (
  echo [run-win.bat] First build may take a minute while CGO compiles walk.
) else (
  echo [run-win.bat] Rebuild usually finishes in seconds when only Go code changed.
)
echo.

go build -v -o dev\RelayPane-walk-dev.exe ./cmd/relaypane-walk
if errorlevel 1 (
  echo.
  echo [run-win.bat] Build failed.
  pause
  exit /b 1
)

echo.
echo [run-win.bat] Launching...
dev\RelayPane-walk-dev.exe
if errorlevel 1 (
  echo [run-win.bat] RelayPane-walk exited with error.
  pause
)
exit /b %ERRORLEVEL%
