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
  echo [run.bat] Using GCC: !MINGW!
) else (
  echo [run.bat] GCC not found. Install MinGW, e.g.:
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  pause
  exit /b 1
)

:gcc_ok
if not exist "dev" mkdir dev

echo.
echo [run.bat] Compiling RelayPane...
echo [run.bat] First compile may take 5-15 minutes ^(Fyne + CGO^). Please wait.
echo [run.bat] Later runs are much faster ^(incremental build^).
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
  echo.
  echo [run.bat] RelayPane exited with an error.
  pause
)
exit /b %ERRORLEVEL%
