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
  echo [run.bat] GCC not found. Install MinGW:
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  pause
  exit /b 1
)

:gcc_ok
if not exist "dev" mkdir dev

echo.
echo [run.bat] Fyne UI — for Win32 walk UI ^(Lovable design^) use run-win.bat instead.
echo [run.bat] Building dev\RelayPane-dev.exe ^(incremental — code changes always apply^)...
if not exist "dev\RelayPane-dev.exe" (
  echo [run.bat] First build: after "internal/ui" GCC may run 10-20 min with no output. Please wait.
) else (
  echo [run.bat] Rebuild usually finishes in seconds when only Go code changed.
)
echo.

go build -v -o dev\RelayPane-dev.exe ./cmd/relaypane
if errorlevel 1 (
  echo.
  echo [run.bat] Build failed.
  pause
  exit /b 1
)

echo.
echo [run.bat] Launching...
dev\RelayPane-dev.exe
if errorlevel 1 (
  echo [run.bat] RelayPane exited with error.
  pause
)
exit /b %ERRORLEVEL%
