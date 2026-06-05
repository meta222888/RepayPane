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
  pause
  exit /b 1
)

:gcc_ok
echo.
echo [build-walk.bat] Embedding manifest...
go run github.com/akavel/rsrc -manifest cmd/relaypane-walk/app.manifest -o cmd/relaypane-walk/rsrc.syso
if errorlevel 1 (
  echo [build-walk.bat] rsrc failed; trying go install...
  go install github.com/akavel/rsrc@latest
  rsrc -manifest cmd/relaypane-walk/app.manifest -o cmd/relaypane-walk/rsrc.syso
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
