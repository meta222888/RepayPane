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

REM --- Fast path: already compiled ---
if /i not "%~1"=="rebuild" (
  if exist "dev\RelayPane-dev.exe" goto :launch_dev
  if exist "RelayPane.exe" (
    echo [run.bat] Starting RelayPane.exe
    RelayPane.exe
    exit /b !ERRORLEVEL!
  )
)

REM --- First-time compile (CGO/Fyne: slow, often silent after internal/ui) ---
echo.
echo [run.bat] First compile: compiling to dev\RelayPane-dev.exe
echo [run.bat] After "internal/ui" GCC may run 10-20 min with NO new lines. Normal.
echo [run.bat] Next time run.bat will start instantly ^(skip compile^).
echo [run.bat] Force rebuild later: run.bat rebuild
echo.

go build -x -o dev\RelayPane-dev.exe ./cmd/relaypane
if errorlevel 1 (
  echo.
  echo [run.bat] Build failed.
  pause
  exit /b 1
)

:launch_dev
echo.
echo [run.bat] Launching dev\RelayPane-dev.exe ...
dev\RelayPane-dev.exe
if errorlevel 1 (
  echo [run.bat] RelayPane exited with error.
  pause
)
exit /b %ERRORLEVEL%
